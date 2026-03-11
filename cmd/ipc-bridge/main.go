package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"ipc-bridge/internal/endpoint"
)

type closeWriter interface {
	CloseWrite() error
}

func main() {
	var (
		fromRaw  string
		toRaw    string
		once     bool
		maxConns int
		timeout  time.Duration
		logLevel string
	)

	flag.StringVar(&fromRaw, "from", "", "listener endpoint: tcp://host:port, unix:///path, or npipe:////./pipe/name")
	flag.StringVar(&toRaw, "to", "", "dial endpoint: tcp://host:port, unix:///path, or npipe:////./pipe/name")
	flag.BoolVar(&once, "once", false, "accept a single incoming connection and exit after it closes")
	flag.IntVar(&maxConns, "max-conns", 16, "maximum concurrent bridged connections")
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "dial timeout for outbound connections")
	flag.StringVar(&logLevel, "log-level", "info", "log verbosity: error, info, debug")
	flag.Parse()

	if fromRaw == "" || toRaw == "" {
		flag.Usage()
		os.Exit(2)
	}

	if maxConns < 1 {
		log.Fatalf("invalid --max-conns: %d", maxConns)
	}

	logger := newLogger(logLevel)

	fromSpec, err := endpoint.Parse(fromRaw)
	if err != nil {
		log.Fatalf("invalid --from: %v", err)
	}

	toSpec, err := endpoint.Parse(toRaw)
	if err != nil {
		log.Fatalf("invalid --to: %v", err)
	}

	listener, err := endpoint.Listen(fromSpec)
	if err != nil {
		log.Fatalf("listen failed: %v", err)
	}
	defer listener.Close()

	logger.info("listening on %s and forwarding to %s", fromSpec.String(), toSpec.String())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals
		logger.info("shutdown requested")
		_ = listener.Close()
	}()

	sem := make(chan struct{}, maxConns)
	var wg sync.WaitGroup
	accepted := 0

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}

			logger.error("accept failed: %v", err)
			continue
		}

		accepted++
		sem <- struct{}{}
		wg.Add(1)

		go func(id int, incoming net.Conn) {
			defer func() {
				<-sem
				wg.Done()
			}()

			if err := bridgeConnection(id, incoming, toSpec, timeout, logger); err != nil {
				logger.error("connection %d failed: %v", id, err)
			}
		}(accepted, conn)

		if once {
			logger.info("--once enabled; waiting for connection %d to finish", accepted)
			break
		}
	}

	wg.Wait()
}

func bridgeConnection(id int, incoming net.Conn, target endpoint.Spec, timeout time.Duration, logger *leveledLogger) error {
	defer incoming.Close()

	outgoing, err := endpoint.Dial(target, timeout)
	if err != nil {
		return fmt.Errorf("dial %s: %w", target.String(), err)
	}
	defer outgoing.Close()

	logger.info("connection %d established: %s -> %s", id, incoming.RemoteAddr(), target.String())

	errCh := make(chan error, 2)

	go relay(errCh, outgoing, incoming)
	go relay(errCh, incoming, outgoing)

	firstErr := <-errCh
	secondErr := <-errCh

	if err := normalizeRelayError(firstErr); err != nil {
		return err
	}

	if err := normalizeRelayError(secondErr); err != nil {
		return err
	}

	logger.debug("connection %d closed cleanly", id)
	return nil
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
	return netCopy(dst, src)
}
