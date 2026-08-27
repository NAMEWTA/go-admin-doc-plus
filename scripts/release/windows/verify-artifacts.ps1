[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)] [string] $Version,
    [Parameter(Mandatory = $true)] [string] $InstallerFile,
    [Parameter(Mandatory = $true)] [string] $ApplicationFile
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$repository = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
$identity = Get-Content -LiteralPath (Join-Path $repository 'release/windows/identity.json') -Raw | ConvertFrom-Json
$installer = Get-Item -LiteralPath $InstallerFile
$application = Get-Item -LiteralPath $ApplicationFile
$expected = "$($identity.artifactBasename)-$Version-windows-x64-setup.exe"
if ($installer.Name -ne $expected) { throw "Unexpected installer name: $($installer.Name)" }
$expectedThumbprint = $env:AZURE_ARTIFACT_SIGNING_EXPECTED_THUMBPRINT
if ($expectedThumbprint -notmatch '^[0-9A-Fa-f]{40}$') { throw 'Expected signing certificate thumbprint is invalid.' }

function Assert-Signed {
    param([System.IO.FileInfo] $File)
    $signature = Get-AuthenticodeSignature -LiteralPath $File.FullName
    if ($signature.Status -ne 'Valid' -or $null -eq $signature.SignerCertificate -or
        $signature.SignerCertificate.Thumbprint -ne $expectedThumbprint) {
        throw "$($File.Name) does not have a valid Authenticode signature."
    }
    if ($null -eq $signature.TimeStamperCertificate) { throw "$($File.Name) is not timestamped." }
}
function Assert-X64Pe {
    param([System.IO.FileInfo] $File)
    $stream = $File.OpenRead()
    try {
        $reader = [System.IO.BinaryReader]::new($stream)
        if ($reader.ReadUInt16() -ne 0x5A4D) { throw "$($File.Name) is not a PE file." }
        $stream.Position = 0x3c
        $peOffset = $reader.ReadInt32()
        $stream.Position = $peOffset
        if ($reader.ReadUInt32() -ne 0x00004550 -or $reader.ReadUInt16() -ne 0x8664) {
            throw "$($File.Name) is not an x64 PE executable."
        }
    } finally { $stream.Dispose() }
}

Assert-Signed $installer
Assert-Signed $application
Assert-X64Pe $application
$listing = (& 7z.exe l $installer.FullName 2>&1) -join "`n"
if ($LASTEXITCODE -ne 0 -or $listing -notmatch 'go-admin-plus-desktop\.exe' -or
    $listing -notmatch 'go-admin-sidecar\.exe' -or $listing -notmatch 'generator') {
    throw 'NSIS installer does not contain the signed app, sidecar, and Generator resources.'
}
Write-Host 'GO_ADMIN_WINDOWS_ARTIFACT_VERIFY_PASS'
