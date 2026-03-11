package main

import (
	"errors"
	"log"
	"net"
	"os"
)

type closeWriter interface {
	CloseWrite() error
}

func main() {
	cfg, err := parseConfig(os.Args[1:], os.Stderr)
	if err != nil {
		if errors.Is(err, errUsage) {
			os.Exit(2)
		}
		log.Fatal(err)
	}

	if err := runBridge(cfg, newLogger(cfg.logLevel)); err != nil {
		log.Fatal(err)
	}
}

func relay(errCh chan<- error, dst net.Conn, src net.Conn) {
	_, err := ioCopy(dst, src)
	if cw, ok := dst.(closeWriter); ok {
		_ = cw.CloseWrite()
	} else {
		_ = dst.Close()
	}
	errCh <- err
}

func normalizeRelayError(err error) error {
	if err == nil || errors.Is(err, net.ErrClosed) {
		return nil
	}

	return err
}

func ioCopy(dst net.Conn, src net.Conn) (int64, error) {
	return copyConn(dst, src)
}
