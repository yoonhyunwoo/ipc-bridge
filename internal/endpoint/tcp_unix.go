package endpoint

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

type cleanupListener struct {
	net.Listener
	cleanup func() error
}

func (l cleanupListener) Close() error {
	err := l.Listener.Close()
	if l.cleanup != nil {
		if cleanupErr := l.cleanup(); cleanupErr != nil && err == nil {
			err = cleanupErr
		}
	}
	return err
}

func Listen(spec Spec) (net.Listener, error) {
	switch spec.Kind {
	case KindTCP:
		return net.Listen("tcp", spec.Address)
	case KindUnix:
		if spec.Address == "" {
			return nil, fmt.Errorf("unix socket path is empty")
		}

		if err := os.MkdirAll(filepath.Dir(spec.Address), 0o755); err != nil {
			return nil, err
		}

		if err := os.Remove(spec.Address); err != nil && !os.IsNotExist(err) {
			return nil, err
		}

		listener, err := net.Listen("unix", spec.Address)
		if err != nil {
			return nil, err
		}

		return cleanupListener{
			Listener: listener,
			cleanup: func() error {
				if err := os.Remove(spec.Address); err != nil && !os.IsNotExist(err) {
					return err
				}
				return nil
			},
		}, nil
	default:
		return listenPlatform(spec)
	}
}

func Dial(spec Spec, timeout time.Duration) (net.Conn, error) {
	switch spec.Kind {
	case KindTCP:
		return net.DialTimeout("tcp", spec.Address, timeout)
	case KindUnix:
		return net.DialTimeout("unix", spec.Address, timeout)
	default:
		return dialPlatform(spec, timeout)
	}
}
