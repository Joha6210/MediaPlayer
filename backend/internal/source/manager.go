package source

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
)

type Manager struct {
	mu            sync.RWMutex
	player        Player
	adapters      map[string]Adapter
	state         SourceState
	subs          map[chan SourceState]struct{}
	currAdapter   Adapter
	ActiveStation Station
}

func NewManager(player Player, defaultVolume int) *Manager {
	m := &Manager{
		player:   player,
		adapters: make(map[string]Adapter),
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
			// Opdater m.state direkte med live-data fra MPV streamen
			m.state.StreamTitle = info.Title
			m.state.Paused = info.Paused

			// Genbrug din eksisterende notify-mekanisme til at skubbe data til frontenden
			m.notifyLocked()
			m.mu.Unlock()
		}
	}()
}

func (m *Manager) Register(name string, adapter Adapter) {
	m.mu.Lock()
	m.adapters[name] = adapter
	m.mu.Unlock()

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
	if m.state.ActiveSource != "bluetooth" {
		if err := m.player.SetVolume(normalized); err != nil {
			return err
		}
	}
	m.state.Volume = normalized
	m.notifyLocked()
	return nil
}

func (m *Manager) Select(ctx context.Context, req SelectRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

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

	m.state.ActiveSource = req.Source
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
