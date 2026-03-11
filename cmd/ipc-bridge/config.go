package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"ipc-bridge/internal/endpoint"
)

var errUsage = errors.New("usage")

type bridgeConfig struct {
	fromSpec endpoint.Spec
	toSpec   endpoint.Spec
	once     bool
	maxConns int
	timeout  time.Duration
	logLevel string
}

func parseConfig(args []string, output io.Writer) (bridgeConfig, error) {
	cfg := bridgeConfig{}

	fs := flag.NewFlagSet("ipc-bridge", flag.ContinueOnError)
	fs.SetOutput(output)

	var fromRaw string
	var toRaw string

	fs.StringVar(&fromRaw, "from", "", "listener endpoint: tcp://host:port, unix:///path, or npipe:////./pipe/name")
	fs.StringVar(&toRaw, "to", "", "dial endpoint: tcp://host:port, unix:///path, or npipe:////./pipe/name")
	fs.BoolVar(&cfg.once, "once", false, "accept a single incoming connection and exit after it closes")
	fs.IntVar(&cfg.maxConns, "max-conns", 16, "maximum concurrent bridged connections")
	fs.DurationVar(&cfg.timeout, "timeout", 10*time.Second, "dial timeout for outbound connections")
	fs.StringVar(&cfg.logLevel, "log-level", "info", "log verbosity: error, info, debug")

	if err := fs.Parse(args); err != nil {
		return bridgeConfig{}, fmt.Errorf("%w: %v", errUsage, err)
	}

	if fromRaw == "" || toRaw == "" {
		fs.Usage()
		return bridgeConfig{}, errUsage
	}

	if cfg.maxConns < 1 {
		return bridgeConfig{}, fmt.Errorf("invalid --max-conns: %d", cfg.maxConns)
	}

	fromSpec, err := endpoint.Parse(fromRaw)
	if err != nil {
		return bridgeConfig{}, fmt.Errorf("invalid --from: %w", err)
	}

	toSpec, err := endpoint.Parse(toRaw)
	if err != nil {
		return bridgeConfig{}, fmt.Errorf("invalid --to: %w", err)
	}

	cfg.fromSpec = fromSpec
	cfg.toSpec = toSpec

	return cfg, nil
}
