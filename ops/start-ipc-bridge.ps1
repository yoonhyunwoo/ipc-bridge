[CmdletBinding()]
param(
    [string]$BridgeExePath = "C:\Program Files\LeappIpcBridge\ipc-bridge.exe",
    [string]$ListenEndpoint = "tcp://127.0.0.1:43827",
    [string]$TargetEndpoint = "npipe:////./pipe/leapp_da",
    [string]$LogRoot = "C:\ProgramData\LeappIpcBridge\logs"
)

$ErrorActionPreference = "Stop"

if (!(Test-Path -LiteralPath $BridgeExePath)) {
    throw "ipc-bridge executable not found at $BridgeExePath"
}

New-Item -ItemType Directory -Force -Path $LogRoot | Out-Null

$processName = [System.IO.Path]::GetFileNameWithoutExtension($BridgeExePath)
$running = Get-CimInstance Win32_Process -Filter "Name = '$($processName).exe'" |
    Where-Object {
        $_.ExecutablePath -eq $BridgeExePath -and
        $_.CommandLine -like "*--from=$ListenEndpoint*" -and
        $_.CommandLine -like "*--to=$TargetEndpoint*"
    }

if ($running) {
    Write-Host "ipc-bridge is already running"
    exit 0
}

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$stdoutPath = Join-Path $LogRoot "ipc-bridge-$timestamp.stdout.log"
$stderrPath = Join-Path $LogRoot "ipc-bridge-$timestamp.stderr.log"

$argumentList = @(
    "--from=$ListenEndpoint"
    "--to=$TargetEndpoint"
)

Start-Process -FilePath $BridgeExePath `
    -ArgumentList $argumentList `
    -WindowStyle Hidden `
    -RedirectStandardOutput $stdoutPath `
    -RedirectStandardError $stderrPath
