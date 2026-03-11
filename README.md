# ipc-bridge

`ipc-bridge` is a small Go CLI for forwarding byte streams between local IPC endpoints and network sockets.

## Supported Endpoints

- `tcp://host:port`
- `unix:///absolute/path.sock`
- `npipe:////./pipe/name` on Windows

## Examples

Expose a Windows named pipe on localhost:

```bash
ipc-bridge --from=tcp://127.0.0.1:43827 --to=npipe:////./pipe/my_app
```

Expose a Unix socket and forward it to a TCP bridge:

```bash
ipc-bridge --from=unix:///tmp/my-app.sock --to=tcp://127.0.0.1:43827
```

Expose a TCP socket and forward it to a Unix socket:

```bash
ipc-bridge --from=tcp://127.0.0.1:9000 --to=unix:///tmp/my-app.sock
```

For a runnable Leapp integration example with the Node shim and Windows/WSL startup automation, see [`examples/leapp/README.md`](examples/leapp/README.md).

## Build

```bash
go build ./cmd/ipc-bridge
```

## Test

```bash
go test ./...
```
