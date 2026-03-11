package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"ipc-bridge/internal/endpoint"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		wantErr   string
		wantUsage bool
		want      bridgeConfig
	}{
		{
			name:      "missing required endpoints",
			args:      nil,
			wantUsage: true,
		},
		{
			name:    "invalid max conns",
			args:    []string{"--from=tcp://127.0.0.1:1", "--to=unix:///tmp/out.sock", "--max-conns=0"},
			wantErr: "invalid --max-conns: 0",
		},
		{
			name:      "invalid duration is usage",
			args:      []string{"--from=tcp://127.0.0.1:1", "--to=unix:///tmp/out.sock", "--timeout=bad"},
			wantUsage: true,
		},
		{
			name:    "invalid from endpoint",
			args:    []string{"--from=bad://thing", "--to=unix:///tmp/out.sock"},
			wantErr: `invalid --from: unsupported endpoint "bad://thing"`,
		},
		{
			name: "valid config",
			args: []string{
				"--from=tcp://127.0.0.1:1",
				"--to=unix:///tmp/out.sock",
				"--once",
				"--max-conns=3",
				"--timeout=3s",
				"--log-level=debug",
			},
			want: bridgeConfig{
				fromSpec: endpoint.Spec{Kind: endpoint.KindTCP, Address: "127.0.0.1:1", Raw: "tcp://127.0.0.1:1"},
				toSpec:   endpoint.Spec{Kind: endpoint.KindUnix, Address: "/tmp/out.sock", Raw: "unix:///tmp/out.sock"},
				once:     true,
				maxConns: 3,
				timeout:  3 * time.Second,
				logLevel: "debug",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer
			got, err := parseConfig(tt.args, &output)

			if tt.wantUsage {
				if !errors.Is(err, errUsage) {
					t.Fatalf("expected usage error, got %v", err)
				}
				if output.Len() == 0 {
					t.Fatal("expected usage output")
				}
				return
			}

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected config: %#v", got)
			}
			if strings.Contains(output.String(), "Usage") {
				t.Fatalf("did not expect usage output, got %q", output.String())
			}
		})
	}
}
