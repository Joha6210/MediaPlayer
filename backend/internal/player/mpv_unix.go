//go:build !windows

package player

import (
	"net"
	"time"
)

func dialSocket(path string) (net.Conn, error) {
	timeout := 2 * time.Second
	return net.DialTimeout("unix", path, timeout)
}
