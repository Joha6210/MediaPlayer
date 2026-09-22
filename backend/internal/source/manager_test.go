package source

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakePlayer struct {
	lastURL    string
	lastVolume int
	stopCalled bool
}

func (p *fakePlayer) Play(url string, volume int, _ map[string]string) error {
	p.lastURL = url
	p.lastVolume = volume
	return nil
}

func (p *fakePlayer) Stop() error {
	p.stopCalled = true
	return nil
}

func (p *fakePlayer) PlayPause(_ bool) error {
	return nil
}

func (p *fakePlayer) SetVolume(v int) error {
	p.lastVolume = v
	return nil
}

func (p *fakePlayer) NextTrack() error {
	return nil
}

func (p *fakePlayer) PrevTrack() error {
	return nil
}

func (p *fakePlayer) CatchUp() error {
	return nil
}

func (p *fakePlayer) Close() error {
	return nil
}

func (p *fakePlayer) ListenEvents() (<-chan struct {
	Title  string
	Paused bool
}, error) {
	return nil, errors.New("not implemented")
}

type fakeAdapter struct {
	playReq PlayRequest
	err     error
}

type fakeControllerAdapter struct {
	fakeAdapter
	events     chan PlaybackEvent
	paused     bool
	nextCalled bool
	prevCalled bool
	volume     int
}

func (a *fakeControllerAdapter) PlayPause(paused bool) error {
	a.paused = paused
	return nil
}

func (a *fakeControllerAdapter) NextTrack() error {
	a.nextCalled = true
	return nil
}

func (a *fakeControllerAdapter) PrevTrack() error {
	a.prevCalled = true
	return nil
}

func (a *fakeControllerAdapter) SetVolume(volume int) error {
	a.volume = volume
	return nil
}

func (a *fakeControllerAdapter) ListenEvents() (<-chan PlaybackEvent, error) {
	return a.events, nil
}

func (a *fakeAdapter) Resolve(_ context.Context, _ SelectRequest) (PlayRequest, error) {
	if a.err != nil {
		return PlayRequest{}, a.err
	}
	return a.playReq, nil
}

func (a *fakeAdapter) GetStations() []Station {
	return []Station{}
}

func TestSelectPlaybackSource(t *testing.T) {
	player := &fakePlayer{}
	m := NewManager(player, 50)
	m.Register("internet-radio", &fakeAdapter{
		playReq: PlayRequest{URL: "https://example.com/stream.mp3", UsePlayer: true, Title: "Test"},
	})

	if err := m.Select(context.Background(), SelectRequest{Source: "internet-radio"}); err != nil {
		t.Fatalf("select returned error: %v", err)
	}

	state := m.State()
	if state.ActiveSource != "internet-radio" {
		t.Fatalf("expected active source internet-radio, got %s", state.ActiveSource)
	}
	if !state.Playing {
		t.Fatalf("expected playing true")
	}
	if player.lastURL == "" {
		t.Fatalf("expected url to be sent to player")
	}
}

func TestSelectBluetoothStopsPlayer(t *testing.T) {
	player := &fakePlayer{}
	m := NewManager(player, 50)
	m.Register("bluetooth", &fakeAdapter{
		playReq: PlayRequest{UsePlayer: false, Title: "Bluetooth Sink"},
	})

	if err := m.Select(context.Background(), SelectRequest{Source: "bluetooth"}); err != nil {
		t.Fatalf("select returned error: %v", err)
	}

	if !player.stopCalled {
		t.Fatalf("expected stop to be called")
	}
	if m.State().Playing {
		t.Fatalf("expected playing false")
	}
}

func TestSelectUnknownSource(t *testing.T) {
	player := &fakePlayer{}
	m := NewManager(player, 50)
	err := m.Select(context.Background(), SelectRequest{Source: "missing"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBluetoothControllerReceivesControlsAndEvents(t *testing.T) {
	player := &fakePlayer{}
	controller := &fakeControllerAdapter{
		fakeAdapter: fakeAdapter{playReq: PlayRequest{UsePlayer: false, Title: "Bluetooth Sink"}},
		events:      make(chan PlaybackEvent, 1),
	}
	m := NewManager(player, 50)
	m.Register("bluetooth", controller)

	if err := m.Select(context.Background(), SelectRequest{Source: "bluetooth"}); err != nil {
		t.Fatalf("select returned error: %v", err)
	}
	if err := m.PlayPause(true); err != nil {
		t.Fatalf("play/pause returned error: %v", err)
	}
	if err := m.NextTrack(); err != nil {
		t.Fatalf("next returned error: %v", err)
	}
	if err := m.PrevTrack(); err != nil {
		t.Fatalf("previous returned error: %v", err)
	}
	if err := m.SetVolume(72); err != nil {
		t.Fatalf("volume returned error: %v", err)
	}

	controller.events <- PlaybackEvent{
		Title: "Song", Artist: "Artist", Album: "Album",
		Playing: true, Position: 12_000_000, Duration: 245_000_000,
	}

	deadline := time.After(time.Second)
	for {
		state := m.State()
		if state.StreamTitle == "Song" {
			if state.Artist != "Artist" || state.Album != "Album" || state.Position != 12_000_000 {
				t.Fatalf("unexpected controller state: %+v", state)
			}
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for controller event")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	if !controller.paused || !controller.nextCalled || !controller.prevCalled || controller.volume != 72 {
		t.Fatalf("controller did not receive all controls: %+v", controller)
	}
	if player.lastVolume != 0 {
		t.Fatalf("expected mpv volume to remain untouched, got %d", player.lastVolume)
	}
}

func TestSelectAdapterError(t *testing.T) {
	player := &fakePlayer{}
	m := NewManager(player, 50)
	m.Register("internet-radio", &fakeAdapter{err: errors.New("boom")})

	err := m.Select(context.Background(), SelectRequest{Source: "internet-radio"})
	if err == nil {
		t.Fatal("expected error")
	}
}
