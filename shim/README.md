# Leapp Node-IPC Shim

This shim exposes a local `node-ipc` server named `leapp_da` inside WSL and forwards each client connection to the Go bridge over TCP.

## Environment

- `LEAPP_SHIM_SERVER_ID` default: `leapp_da`
- `LEAPP_SHIM_TCP_HOST` default: `127.0.0.1`
- `LEAPP_SHIM_TCP_PORT` default: `43827`
- `LEAPP_SHIM_LOG_LEVEL` default: `info`

## Run

```bash
npm install
npm start
```

The expected runtime order is:

1. Leapp Desktop on Windows
2. `ipc-bridge` exposing the Windows named pipe on TCP
3. This shim inside WSL
4. `leapp-cli` inside WSL

For persistent startup with Windows Task Scheduler and a WSL `systemd` service, see [`../ops/README.md`](../ops/README.md). If `/usr/bin/node` is not Node 22+, install with `NODE_BIN="$(readlink -f "$(command -v node)")"` when running the WSL installer.
