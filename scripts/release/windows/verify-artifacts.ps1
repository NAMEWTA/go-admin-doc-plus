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
    } finally { $reader.Dispose(); $stream.Dispose() }
}

Assert-X64Pe $application
$listing = (& 7z.exe l $installer.FullName 2>&1) -join "`n"
if ($LASTEXITCODE -ne 0 -or $listing -notmatch 'go-admin-plus-desktop\.exe' -or
    $listing -notmatch 'go-admin-sidecar\.exe') {
    throw 'NSIS installer does not contain the x64 app and sidecar.'
}
Write-Host 'GO_ADMIN_WINDOWS_ARTIFACT_VERIFY_PASS'
