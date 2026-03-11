# Leapp Integration Example

This example shows how to connect Leapp on Windows and WSL through `ipc-bridge`.

The runtime chain is:

```text
leapp-cli -> node-ipc shim -> tcp://127.0.0.1:43827 -> ipc-bridge -> npipe:////./pipe/leapp_da
```

## Layout

- Node shim: `examples/leapp/shim/`
- WSL installer: `examples/leapp/ops/install-wsl-shim.sh`
- Windows startup scripts: `examples/leapp/ops/register-ipc-bridge-tasks.ps1` and `examples/leapp/ops/start-ipc-bridge.ps1`

## Requirements

- Leapp Desktop installed on Windows
- WSL with `systemd` enabled
- Node.js 22 or newer available in WSL
- `git`, `npm`, and `systemctl` available inside WSL
- `ipc-bridge.exe` copied to `C:\ProgramData\LeappIpcBridge\ipc-bridge.exe`

## Shared Defaults

- Node IPC server id: `leapp_da`
- TCP bridge endpoint: `tcp://127.0.0.1:43827`
- Windows named pipe target: `npipe:////./pipe/leapp_da`
- Windows bridge binary: `C:\ProgramData\LeappIpcBridge\ipc-bridge.exe`
- Windows copied wrapper: `C:\ProgramData\LeappIpcBridge\start-ipc-bridge.ps1`
- Windows logs: `C:\ProgramData\LeappIpcBridge\logs\`
- WSL cloned repo used by the service: `/opt/leapp-ipc-bridge/repo`
- WSL env file: `/etc/default/leapp-ipc-shim`
- WSL service unit: `/etc/systemd/system/leapp-ipc-shim.service`

## Run the Shim Manually

From the repository root inside WSL:

```bash
cd examples/leapp/shim
npm install
npm start
```

Environment defaults used by the shim:

- `LEAPP_SHIM_SERVER_ID=leapp_da`
- `LEAPP_SHIM_TCP_HOST=127.0.0.1`
- `LEAPP_SHIM_TCP_PORT=43827`
- `LEAPP_SHIM_LOG_LEVEL=info`

## Install the WSL Shim Service

From the repository root inside WSL:

```bash
sudo bash examples/leapp/ops/install-wsl-shim.sh
```

If your default `node` is version 22 but `/usr/bin/node` is older, pass the Node 22 binary explicitly:

```bash
sudo NODE_BIN="$(readlink -f "$(command -v node)")" bash examples/leapp/ops/install-wsl-shim.sh
```

The installer will:

- clone `https://github.com/yoonhyunwoo/ipc-bridge.git` into `/opt/leapp-ipc-bridge/repo`
- install the shim dependencies with `npm ci --omit=dev`
- create `/etc/default/leapp-ipc-shim`
- create and enable `leapp-ipc-shim.service`
- restart the service immediately

You can override defaults with environment variables before running the installer:

```bash
sudo LEAPP_SHIM_TCP_PORT=43827 LEAPP_SHIM_LOG_LEVEL=debug bash examples/leapp/ops/install-wsl-shim.sh
```

## Register the Windows Startup Tasks

Open an elevated PowerShell prompt in the repository root and run:

```powershell
Set-ExecutionPolicy -Scope Process Bypass
.\examples\leapp\ops\register-ipc-bridge-tasks.ps1 -BridgeExePath "C:\ProgramData\LeappIpcBridge\ipc-bridge.exe"
```

If you want to target a specific WSL distro instead of the default one:

```powershell
.\examples\leapp\ops\register-ipc-bridge-tasks.ps1 -BridgeExePath "C:\ProgramData\LeappIpcBridge\ipc-bridge.exe" -WslDistro "Ubuntu-24.04"
```

The script will:

- copy `start-ipc-bridge.ps1` to `C:\ProgramData\LeappIpcBridge\start-ipc-bridge.ps1`
- register `\Leapp\LeappIpcBridge`
- register `\Leapp\LeappWslShimBootstrap`

The PowerShell session must be running as Administrator, otherwise `Register-ScheduledTask` will fail with `0x80070005` access denied.

## Verify the Setup

In WSL:

```bash
systemctl status leapp-ipc-shim.service
journalctl -u leapp-ipc-shim -n 100 --no-pager
```

In Windows PowerShell:

```powershell
Get-ScheduledTask -TaskPath "\Leapp\" -TaskName "LeappIpcBridge","LeappWslShimBootstrap"
Start-ScheduledTask -TaskPath "\Leapp\" -TaskName "LeappIpcBridge"
Start-ScheduledTask -TaskPath "\Leapp\" -TaskName "LeappWslShimBootstrap"
Get-Content C:\ProgramData\LeappIpcBridge\logs\*.stderr.log -Tail 50
```

Expected healthy status:

- `leapp-ipc-shim.service` is `enabled` and `active`
- `\Leapp\LeappIpcBridge` and `\Leapp\LeappWslShimBootstrap` are registered and runnable
- bridge stderr log contains:

```text
[ipc-bridge] ... INFO listening on tcp://127.0.0.1:43827 and forwarding to npipe:////./pipe/leapp_da
```
