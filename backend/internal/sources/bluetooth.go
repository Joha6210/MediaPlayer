package sources

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"mediaplayer/backend/internal/source"
)

type bluetoothAdapter struct {
	commandRunner commandRunner
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
	return &bluetoothAdapter{
		commandRunner: &execRunner{},
		testMode:      testMode,
	}
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

func (a *bluetoothAdapter) GetStations() []source.Station {
	return []source.Station{}
}
