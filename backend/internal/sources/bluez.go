package sources

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"

	"mediaplayer/backend/internal/source"
)

const (
	bluezService       = "org.bluez"
	bluezObjectManager = "org.freedesktop.DBus.ObjectManager"
	bluezPlayer        = "org.bluez.MediaPlayer1"
	propertiesIface    = "org.freedesktop.DBus.Properties"
)

type bluezController struct {
	conn   *dbus.Conn
	signal chan *dbus.Signal
	events chan source.PlaybackEvent

	mu   sync.RWMutex
	path dbus.ObjectPath
	last source.PlaybackEvent
}

func newBluezController() *bluezController {
	return &bluezController{}
}

func (c *bluezController) ensureConnected() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return nil
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("connecting to system D-Bus: %w", err)
	}
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface(propertiesIface),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		conn.Close()
		return fmt.Errorf("subscribing to BlueZ properties: %w", err)
	}
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface(bluezObjectManager),
		dbus.WithMatchMember("InterfacesAdded"),
	); err != nil {
		conn.Close()
		return fmt.Errorf("subscribing to BlueZ additions: %w", err)
	}
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface(bluezObjectManager),
		dbus.WithMatchMember("InterfacesRemoved"),
	); err != nil {
		conn.Close()
		return fmt.Errorf("subscribing to BlueZ removals: %w", err)
	}

	c.conn = conn
	c.signal = make(chan *dbus.Signal, 32)
	conn.Signal(c.signal)
	go c.listenSignals()
	return c.refreshLocked()
}

func (c *bluezController) refreshLocked() error {
	var objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	call := c.conn.Object(bluezService, dbus.ObjectPath("/")).Call(bluezObjectManager+".GetManagedObjects", 0)
	if err := call.Store(&objects); err != nil {
		return fmt.Errorf("discovering BlueZ media players: %w", err)
	}

	for path, interfaces := range objects {
		properties, ok := interfaces[bluezPlayer]
		if !ok {
			continue
		}
		c.path = path
		c.last = eventFromProperties(properties, c.last)
		return nil
	}
	c.path = ""
	c.last = source.PlaybackEvent{}
	return nil
}

func (c *bluezController) listenSignals() {
	for signal := range c.signal {
		if signal == nil || len(signal.Body) < 2 {
			continue
		}
		if signal.Name == bluezObjectManager+".InterfacesAdded" {
			c.handleInterfaceAdded(signal)
			continue
		}
		if signal.Name == bluezObjectManager+".InterfacesRemoved" {
			c.handleInterfaceRemoved(signal)
			continue
		}
		if signal.Name != propertiesIface+".PropertiesChanged" || signal.Path == "/" {
			continue
		}
		iface, ok := signal.Body[0].(string)
		if !ok || iface != bluezPlayer {
			continue
		}
		changed, ok := signal.Body[1].(map[string]dbus.Variant)
		if !ok {
			continue
		}

		c.mu.Lock()
		if c.path == "" || c.path == signal.Path {
			c.path = signal.Path
			c.last = eventFromProperties(changed, c.last)
			event := c.last
			c.mu.Unlock()
			c.emit(event)
			continue
		}
		c.mu.Unlock()
	}
}

func (c *bluezController) handleInterfaceAdded(signal *dbus.Signal) {
	path, ok := signal.Body[0].(dbus.ObjectPath)
	if !ok {
		return
	}
	interfaces, ok := signal.Body[1].(map[string]map[string]dbus.Variant)
	if !ok {
		return
	}
	properties, ok := interfaces[bluezPlayer]
	if !ok {
		return
	}

	c.mu.Lock()
	c.path = path
	c.last = eventFromProperties(properties, source.PlaybackEvent{})
	event := c.last
	c.mu.Unlock()
	c.emit(event)
}

func (c *bluezController) handleInterfaceRemoved(signal *dbus.Signal) {
	path, ok := signal.Body[0].(dbus.ObjectPath)
	if !ok {
		return
	}
	interfaces, ok := signal.Body[1].([]string)
	if !ok {
		return
	}
	for _, iface := range interfaces {
		if iface != bluezPlayer {
			continue
		}
		c.mu.Lock()
		if c.path == path {
			c.path = ""
			c.last = source.PlaybackEvent{}
		}
		c.mu.Unlock()
		c.emit(source.PlaybackEvent{})
	}
}

func (c *bluezController) emit(event source.PlaybackEvent) {
	if c.events == nil {
		return
	}
	select {
	case c.events <- event:
	default:
	}
}

func (c *bluezController) call(method string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	c.mu.RLock()
	path := c.path
	conn := c.conn
	c.mu.RUnlock()
	if path == "" {
		return errors.New("no Bluetooth media player connected")
	}
	if err := conn.Object(bluezService, path).Call(bluezPlayer+"."+method, 0).Err; err != nil {
		return fmt.Errorf("BlueZ %s: %w", method, err)
	}
	return nil
}

func (c *bluezController) PlayPause(paused bool) error {
	if paused {
		return c.call("Pause")
	}
	return c.call("Play")
}

func (c *bluezController) NextTrack() error { return c.call("Next") }
func (c *bluezController) PrevTrack() error { return c.call("Previous") }
func (c *bluezController) SetVolume(_ int) error {
	return errors.New("Bluetooth AVRCP volume is not implemented; use the local sink")
}

func (c *bluezController) ListenEvents() (<-chan source.PlaybackEvent, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	if c.events == nil {
		c.events = make(chan source.PlaybackEvent, 16)
		initial := c.last
		c.mu.Unlock()
		c.emit(initial)
		go c.pollPosition()
	} else {
		c.mu.Unlock()
	}
	return c.events, nil
}

func (c *bluezController) pollPosition() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.RLock()
		conn := c.conn
		path := c.path
		previous := c.last.Position
		c.mu.RUnlock()
		if conn == nil || path == "" {
			continue
		}

		var value dbus.Variant
		call := conn.Object(bluezService, path).Call(
			propertiesIface+".Get",
			0,
			bluezPlayer,
			"Position",
		)
		if err := call.Store(&value); err != nil {
			continue
		}
		position, ok := variantInt64(value)
		if !ok || position == previous {
			continue
		}

		c.mu.Lock()
		if c.path == path {
			c.last.Position = position
			event := c.last
			c.mu.Unlock()
			c.emit(event)
			continue
		}
		c.mu.Unlock()
	}
}

func eventFromProperties(properties map[string]dbus.Variant, previous source.PlaybackEvent) source.PlaybackEvent {
	event := previous
	if value, ok := properties["Track"]; ok {
		if track, ok := value.Value().(map[string]dbus.Variant); ok {
			if title, ok := track["Title"].Value().(string); ok {
				event.Title = title
			}
			if artist, ok := track["Artist"].Value().(string); ok {
				event.Artist = artist
			}
			if album, ok := track["Album"].Value().(string); ok {
				event.Album = album
			}
			if duration, ok := track["Duration"].Value().(uint32); ok {
				event.Duration = int64(duration)
			}
		}
	}
	if value, ok := properties["Position"]; ok {
		if position, ok := variantInt64(value); ok {
			event.Position = position
		}
	}
	if value, ok := properties["Status"]; ok {
		if status, ok := value.Value().(string); ok {
			event.Playing = status == "playing"
			event.Paused = status != "playing"
		}
	}
	return event
}

func variantInt64(value dbus.Variant) (int64, bool) {
	switch number := value.Value().(type) {
	case uint32:
		return int64(number), true
	case uint64:
		return int64(number), true
	case int32:
		return int64(number), true
	case int64:
		return number, true
	default:
		return 0, false
	}
}
