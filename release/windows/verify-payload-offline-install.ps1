[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$PayloadFile
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$payload = (Resolve-Path -LiteralPath $PayloadFile).Path
$ruleName = "GoAdminWebView2Offline-$([guid]::NewGuid().ToString('N'))"
try {
    New-NetFirewallRule -DisplayName $ruleName -Direction Outbound -Action Block -Program $payload | Out-Null
    $process = Start-Process -FilePath $payload -ArgumentList @('/silent', '/install') -Wait -PassThru
    if ($process.ExitCode -notin @(0, 3010)) {
        throw "Offline WebView2 installer exited with code $($process.ExitCode)."
    }
} finally {
    Remove-NetFirewallRule -DisplayName $ruleName -ErrorAction SilentlyContinue
}

$clientId = '{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}'
$registryPaths = @(
    "HKLM:\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\$clientId",
    "HKLM:\SOFTWARE\Microsoft\EdgeUpdate\Clients\$clientId",
    "HKCU:\Software\Microsoft\EdgeUpdate\Clients\$clientId"
)
$versions = @($registryPaths | ForEach-Object {
    if (Test-Path -LiteralPath $_) {
        (Get-ItemProperty -LiteralPath $_ -Name pv -ErrorAction SilentlyContinue).pv
    }
} | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
if ($versions.Count -eq 0) {
    throw 'The offline payload completed but no WebView2 runtime registration was found.'
}

Write-Host "GO_ADMIN_WINDOWS_WEBVIEW2_OFFLINE_INSTALL_PASS version=$($versions[0])"
