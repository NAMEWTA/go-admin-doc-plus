[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$IdentityFile,

    [Parameter(Mandatory = $true)]
    [string]$PayloadFile,

    [Parameter(Mandatory = $true)]
    [string]$MetadataFile
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$identity = Get-Content -LiteralPath $IdentityFile -Raw | ConvertFrom-Json
$payload = Get-Item -LiteralPath $PayloadFile
$hash = (Get-FileHash -LiteralPath $payload.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
if ($hash -ne $identity.webview2.sha256) {
    throw "WebView2 SHA-256 $hash does not match the pinned identity."
}
if ($payload.Length -ne $identity.webview2.size_bytes) {
    throw "WebView2 payload size $($payload.Length) does not match the pinned identity."
}

$signature = Get-AuthenticodeSignature -LiteralPath $payload.FullName
if ($signature.Status -ne 'Valid' -or $null -eq $signature.SignerCertificate) {
    throw "WebView2 Authenticode status is $($signature.Status), expected Valid."
}
if ($signature.SignerCertificate.Subject -notmatch "O=$([regex]::Escape($identity.webview2.signer_organization))(,|$)") {
    throw "Unexpected WebView2 signer: $($signature.SignerCertificate.Subject)"
}

$version = $payload.VersionInfo.ProductVersion
if ([string]::IsNullOrWhiteSpace($version)) {
    throw 'WebView2 payload has no ProductVersion metadata.'
}

[ordered]@{
    deployment = $identity.webview2.deployment
    architecture = $identity.webview2.architecture
    source_url = $identity.webview2.source_url
    filename = $payload.Name
    product_version = $version
    size_bytes = $payload.Length
    sha256 = $hash
    authenticode_status = $signature.Status.ToString()
    signer_subject = $signature.SignerCertificate.Subject
    signer_thumbprint = $signature.SignerCertificate.Thumbprint
} | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $MetadataFile -Encoding utf8NoBOM

Write-Host "GO_ADMIN_WINDOWS_WEBVIEW2_PAYLOAD_PASS version=$version"
