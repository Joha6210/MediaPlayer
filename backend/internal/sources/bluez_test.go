package sources

import (
	"testing"

	"github.com/godbus/dbus/v5"

	"mediaplayer/backend/internal/source"
)

func TestEventFromPropertiesDecodesTrackAndPlayback(t *testing.T) {
	properties := map[string]dbus.Variant{
		"Track": dbus.MakeVariant(map[string]dbus.Variant{
			"Title":    dbus.MakeVariant("Song"),
			"Artist":   dbus.MakeVariant("Artist"),
			"Album":    dbus.MakeVariant("Album"),
			"Duration": dbus.MakeVariant(uint32(245000000)),
		}),
		"Position": dbus.MakeVariant(uint32(12000000)),
		"Status":   dbus.MakeVariant("playing"),
	}

	event := eventFromProperties(properties, source.PlaybackEvent{})
	if event.Title != "Song" || event.Artist != "Artist" || event.Album != "Album" {
		t.Fatalf("unexpected metadata: %+v", event)
	}
	if event.Duration != 245000000 || event.Position != 12000000 || !event.Playing || event.Paused {
		t.Fatalf("unexpected playback state: %+v", event)
	}
}
