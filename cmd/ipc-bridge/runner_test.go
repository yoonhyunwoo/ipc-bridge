package main

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"ipc-bridge/internal/endpoint"
)

func TestNormalizeRelayError(t *testing.T) {
	t.Parallel()

	if err := normalizeRelayError(nil); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if err := normalizeRelayError(net.ErrClosed); err != nil {
		t.Fatalf("expected nil for net.ErrClosed, got %v", err)
	}

	want := io.EOF
	if err := normalizeRelayError(want); !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}

func TestBridgeConnectionDialError(t *testing.T) {
	t.Parallel()

	originalDial := dialEndpoint
	t.Cleanup(func() {
		dialEndpoint = originalDial
	})

	dialEndpoint = func(endpoint.Spec, time.Duration) (net.Conn, error) {
		return nil, errors.New("boom")
	}

	err := bridgeConnection(7, &stubConn{}, endpoint.Spec{Raw: "tcp://target"}, time.Second, newLogger("error"))
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "dial tcp://target: boom" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRelayPrefersCloseWrite(t *testing.T) {
	t.Parallel()

	originalCopy := copyConn
	t.Cleanup(func() {
		copyConn = originalCopy
	})

	copyConn = func(dst net.Conn, src net.Conn) (int64, error) {
		return 0, nil
	}

	errCh := make(chan error, 1)
	dst := &stubConn{}
	src := &stubConn{}

	relay(errCh, dst, src)

	if err := <-errCh; err != nil {
		t.Fatalf("unexpected relay error: %v", err)
	}
	if !dst.closeWriteCalled {
		t.Fatal("expected CloseWrite to be called")
	}
	if dst.closeCalled {
		t.Fatal("expected Close to be skipped when CloseWrite is available")
	}
}

type stubConn struct {
	closeCalled      bool
	closeWriteCalled bool
}

func (c *stubConn) Read([]byte) (int, error)         { return 0, io.EOF }
func (c *stubConn) Write(p []byte) (int, error)      { return len(p), nil }
func (c *stubConn) Close() error                     { c.closeCalled = true; return nil }
func (c *stubConn) LocalAddr() net.Addr              { return stubAddr("local") }
func (c *stubConn) RemoteAddr() net.Addr             { return stubAddr("remote") }
func (c *stubConn) SetDeadline(time.Time) error      { return nil }
func (c *stubConn) SetReadDeadline(time.Time) error  { return nil }
func (c *stubConn) SetWriteDeadline(time.Time) error { return nil }
func (c *stubConn) CloseWrite() error                { c.closeWriteCalled = true; return nil }

type stubAddr string

func (a stubAddr) Network() string { return string(a) }
func (a stubAddr) String() string  { return string(a) }
