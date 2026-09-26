package source

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
)

type Manager struct {
	mu               sync.RWMutex
	player           Player
	adapters         map[string]Adapter
	controllers      map[string]Controller
	volumes          map[string]VolumeController
	state            SourceState
	controllerStates map[string]PlaybackEvent
	subs             map[chan SourceState]struct{}
	currAdapter      Adapter
	ActiveStation    Station
}

func NewManager(player Player, defaultVolume int) *Manager {
	m := &Manager{
		player:           player,
		adapters:         make(map[string]Adapter),
		controllers:      make(map[string]Controller),
		volumes:          make(map[string]VolumeController),
		controllerStates: make(map[string]PlaybackEvent),
		state: SourceState{
			Volume:  clampVolume(defaultVolume),
			Playing: false,
		},
		subs: make(map[chan SourceState]struct{}),
	}
	m.startEventListener()
	return m
}

func (m *Manager) startEventListener() {
	ch, err := m.player.ListenEvents()
	if err != nil {
		log.Printf("Kunne ikke starte mpv eventlytter: %v", err)
		return
	}
	if ch == nil {
		return
	}

	go func() {
		for info := range ch {
			m.mu.Lock()
			if _, controlled := m.controllers[m.state.ActiveSource]; controlled {
				m.mu.Unlock()
				continue
			}
			// Opdater m.state direkte med live-data fra MPV streamen
			m.state.StreamTitle = info.Title
			m.state.Artist = ""
			m.state.Album = ""
			m.state.Paused = info.Paused
			m.state.Playing = !info.Paused
			m.state.Position = 0
			m.state.Duration = 0

			// Genbrug din eksisterende notify-mekanisme til at skubbe data til frontenden
			m.notifyLocked()
			m.mu.Unlock()
		}
	}()
}

func (m *Manager) Register(name string, adapter Adapter) {
	m.mu.Lock()
	m.adapters[name] = adapter
	controller, hasController := adapter.(Controller)
	if hasController {
		m.controllers[name] = controller
	}
	volumeController, hasVolumeController := adapter.(VolumeController)
	if hasVolumeController {
		m.volumes[name] = volumeController
	}
	m.mu.Unlock()

	if hasController {
		ch, err := controller.ListenEvents()
		if err != nil {
			log.Printf("Could not start %s playback listener: %v", name, err)
		} else {
			go m.listenController(name, ch)
		}
	}

	if listener, ok := adapter.(MetadataListener); ok {
		ch, err := listener.ListenMetadata()
		if err != nil {
			log.Printf("Could not start %s metadata listener: %v", name, err)
			return
		}
		go func() {
			for metadata := range ch {
				m.mu.Lock()
				if m.state.ActiveSource == name {
					m.state.StreamTitle = metadata.Title
					if metadata.Artist != "" {
						m.state.StreamTitle = metadata.Artist + " - " + metadata.Title
					}
					m.notifyLocked()
				}
				m.mu.Unlock()
			}
		}()
	}
}

func (m *Manager) listenController(name string, ch <-chan PlaybackEvent) {
	for event := range ch {
		m.mu.Lock()
		m.controllerStates[name] = event
		if m.state.ActiveSource == name {
			m.state.StreamTitle = event.Title
			m.state.Artist = event.Artist
			m.state.Album = event.Album
			m.state.Playing = event.Playing
			m.state.Paused = event.Paused
			m.state.Position = event.Position
			m.state.Duration = event.Duration
			m.notifyLocked()
		}
		m.mu.Unlock()
	}
}

func (m *Manager) State() SourceState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *Manager) GetAdapters() map[string]Adapter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.adapters
}

func (m *Manager) SetVolume(volume int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	normalized := clampVolume(volume)
	if controller, ok := m.volumes[m.state.ActiveSource]; ok {
		if err := controller.SetVolume(normalized); err != nil {
			return err
		}
	} else {
		if err := m.player.SetVolume(normalized); err != nil {
			return err
		}
	}
	m.state.Volume = normalized
	m.notifyLocked()
	return nil
}

func (m *Manager) PlayPause(state bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if controller, ok := m.controllers[m.state.ActiveSource]; ok {
		if err := controller.PlayPause(state); err != nil {
			return err
		}
	} else if err := m.player.PlayPause(state); err != nil {
		return err
	}
	m.state.Paused = state
	m.notifyLocked()
	return nil
}

func (m *Manager) NextTrack() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if controller, ok := m.controllers[m.state.ActiveSource]; ok {
		if err := controller.NextTrack(); err != nil {
			return err
		}
	} else if err := m.player.NextTrack(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) PrevTrack() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if controller, ok := m.controllers[m.state.ActiveSource]; ok {
		if err := controller.PrevTrack(); err != nil {
			return err
		}
	} else if err := m.player.PrevTrack(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) CatchUp() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.player.CatchUp(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Select(ctx context.Context, req SelectRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.player.Stop()

	log.Printf("Selecting source: %s", req.Source)

	adapter, ok := m.adapters[req.Source]
	if !ok {
		return fmt.Errorf("unknown source %q", req.Source)
	}

	m.currAdapter = adapter

	playReq, err := adapter.Resolve(ctx, req)
	if err != nil {
		return err
	}

	if playReq.UsePlayer {
		if playReq.URL == "" {
			return errors.New("playback source resolved without URL")
		}
		if err := m.player.Play(playReq.URL, m.state.Volume, playReq.Headers); err != nil {
			return err
		}
		log.Printf("Playing source: %s", req.Source)
		m.state.Playing = true
		m.state.ActiveStation = req.Station

	} else {
		if err := m.player.Stop(); err != nil {
			return err
		}
		m.state.Playing = false
	}

	m.state.StreamTitle = ""
	m.state.Artist = ""
	m.state.Album = ""
	m.state.Paused = false
	m.state.Position = 0
	m.state.Duration = 0
	m.state.ActiveSource = req.Source
	if event, ok := m.controllerStates[req.Source]; ok {
		m.state.StreamTitle = event.Title
		m.state.Artist = event.Artist
		m.state.Album = event.Album
		m.state.Playing = event.Playing
		m.state.Paused = event.Paused
		m.state.Position = event.Position
		m.state.Duration = event.Duration
	}
	if playReq.Title != "" {
		m.state.Label = playReq.Title
	} else if req.Title != "" {
		m.state.Label = req.Title
	} else {
		m.state.Label = req.Source
	}
	m.notifyLocked()
	return nil
}

func (m *Manager) Subscribe() (chan SourceState, func()) {
	ch := make(chan SourceState, 4)

	m.mu.Lock()
	m.subs[ch] = struct{}{}
	state := m.state
	m.mu.Unlock()

	ch <- state

	unsubscribe := func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if _, ok := m.subs[ch]; ok {
			delete(m.subs, ch)
			close(ch)
		}
	}
	return ch, unsubscribe
}

func (m *Manager) GetCurrentAdapter() Adapter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currAdapter
}

func (m *Manager) notifyLocked() {
	for sub := range m.subs {
		select {
		case sub <- m.state:
		default:
		}
	}
}

func clampVolume(volume int) int {
	if volume < 0 {
		return 0
	}
	if volume > 100 {
		return 100
	}
	return volume
}
