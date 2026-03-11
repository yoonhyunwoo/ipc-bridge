//go:build !windows

package endpoint

import (
	"fmt"
	"net"
	"time"
)

func listenPlatform(spec Spec) (net.Listener, error) {
	return nil, fmt.Errorf("%s endpoints are only supported on Windows", spec.Kind)
}

func dialPlatform(spec Spec, _ time.Duration) (net.Conn, error) {
	return nil, fmt.Errorf("%s endpoints are only supported on Windows", spec.Kind)
}
