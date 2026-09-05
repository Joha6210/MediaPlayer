package player

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/Microsoft/go-winio"
)

// PlayInfo indeholder de data, dit frontend skal bruge
type PlayInfo struct {
	Title  string
	Paused bool
}

// MPVEvent matcher strukturen på de JSON-beskeder, mpv sender asynkront
type MPVEvent struct {
	Event string `json:"event"`
	Name  string `json:"name"`
	Data  any    `json:"data"`
}

type MPVClient struct {
	socketPath string
}

func NewMPVClient(socketPath string) *MPVClient {
	// Hvis socketPath er tom, sæt en standard
	if socketPath == "" {
		if runtime.GOOS == "windows" {
			socketPath = "//./pipe/mpvsocket"
		} else {
			socketPath = "/tmp/mpvsocket"
		}
	}

	cmd := exec.Command("mpv", "--input-ipc-server="+socketPath, "--idle")
	//cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("Fejl ved start af mpv: %v\n", err)
	}

	return &MPVClient{socketPath: socketPath}
}

func dialSocket(path string) (net.Conn, error) {
	timeout := 2 * time.Second
	if runtime.GOOS == "windows" {
		return winio.DialPipe(path, &timeout)
	}
	return net.DialTimeout("unix", path, timeout)
}

func (m *MPVClient) ListenEvents() (<-chan struct {
	Title  string
	Paused bool
}, error) {
	outChan := make(chan struct {
		Title  string
		Paused bool
	}, 10)

	conn, err := dialSocket(m.socketPath)
	if err != nil {
		return nil, fmt.Errorf("kunne ikke forbinde event-lytter: %w", err)
	}

	// Registrer hvilke properties vi vil overvåge i mpv
	// 1 = media-title (sang/radiokanal info), 2 = pause status
	observeTitle := `{"command": ["observe_property", 1, "media-title"]}` + "\n"
	observePause := `{"command": ["observe_property", 2, "pause"]}` + "\n"
	observeIcy := `{"command": ["observe_property", 3, "metadata"]}` + "\n"

	if _, err := conn.Write([]byte(observeTitle + observePause + observeIcy)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("kunne ikke registrere property observers: %w", err)
	}

	fmt.Println("Subscribing to mpv info...")

	// Start asynkron lytning i baggrunden
	go func() {
		defer conn.Close()
		defer close(outChan)

		scanner := bufio.NewScanner(conn)
		var currentStatus PlayInfo

		for scanner.Scan() {
			var ev MPVEvent
			if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
				continue
			}

			// Tjek om det er en opdatering på en af vores observerede værdier
			if ev.Event == "property-change" {
				updated := false

				switch ev.Name {
				case "media-title":
					if ev.Data != nil {
						if str, ok := ev.Data.(string); ok {
							currentStatus.Title = str
							updated = true
						}
					}
				case "pause":
					if ev.Data != nil {
						if paused, ok := ev.Data.(bool); ok {
							currentStatus.Paused = paused
							updated = true
						}
					}
				case "metadata":
					// mpv often sends stream metadata (like Icecast/ICY song titles) inside a map under metadata
					if mp, ok := ev.Data.(map[string]any); ok {
						if title, exists := mp["icy-title"]; exists {
							if str, ok := title.(string); ok && str != "" {
								currentStatus.Title = str
								updated = true
							}
						} else if title, exists := mp["TITLE"]; exists {
							if str, ok := title.(string); ok && str != "" {
								currentStatus.Title = str
								updated = true
							}
						}
					}
				}

				// Hvis status ændrede sig, sender vi den nye PlayInfo ud på kanalen
				if updated {
					outChan <- currentStatus
				}
			}
		}
	}()

	return outChan, nil
}

func (m *MPVClient) Play(url string, volume int, headers map[string]string) error {
	if len(headers) > 0 {
		if err := m.command("set_property", "http-header-fields", flattenHeaders(headers)); err != nil {
			return err
		}
	}
	if err := m.command("loadfile", url, "replace"); err != nil {
		return err
	}
	return m.command("set_property", "volume", volume)
}

func (m *MPVClient) Stop() error {
	return m.command("stop")
}

func (m *MPVClient) Pause(paused bool) error {
	return m.command("set_property", "pause", paused)
}

func (m *MPVClient) SetVolume(volume int) error {
	return m.command("set_property", "volume", volume)
}

func (m *MPVClient) command(args ...any) error {

	conn, err := dialSocket(m.socketPath)

	if err != nil {
		return fmt.Errorf("connecting mpv socket: %w", err)
	}
	defer conn.Close()

	payload := map[string]any{"command": args}
	enc := json.NewEncoder(conn)
	if err := enc.Encode(payload); err != nil {
		return fmt.Errorf("sending mpv command: %w", err)
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("reading mpv response: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buffer[:n]), &result); err != nil {
		return fmt.Errorf("parsing mpv response: %w", err)
	}

	if status, ok := result["error"].(string); !ok || status != "success" {
		return fmt.Errorf("mpv command failed: %v", result["error"])
	}

	return nil
}

func flattenHeaders(headers map[string]string) []string {
	out := make([]string, 0, len(headers))
	for k, v := range headers {
		out = append(out, fmt.Sprintf("%s: %s", k, v))
	}
	return out
}
