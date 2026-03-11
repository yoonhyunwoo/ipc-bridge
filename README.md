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

Expose a local `node-ipc` server for `leapp-cli` inside WSL and forward it to the TCP bridge:

```bash
cd shim
npm install
npm start
```

The expected runtime chain for Leapp on WSL is:

```text
leapp-cli -> node-ipc shim -> tcp://127.0.0.1:43827 -> ipc-bridge -> npipe:////./pipe/leapp_da
```

To keep both sides running across reboots, use the installation assets in [`ops/README.md`](ops/README.md). The recommended Windows binary path is `C:\ProgramData\LeappIpcBridge\ipc-bridge.exe`.

## Build

```bash
go build ./cmd/ipc-bridge
```
