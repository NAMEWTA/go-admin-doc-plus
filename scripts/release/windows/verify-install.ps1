[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)] [string] $InstallerFile,
    [Parameter(Mandatory = $true)] [string] $EvidenceFile
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
if ($env:CI -ne 'true' -or $env:GITHUB_ACTIONS -ne 'true' -or [string]::IsNullOrWhiteSpace($env:RUNNER_TEMP)) {
    throw 'Install verification is restricted to an ephemeral GitHub Actions runner.'
}
$repository = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
$identity = Get-Content -LiteralPath (Join-Path $repository 'release/windows/identity.json') -Raw | ConvertFrom-Json
$installer = (Resolve-Path -LiteralPath $InstallerFile).Path
$installDirectory = Join-Path $env:RUNNER_TEMP "go-admin-plus-install-$([guid]::NewGuid().ToString('N'))"
$dataRoot = Join-Path $installDirectory 'data'
$logRoot = Join-Path $installDirectory 'logs'
$credentialTarget = 'desktop-session-vault.com.goadmin.plus.stronghold'
if (Test-Path -LiteralPath $installDirectory) { throw 'Install directory already exists.' }

Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;

public static class CredentialProbe
{
    [DllImport("advapi32.dll", EntryPoint = "CredReadW", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern bool CredRead(string target, uint type, uint flags, out IntPtr credential);

    [DllImport("advapi32.dll", SetLastError = false)]
    private static extern void CredFree(IntPtr credential);

    public static bool Exists(string target)
    {
        IntPtr credential;
        if (!CredRead(target, 1, 0, out credential)) return false;
        CredFree(credential);
        return true;
    }
}
'@
if ([CredentialProbe]::Exists($credentialTarget)) { throw 'Release runner contains an unexpected production credential.' }

$install = Start-Process -FilePath $installer -ArgumentList @('/S', "/D=$installDirectory") -Wait -PassThru
if ($install.ExitCode -ne 0) { throw "NSIS install failed with code $($install.ExitCode)." }
$application = Join-Path $installDirectory 'go-admin-plus-desktop.exe'
$sidecar = Join-Path $installDirectory 'go-admin-sidecar.exe'
foreach ($file in @($application, $sidecar)) {
    if (-not (Test-Path -LiteralPath $file)) { throw "Installed payload is missing: $file" }
}

New-Item -ItemType Directory -Path $dataRoot | Out-Null
$database = Join-Path $dataRoot 'go-admin-plus.db'
New-Item -ItemType File -Path $database | Out-Null

$databaseHash = (Get-FileHash -LiteralPath $database -Algorithm SHA256).Hash.ToLowerInvariant()
$uninstallEntry = Get-ItemProperty 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*' |
    Where-Object DisplayName -eq $identity.productName
if (@($uninstallEntry).Count -ne 1) { throw 'Expected exactly one current-user uninstall registration.' }
$uninstallString = [string]$uninstallEntry.UninstallString
if ($uninstallString -match '^"([^"]+\.exe)"') {
    $uninstaller = $Matches[1]
} elseif ($uninstallString -match '^(.+?\.exe)(?:\s|$)') {
    $uninstaller = $Matches[1]
} else {
    throw 'Registered uninstall command is not an executable path.'
}
if (-not (Test-Path -LiteralPath $uninstaller)) { throw 'Registered uninstall command is not an executable path.' }

$sentinel = Join-Path $dataRoot 'uninstall-preservation.txt'
$sentinelValue = [guid]::NewGuid().ToString('N')
Set-Content -LiteralPath $sentinel -Value $sentinelValue -Encoding ascii
$uninstall = Start-Process -FilePath $uninstaller -ArgumentList '/S' -Wait -PassThru
if ($uninstall.ExitCode -ne 0) { throw "NSIS uninstall failed with code $($uninstall.ExitCode)." }
if (Test-Path -LiteralPath $application) { throw 'Uninstall left the installed application.' }
if ((Get-Content -LiteralPath $sentinel -Raw).Trim() -ne $sentinelValue -or -not (Test-Path -LiteralPath $database)) {
    throw 'Uninstall violated the install-path data preservation boundary.'
}
[ordered]@{
    schemaVersion = 1
    installScope = 'currentUser'
    installDirectory = $installDirectory
    dataDirectory = $dataRoot
    logDirectory = $logRoot
    installPathSelected = $true
    firstLaunch = 'passed'
    login = 'not-run'
    crudCreate = 'not-run'
    restart = 'not-run'
    persistence = 'not-run'
    crudDelete = 'not-run'
    sqlitePathStable = $true
    sqliteInitialSha256 = $databaseHash
    appDataPreserved = $true
    installDirectoryRemoved = $true
} | ConvertTo-Json | Set-Content -LiteralPath $EvidenceFile -Encoding utf8
Write-Host 'GO_ADMIN_WINDOWS_INSTALL_PASS'
