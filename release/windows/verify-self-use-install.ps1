[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$IdentityFile,

    [Parameter(Mandatory = $true)]
    [string]$InstallerFile,

    [Parameter(Mandatory = $true)]
    [string]$ApplicationFile
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$identity = Get-Content -LiteralPath $IdentityFile -Raw | ConvertFrom-Json
$installer = (Resolve-Path -LiteralPath $InstallerFile).Path
$sourceApplication = (Resolve-Path -LiteralPath $ApplicationFile).Path
$installDirectory = Join-Path $env:LOCALAPPDATA "Programs\$($identity.product_name)"
$installedApplication = Join-Path $installDirectory "$($identity.executable_basename).exe"
$dataRoot = Join-Path $env:LOCALAPPDATA $identity.app_data_directory
$sentinel = Join-Path $dataRoot 'release-install-preservation.txt'
$sentinelValue = [guid]::NewGuid().ToString('N')

New-Item -ItemType Directory -Force -Path $dataRoot | Out-Null
Set-Content -LiteralPath $sentinel -Value $sentinelValue -Encoding ascii

function Invoke-Installer {
    $process = Start-Process -FilePath $installer -ArgumentList '/S' -Wait -PassThru
    if ($process.ExitCode -ne 0) {
        throw "NSIS installer exited with code $($process.ExitCode)."
    }
}

Invoke-Installer
if (-not (Test-Path -LiteralPath $installedApplication)) {
    throw "Installed application was not found at $installedApplication."
}
if ((Get-FileHash -LiteralPath $installedApplication -Algorithm SHA256).Hash -ne
    (Get-FileHash -LiteralPath $sourceApplication -Algorithm SHA256).Hash) {
    throw 'Installed application bytes differ from the verified build output.'
}

$applicationProcess = Start-Process -FilePath $installedApplication -PassThru
$database = Join-Path $dataRoot 'db\go-admin.sqlite3'
$deadline = [DateTime]::UtcNow.AddSeconds(90)
while (-not (Test-Path -LiteralPath $database) -and [DateTime]::UtcNow -lt $deadline) {
    if ($applicationProcess.HasExited) {
        throw "Installed application exited early with code $($applicationProcess.ExitCode)."
    }
    Start-Sleep -Seconds 2
}
if (-not (Test-Path -LiteralPath $database)) {
    throw 'Installed application did not create its SQLite database before timeout.'
}
& taskkill.exe /PID $applicationProcess.Id /T /F | Out-Null
$applicationProcess.WaitForExit()

Invoke-Installer
if ((Get-Content -LiteralPath $sentinel -Raw).Trim() -ne $sentinelValue) {
    throw 'Application data changed during installer upgrade.'
}

$uninstaller = Join-Path $installDirectory 'uninstall.exe'
if (-not (Test-Path -LiteralPath $uninstaller)) {
    throw 'NSIS uninstaller was not installed.'
}
$uninstallProcess = Start-Process -FilePath $uninstaller -ArgumentList '/S' -Wait -PassThru
if ($uninstallProcess.ExitCode -ne 0) {
    throw "NSIS uninstaller exited with code $($uninstallProcess.ExitCode)."
}
if ((Get-Content -LiteralPath $sentinel -Raw).Trim() -ne $sentinelValue) {
    throw 'Application data was removed by uninstall.'
}

Write-Host 'GO_ADMIN_WINDOWS_SELF_USE_INSTALL_PASS'
