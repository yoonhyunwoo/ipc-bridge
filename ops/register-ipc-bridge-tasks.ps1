[CmdletBinding()]
param(
    [string]$BridgeExePath = "C:\Program Files\LeappIpcBridge\ipc-bridge.exe",
    [string]$WrapperTargetPath = "C:\ProgramData\LeappIpcBridge\start-ipc-bridge.ps1",
    [string]$ListenEndpoint = "tcp://127.0.0.1:43827",
    [string]$TargetEndpoint = "npipe:////./pipe/leapp_da",
    [string]$WslDistro = "",
    [string]$WslSystemctlPath = "/usr/bin/systemctl",
    [string]$ShimServiceName = "leapp-ipc-shim.service",
    [string]$TaskFolder = "\Leapp\",
    [int]$BridgeDelaySeconds = 5,
    [int]$ShimDelaySeconds = 15
)

$ErrorActionPreference = "Stop"

if (!(Test-Path -LiteralPath $BridgeExePath)) {
    throw "ipc-bridge executable not found at $BridgeExePath"
}

$scriptSource = Split-Path -Parent $MyInvocation.MyCommand.Path
$wrapperSource = Join-Path $scriptSource "start-ipc-bridge.ps1"
if (!(Test-Path -LiteralPath $wrapperSource)) {
    throw "wrapper script not found next to register script: $wrapperSource"
}

$wrapperDir = Split-Path -Parent $WrapperTargetPath
New-Item -ItemType Directory -Force -Path $wrapperDir | Out-Null
Copy-Item -LiteralPath $wrapperSource -Destination $WrapperTargetPath -Force

$taskPrincipal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest
$startupTrigger = New-ScheduledTaskTrigger -AtStartup
$startupTrigger.Delay = "PT${BridgeDelaySeconds}S"

$bridgeAction = New-ScheduledTaskAction `
    -Execute "powershell.exe" `
    -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$WrapperTargetPath`" -BridgeExePath `"$BridgeExePath`" -ListenEndpoint `"$ListenEndpoint`" -TargetEndpoint `"$TargetEndpoint`""
$bridgeSettings = New-ScheduledTaskSettingsSet `
    -AllowStartIfOnBatteries `
    -DontStopIfGoingOnBatteries `
    -ExecutionTimeLimit (New-TimeSpan -Hours 0) `
    -MultipleInstances IgnoreNew `
    -RestartCount 999 `
    -RestartInterval (New-TimeSpan -Minutes 1) `
    -StartWhenAvailable

Register-ScheduledTask `
    -TaskName "LeappIpcBridge" `
    -TaskPath $TaskFolder `
    -Action $bridgeAction `
    -Trigger $startupTrigger `
    -Principal $taskPrincipal `
    -Settings $bridgeSettings `
    -Force | Out-Null

$shimCommandParts = @()
if ($WslDistro) {
    $shimCommandParts += "-d"
    $shimCommandParts += $WslDistro
}
$shimCommandParts += "--user"
$shimCommandParts += "root"
$shimCommandParts += "--exec"
$shimCommandParts += $WslSystemctlPath
$shimCommandParts += "start"
$shimCommandParts += $ShimServiceName
$shimCommand = ($shimCommandParts | ForEach-Object {
    if ($_ -match '\s') { '"' + $_ + '"' } else { $_ }
}) -join " "

$shimTrigger = New-ScheduledTaskTrigger -AtStartup
$shimTrigger.Delay = "PT${ShimDelaySeconds}S"
$shimAction = New-ScheduledTaskAction -Execute "wsl.exe" -Argument $shimCommand
$shimSettings = New-ScheduledTaskSettingsSet `
    -AllowStartIfOnBatteries `
    -DontStopIfGoingOnBatteries `
    -ExecutionTimeLimit (New-TimeSpan -Hours 0) `
    -MultipleInstances IgnoreNew `
    -RestartCount 999 `
    -RestartInterval (New-TimeSpan -Minutes 1) `
    -StartWhenAvailable

Register-ScheduledTask `
    -TaskName "LeappWslShimBootstrap" `
    -TaskPath $TaskFolder `
    -Action $shimAction `
    -Trigger $shimTrigger `
    -Principal $taskPrincipal `
    -Settings $shimSettings `
    -Force | Out-Null

Write-Host "Registered tasks:"
Write-Host " - ${TaskFolder}LeappIpcBridge"
Write-Host " - ${TaskFolder}LeappWslShimBootstrap"
