package endpoint

import "testing"

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    Spec
		wantErr string
	}{
		{
			name: "tcp",
			raw:  "tcp://127.0.0.1:1234",
			want: Spec{Kind: KindTCP, Address: "127.0.0.1:1234", Raw: "tcp://127.0.0.1:1234"},
		},
		{
			name: "unix",
			raw:  "unix:///tmp/test.sock",
			want: Spec{Kind: KindUnix, Address: "/tmp/test.sock", Raw: "unix:///tmp/test.sock"},
		},
		{
			name: "npipe normalized",
			raw:  "npipe:////./pipe/demo",
			want: Spec{Kind: KindNPipe, Address: `\\.\pipe\demo`, Raw: "npipe:////./pipe/demo"},
		},
		{
			name:    "unsupported",
			raw:     "udp://127.0.0.1:1",
			wantErr: `unsupported endpoint "udp://127.0.0.1:1"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(tt.raw)
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
				t.Fatalf("unexpected spec: %#v", got)
			}
		})
	}
}
