package sources

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"mediaplayer/backend/internal/source"
)

type bluetoothAdapter struct {
	commandRunner commandRunner
	controller    *bluezController
	testMode      bool
}

type commandRunner interface {
	Run(ctx context.Context, name string, args ...string) error
}

type execRunner struct{}

func (r *execRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s %v failed: %w (%s)", name, args, err, string(output))
	}
	return nil
}

func NewBluetoothAdapter(testMode bool) source.Adapter {
	adapter := &bluetoothAdapter{
		commandRunner: &execRunner{},
		testMode:      testMode,
	}
	if !testMode {
		adapter.controller = newBluezController()
	}
	return adapter
}

func (a *bluetoothAdapter) Resolve(ctx context.Context, _ source.SelectRequest) (source.PlayRequest, error) {
	if a.testMode {
		return source.PlayRequest{
			Title:     "Bluetooth Sink (Mock)",
			UsePlayer: false,
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := a.commandRunner.Run(ctx, "bluetoothctl", "discoverable", "on"); err != nil {
		return source.PlayRequest{}, err
	}
	if err := a.commandRunner.Run(ctx, "bluetoothctl", "pairable", "on"); err != nil {
		return source.PlayRequest{}, err
	}

	return source.PlayRequest{
		Title:     "Bluetooth Sink",
		UsePlayer: false,
	}, nil
}

func (a *bluetoothAdapter) IsRadioAdapter() bool {
	return false
}

func (a *bluetoothAdapter) GetStations() map[string]source.Station {
	return map[string]source.Station{}
}

func (a *bluetoothAdapter) PlayPause(paused bool) error {
	if a.controller == nil {
		return errors.New("bluetooth controller unavailable")
	}
	return a.controller.PlayPause(paused)
}

func (a *bluetoothAdapter) NextTrack() error {
	if a.controller == nil {
		return errors.New("bluetooth controller unavailable")
	}
	return a.controller.NextTrack()
}

func (a *bluetoothAdapter) PrevTrack() error {
	if a.controller == nil {
		return errors.New("bluetooth controller unavailable")
	}
	return a.controller.PrevTrack()
}

func (a *bluetoothAdapter) SetVolume(volume int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return a.commandRunner.Run(ctx, "wpctl", "set-volume", "@DEFAULT_AUDIO_SINK@", fmt.Sprintf("%d%%", volume))
}

func (a *bluetoothAdapter) ListenEvents() (<-chan source.PlaybackEvent, error) {
	if a.controller == nil {
		ch := make(chan source.PlaybackEvent)
		close(ch)
		return ch, nil
	}
	return a.controller.ListenEvents()
}

func (a *bluetoothAdapter) SetFavoriteStation(stationUUID string) error {
	return nil
}

func (a *bluetoothAdapter) DefaultStation() source.Station {
	return source.Station{}
}

var _ source.Controller = (*bluetoothAdapter)(nil)
var _ source.VolumeController = (*bluetoothAdapter)(nil)
