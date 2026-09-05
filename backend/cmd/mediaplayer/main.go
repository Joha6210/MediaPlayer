package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mediaplayer/backend/internal/api"
	"mediaplayer/backend/internal/config"
	"mediaplayer/backend/internal/player"
	"mediaplayer/backend/internal/source"
	"mediaplayer/backend/internal/sources"
	projectroot "mediaplayer/root"
)

func main() {
	cfgPath := os.Getenv("MEDIAPLAYER_CONFIG")
	if cfgPath == "" {
		cfgPath = "config/default.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	var playbackClient source.Player
	if cfg.Runtime.TestMode {
		log.Printf("starting in test mode with mock player")
		playbackClient = player.NewMockPlayer(cfg.SourceDefaults.DefaultVolume)
	} else {
		log.Printf("Starting backend server...")
		playbackClient = player.NewMPVClient(cfg.MPV.SocketPath)
	}

	manager := source.NewManager(playbackClient, cfg.SourceDefaults.DefaultVolume)
	manager.Register(sources.InternetRadioSource, sources.NewInternetAdapter())
	manager.Register(sources.DRSource, sources.NewDRAdapter(cfg.Runtime.TestMode))
	manager.Register(sources.PlexampSource, sources.NewPlexampAdapter(cfg.Plexamp, cfg.Runtime.TestMode))
	manager.Register(sources.BluetoothSource, sources.NewBluetoothAdapter(cfg.Runtime.TestMode))

	// Hent det indlejrede SvelteKit-filsystem fra din rod-pakke
	publicFS := projectroot.GetFrontendFS()

	// Send det videre til api-serveren
	server := api.NewServer(cfg.HTTP.ListenAddr, manager, publicFS)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		errCh <- server.Start()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	case err := <-errCh:
		if err != nil {
			log.Fatalf("server failed: %v", err)
		}
	}
}
