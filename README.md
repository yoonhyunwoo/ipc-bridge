# ipc-bridge

`ipc-bridge` is a small Go CLI for forwarding byte streams between local IPC endpoints and network sockets.

## Supported Endpoints

- `tcp://host:port`
- `unix:///absolute/path.sock`
- `npipe:////./pipe/name` on Windows

## Example

Expose a Windows named pipe on localhost:

```bash
ipc-bridge --from=tcp://127.0.0.1:43827 --to=npipe:////./pipe/leapp_da
```

Expose a Unix socket and forward it to a TCP bridge:

```bash
ipc-bridge --from=unix:///tmp/leapp_da.sock --to=tcp://127.0.0.1:43827
```

## Build

```bash
go build ./cmd/ipc-bridge
```
