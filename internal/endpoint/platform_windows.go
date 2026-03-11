//go:build windows

package endpoint

import (
	"fmt"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

func listenPlatform(spec Spec) (net.Listener, error) {
	if spec.Kind != KindNPipe {
		return nil, fmt.Errorf("unsupported listener endpoint: %s", spec.Kind)
	}

	return winio.ListenPipe(spec.Address, nil)
}

func dialPlatform(spec Spec, timeout time.Duration) (net.Conn, error) {
	if spec.Kind != KindNPipe {
		return nil, fmt.Errorf("unsupported dial endpoint: %s", spec.Kind)
	}

	return winio.DialPipe(spec.Address, &timeout)
}
