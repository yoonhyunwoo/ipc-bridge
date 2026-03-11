package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"ipc-bridge/internal/endpoint"
)

var (
	listenEndpoint = endpoint.Listen
	dialEndpoint   = endpoint.Dial
)

func runBridge(cfg bridgeConfig, logger *leveledLogger) error {
	listener, err := listenEndpoint(cfg.fromSpec)
	if err != nil {
		return fmt.Errorf("listen failed: %w", err)
	}
	defer listener.Close()

	logger.info("listening on %s and forwarding to %s", cfg.fromSpec.String(), cfg.toSpec.String())

	stopSignalWatch := closeListenerOnSignals(listener, logger)
	defer stopSignalWatch()

	return serveConnections(listener, cfg, logger)
}

func closeListenerOnSignals(listener net.Listener, logger *leveledLogger) func() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		select {
		case <-signals:
			logger.info("shutdown requested")
			_ = listener.Close()
		case <-done:
		}
	}()

	return func() {
		close(done)
		signal.Stop(signals)
	}
}

func serveConnections(listener net.Listener, cfg bridgeConfig, logger *leveledLogger) error {
	sem := make(chan struct{}, cfg.maxConns)
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

			if err := bridgeConnection(id, incoming, cfg.toSpec, cfg.timeout, logger); err != nil {
				logger.error("connection %d failed: %v", id, err)
			}
		}(accepted, conn)

		if cfg.once {
			logger.info("--once enabled; waiting for connection %d to finish", accepted)
			break
		}
	}

	wg.Wait()
	return nil
}

func bridgeConnection(id int, incoming net.Conn, target endpoint.Spec, timeout time.Duration, logger *leveledLogger) error {
	defer incoming.Close()

	outgoing, err := dialEndpoint(target, timeout)
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
