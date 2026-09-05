package source

import (
	"context"
	"errors"
	"testing"
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

func (p *fakePlayer) Pause(_ bool) error {
	return nil
}

func (p *fakePlayer) SetVolume(v int) error {
	p.lastVolume = v
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

func TestSelectAdapterError(t *testing.T) {
	player := &fakePlayer{}
	m := NewManager(player, 50)
	m.Register("internet-radio", &fakeAdapter{err: errors.New("boom")})

	err := m.Select(context.Background(), SelectRequest{Source: "internet-radio"})
	if err == nil {
		t.Fatal("expected error")
	}
}
