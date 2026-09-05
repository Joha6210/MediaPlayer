//go:build windows

package player

import (
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

func dialSocket(path string) (net.Conn, error) {
	timeout := 2 * time.Second
	return winio.DialPipe(path, &timeout)
}
