package endpoint

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestListenUnixRemovesSocketOnClose(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix sockets are not available on Windows")
	}

	socketPath := filepath.Join(t.TempDir(), "bridge.sock")

	listener, err := Listen(Spec{
		Kind:    KindUnix,
		Address: socketPath,
		Raw:     "unix://" + socketPath,
	})
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	if err := listener.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	if _, err := parseSocketStat(socketPath); err == nil {
		t.Fatalf("expected socket %q to be removed", socketPath)
	}
}

func TestListenUnixRejectsEmptyPath(t *testing.T) {
	t.Parallel()

	_, err := Listen(Spec{Kind: KindUnix})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "unix socket path is empty" {
		t.Fatalf("unexpected error: %v", err)
	}
}
