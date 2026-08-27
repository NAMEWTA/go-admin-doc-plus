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
$dataRoot = Join-Path $env:APPDATA $identity.bundleIdentifier
$logRoot = Join-Path $env:LOCALAPPDATA "$($identity.bundleIdentifier)\logs"
$credentialTarget = 'desktop-session-vault.com.goadmin.plus.stronghold'
$expectedThumbprint = $env:AZURE_ARTIFACT_SIGNING_EXPECTED_THUMBPRINT
if ($expectedThumbprint -notmatch '^[0-9A-Fa-f]{40}$') { throw 'Expected signing certificate thumbprint is invalid.' }
if (Test-Path -LiteralPath $dataRoot) { throw 'Standard application data already exists on the release runner.' }

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

$install = Start-Process -FilePath $installer -ArgumentList '/S' -Wait -PassThru
if ($install.ExitCode -ne 0) { throw "NSIS install failed with code $($install.ExitCode)." }
$webViewClients = @()
foreach ($registryPath in @(
    'HKLM:\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\*',
    'HKLM:\SOFTWARE\Microsoft\EdgeUpdate\Clients\*',
    'HKCU:\SOFTWARE\Microsoft\EdgeUpdate\Clients\*'
)) {
    $webViewClients += @(Get-ItemProperty -Path $registryPath -ErrorAction SilentlyContinue |
        Where-Object { $_.name -match 'WebView2' -and $_.pv -match '^\d+\.\d+\.\d+\.\d+$' })
}
$webViewVersions = @($webViewClients | ForEach-Object { [version]$_.pv } | Sort-Object -Descending -Unique)
if ($webViewVersions.Count -eq 0) { throw 'Installed WebView2 Runtime version is unavailable.' }
$webViewVersion = $webViewVersions[0].ToString()
$driverArchive = Join-Path $env:RUNNER_TEMP "edgedriver-$webViewVersion.zip"
$driverDirectory = Join-Path $env:RUNNER_TEMP "edgedriver-$webViewVersion"
curl.exe --fail --location --proto '=https' --tlsv1.2 --retry 3 --output $driverArchive "https://msedgedriver.microsoft.com/$webViewVersion/edgedriver_win64.zip"
if ($LASTEXITCODE -ne 0) { throw 'Matching Microsoft Edge Driver download failed.' }
Expand-Archive -LiteralPath $driverArchive -DestinationPath $driverDirectory
$driverExecutable = Join-Path $driverDirectory 'msedgedriver.exe'
$driverSignature = Get-AuthenticodeSignature -LiteralPath $driverExecutable
if ($driverSignature.Status -ne 'Valid' -or $null -eq $driverSignature.SignerCertificate -or
    $driverSignature.SignerCertificate.Subject -notmatch 'Microsoft') { throw 'Microsoft Edge Driver signature is invalid.' }
$driverSha256 = (Get-FileHash -LiteralPath $driverExecutable -Algorithm SHA256).Hash.ToLowerInvariant()
$env:PATH = "$driverDirectory;$env:PATH"
$entry = Get-ItemProperty 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*' |
    Where-Object DisplayName -eq $identity.productName
if (@($entry).Count -ne 1) { throw 'Expected exactly one current-user uninstall registration.' }
$uninstallString = [string]$entry.UninstallString
if ($uninstallString -match '^"([^"]+\.exe)"') {
    $uninstaller = $Matches[1]
} elseif ($uninstallString -match '^(.+?\.exe)(?:\s|$)') {
    $uninstaller = $Matches[1]
} else {
    throw 'Registered uninstall command is not an executable path.'
}
$installDirectory = Split-Path -Parent $uninstaller
$application = Join-Path $installDirectory 'go-admin-plus-desktop.exe'
$sidecar = Join-Path $installDirectory 'go-admin-sidecar.exe'
foreach ($file in @($application, $sidecar)) {
    if (-not (Test-Path -LiteralPath $file)) { throw "Installed payload is missing: $file" }
    $signature = Get-AuthenticodeSignature -LiteralPath $file
    if ($signature.Status -ne 'Valid' -or $null -eq $signature.SignerCertificate -or
        $signature.SignerCertificate.Thumbprint -ne $expectedThumbprint -or
        $null -eq $signature.TimeStamperCertificate) { throw "Installed payload signature is invalid: $file" }
}

$fixtureRoot = New-Item -ItemType Directory -Path (Join-Path $env:RUNNER_TEMP "go-admin-windows-fixture-$([guid]::NewGuid().ToString('N'))")
go -C (Join-Path $repository 'go-admin-plus') run ./test/desktop/fixture --root $fixtureRoot.FullName --mode previous
if ($LASTEXITCODE -ne 0) { throw 'Previous product fixture preparation failed.' }
New-Item -ItemType Directory -Path $dataRoot | Out-Null
$database = Join-Path $dataRoot 'go-admin-plus.db'
Copy-Item -LiteralPath (Join-Path $fixtureRoot.FullName 'data/go-admin-plus.db') -Destination $database
$traceEvidence = Join-Path $env:RUNNER_TEMP "go-admin-windows-trace-$([guid]::NewGuid().ToString('N')).json"
node (Join-Path $PSScriptRoot 'trace-installed.mjs') --application $application --evidence $traceEvidence
if ($LASTEXITCODE -ne 0 -or -not (Test-Path -LiteralPath $traceEvidence)) { throw 'Installed login and CRUD tracer failed.' }
$trace = Get-Content -LiteralPath $traceEvidence -Raw | ConvertFrom-Json
if ($trace.firstLaunchLogin -ne 'passed' -or $trace.create -ne 'passed' -or $trace.restart -ne 'passed' -or
    $trace.persistence -ne 'passed' -or $trace.delete -ne 'passed') { throw 'Installed tracer evidence is incomplete.' }
$database = (Resolve-Path -LiteralPath $database).Path
$databaseHash = (Get-FileHash -LiteralPath $database -Algorithm SHA256).Hash.ToLowerInvariant()
if (-not (Test-Path -LiteralPath (Join-Path $dataRoot 'session.stronghold'))) { throw 'Stronghold snapshot was not created.' }
if (-not [CredentialProbe]::Exists($credentialTarget)) { throw 'Windows Credential Manager key was not created.' }

$sentinel = Join-Path $dataRoot 'uninstall-preservation.txt'
$sentinelValue = [guid]::NewGuid().ToString('N')
Set-Content -LiteralPath $sentinel -Value $sentinelValue -Encoding ascii
$uninstall = Start-Process -FilePath $uninstaller -ArgumentList '/S' -Wait -PassThru
if ($uninstall.ExitCode -ne 0) { throw "NSIS uninstall failed with code $($uninstall.ExitCode)." }
if (Test-Path -LiteralPath $installDirectory) { throw 'Uninstall left the product installation directory.' }
if ((Get-Content -LiteralPath $sentinel -Raw).Trim() -ne $sentinelValue -or -not (Test-Path -LiteralPath $database)) {
    throw 'Uninstall violated the application-data preservation boundary.'
}
if (-not [CredentialProbe]::Exists($credentialTarget)) { throw 'Uninstall removed the Windows Credential Manager key.' }

[ordered]@{
    schemaVersion = 1
    installScope = 'currentUser'
    applicationSigned = $true
    sidecarSigned = $true
    firstLaunch = 'passed'
    login = $trace.firstLaunchLogin
    crudCreate = $trace.create
    restart = $trace.restart
    persistence = $trace.persistence
    crudDelete = $trace.delete
    sqlitePathStable = $true
    sqliteInitialSha256 = $databaseHash
    webView2RuntimeVersion = $webViewVersion
    edgeDriverSha256 = $driverSha256
    strongholdCreated = $true
    credentialTarget = $credentialTarget
    credentialPreserved = $true
    installDirectoryRemoved = $true
    appDataPreserved = $true
    logDirectory = $logRoot
} | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $EvidenceFile -Encoding utf8NoBOM
Write-Host 'GO_ADMIN_WINDOWS_INSTALL_VERIFY_PASS'
