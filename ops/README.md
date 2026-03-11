# Persistent Leapp WSL Bridge Setup

These scripts install the Leapp WSL shim as a `systemd` service and register Windows startup tasks for `ipc-bridge`.

## Requirements

- Leapp Desktop installed on Windows
- WSL with `systemd` enabled
- Node.js 22 or newer available in WSL
- `git`, `npm`, and `systemctl` available inside WSL
- `ipc-bridge.exe` copied to `C:\ProgramData\LeappIpcBridge\ipc-bridge.exe`

## Recommended paths

- Windows bridge binary: `C:\ProgramData\LeappIpcBridge\ipc-bridge.exe`
- Windows copied wrapper: `C:\ProgramData\LeappIpcBridge\start-ipc-bridge.ps1`
- Windows logs: `C:\ProgramData\LeappIpcBridge\logs\`
- WSL cloned repo used by the service: `/opt/leapp-ipc-bridge/repo`
- WSL env file: `/etc/default/leapp-ipc-shim`
- WSL service unit: `/etc/systemd/system/leapp-ipc-shim.service`

## Install the WSL shim service

From the `ipc-bridge` repository inside WSL:

```bash
sudo bash ops/install-wsl-shim.sh
```

If your default `node` is version 22 but `/usr/bin/node` is older, pass the Node 22 binary explicitly:

```bash
sudo NODE_BIN="$(readlink -f "$(command -v node)")" bash ops/install-wsl-shim.sh
```

The installer will:

- clone `https://github.com/yoonhyunwoo/ipc-bridge.git` into `/opt/leapp-ipc-bridge/repo`
- install the shim dependencies with `npm ci --omit=dev`
- create `/etc/default/leapp-ipc-shim`
- create and enable `leapp-ipc-shim.service`
- restart the service immediately

You can override defaults with environment variables before running the installer:

```bash
sudo LEAPP_SHIM_TCP_PORT=43827 LEAPP_SHIM_LOG_LEVEL=debug bash ops/install-wsl-shim.sh
```

## Register the Windows startup tasks

Open an elevated PowerShell prompt in the `ipc-bridge` repository and run:

```powershell
Set-ExecutionPolicy -Scope Process Bypass
.\ops\register-ipc-bridge-tasks.ps1 -BridgeExePath "C:\ProgramData\LeappIpcBridge\ipc-bridge.exe"
```

If you want to target a specific WSL distro instead of the default one:

```powershell
.\ops\register-ipc-bridge-tasks.ps1 -BridgeExePath "C:\ProgramData\LeappIpcBridge\ipc-bridge.exe" -WslDistro "Ubuntu-24.04"
```

The script will:

- copy `start-ipc-bridge.ps1` to `C:\ProgramData\LeappIpcBridge\start-ipc-bridge.ps1`
- register `\Leapp\LeappIpcBridge`
- register `\Leapp\LeappWslShimBootstrap`

The PowerShell session must be running as Administrator, otherwise `Register-ScheduledTask` will fail with `0x80070005` access denied.

## Verify the setup

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
