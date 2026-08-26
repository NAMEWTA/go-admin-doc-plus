[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$Executable,

    [Parameter(Mandatory = $true)]
    [string]$DataRoot,

    [Parameter(Mandatory = $true)]
    [string]$LogDirectory,

    [ValidateRange(30, 600)]
    [int]$TimeoutSeconds = 180
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

if ($env:OS -ne 'Windows_NT') {
    throw 'The native desktop tracer requires Windows.'
}

$executablePath = (Resolve-Path -LiteralPath $Executable).Path
$dataRootPath = [System.IO.Path]::GetFullPath($DataRoot)
$logDirectoryPath = [System.IO.Path]::GetFullPath($LogDirectory)

New-Item -ItemType Directory -Force -Path $dataRootPath | Out-Null
New-Item -ItemType Directory -Force -Path $logDirectoryPath | Out-Null

$stdoutPath = Join-Path $logDirectoryPath 'desktop.stdout.log'
$stderrPath = Join-Path $logDirectoryPath 'desktop.stderr.log'
$env:GO_ADMIN_DESKTOP_DATA_ROOT = $dataRootPath

$process = $null
try {
    $startArguments = @{
        FilePath = $executablePath
        WorkingDirectory = $dataRootPath
        RedirectStandardOutput = $stdoutPath
        RedirectStandardError = $stderrPath
        PassThru = $true
    }
    $process = Start-Process @startArguments

    if (-not $process.WaitForExit($TimeoutSeconds * 1000)) {
        & taskkill.exe /PID $process.Id /T /F | Out-Null
        throw "Desktop tracer timed out after $TimeoutSeconds seconds."
    }
    $process.WaitForExit()

    $stdout = if (Test-Path -LiteralPath $stdoutPath) {
        @(Get-Content -LiteralPath $stdoutPath)
    } else {
        @()
    }
    $stderr = if (Test-Path -LiteralPath $stderrPath) {
        @(Get-Content -LiteralPath $stderrPath)
    } else {
        @()
    }

    Write-Host '--- desktop stdout ---'
    $stdout | ForEach-Object { Write-Host $_ }
    Write-Host '--- desktop stderr ---'
    $stderr | ForEach-Object { Write-Host $_ }

    if ($process.ExitCode -ne 0) {
        throw "Desktop tracer exited with code $($process.ExitCode)."
    }

    $passCount = @($stdout | Where-Object {
        $_.Trim() -eq 'GO_ADMIN_DESKTOP_E2E_PASS'
    }).Count
    $failureLines = @($stdout + $stderr | Where-Object {
        $_ -match 'GO_ADMIN_DESKTOP_E2E_FAIL'
    })

    if ($failureLines.Count -gt 0) {
        throw "Desktop tracer reported failure: $($failureLines -join '; ')"
    }
    if ($passCount -ne 1) {
        throw "Expected exactly one GO_ADMIN_DESKTOP_E2E_PASS marker, found $passCount."
    }
} finally {
    if ($null -ne $process -and -not $process.HasExited) {
        & taskkill.exe /PID $process.Id /T /F | Out-Null
    }
}

Write-Host 'Windows native desktop tracer passed.'
