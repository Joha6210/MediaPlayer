package sources

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"

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

func (a *bluetoothAdapter) ListenMetadata() (<-chan source.Metadata, error) {
	metadata := make(chan source.Metadata, 4)
	if a.testMode {
		return metadata, nil
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("connecting to system D-Bus: %w", err)
	}
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		conn.Close()
		return nil, fmt.Errorf("subscribing to BlueZ metadata: %w", err)
	}

	signals := make(chan *dbus.Signal, 16)
	conn.Signal(signals)
	go func() {
		defer close(metadata)
		defer conn.RemoveSignal(signals)
		defer conn.Close()
		for signal := range signals {
			if signal == nil || !strings.Contains(string(signal.Path), "/player") || len(signal.Body) < 2 {
				continue
			}
			changed, ok := signal.Body[1].(map[string]dbus.Variant)
			if !ok {
				continue
			}
			item := source.Metadata{}
			if value, ok := changed["Title"]; ok {
				item.Title, _ = value.Value().(string)
			}
			if value, ok := changed["Artist"]; ok {
				item.Artist, _ = value.Value().(string)
			}
			if value, ok := changed["Album"]; ok {
				item.Album, _ = value.Value().(string)
			}
			if item.Title != "" || item.Artist != "" || item.Album != "" {
				metadata <- item
			}
		}
	}()

	return metadata, nil
}
