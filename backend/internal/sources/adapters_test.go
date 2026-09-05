package sources

import (
	"context"
	"testing"

	"mediaplayer/backend/internal/config"
	"mediaplayer/backend/internal/source"
)

func TestDRAdapterResolveInTestMode(t *testing.T) {
	adapter := NewDRAdapter(true)
	req := source.SelectRequest{
		Source: "dr-radio",
		Meta:   map[string]string{"station": "p3"},
	}

	playReq, err := adapter.Resolve(context.Background(), req)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if playReq.URL == "" {
		t.Fatal("expected URL to be resolved")
	}
}

func TestPlexampAdapterResolveInTestMode(t *testing.T) {
	adapter := NewPlexampAdapter(config.PlexampConfig{}, true)
	req := source.SelectRequest{
		Source: "plexamp",
		Meta:   map[string]string{"path": "/library/mock.mp3"},
	}

	playReq, err := adapter.Resolve(context.Background(), req)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if playReq.URL != "http://127.0.0.1:65535/library/mock.mp3" {
		t.Fatalf("unexpected URL %q", playReq.URL)
	}
}

func TestBluetoothAdapterResolveInTestMode(t *testing.T) {
	adapter := NewBluetoothAdapter(true)
	playReq, err := adapter.Resolve(context.Background(), source.SelectRequest{
		Source: "bluetooth",
	})
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if playReq.UsePlayer {
		t.Fatal("expected bluetooth mock to bypass player")
	}
}
