[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)] [string] $Version,
    [Parameter(Mandatory = $true)] [string] $SourceSha,
    [Parameter(Mandatory = $true)] [string] $InstallerFile,
    [Parameter(Mandatory = $true)] [string] $ApplicationFile,
    [Parameter(Mandatory = $true)] [string] $InstallEvidenceFile,
    [Parameter(Mandatory = $true)] [string] $OutputDirectory
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
if ($SourceSha -notmatch '^[0-9a-f]{40}$' -or (Test-Path -LiteralPath $OutputDirectory)) { throw 'Invalid artifact emission request.' }
$repository = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
$identity = Get-Content -LiteralPath (Join-Path $repository 'release/windows/identity.json') -Raw | ConvertFrom-Json
$output = New-Item -ItemType Directory -Path $OutputDirectory
$sbom = New-Item -ItemType Directory -Path (Join-Path $output.FullName 'sbom')
Copy-Item -LiteralPath $InstallerFile -Destination $output.FullName
Copy-Item -LiteralPath $ApplicationFile -Destination $output.FullName
Copy-Item -LiteralPath $InstallEvidenceFile -Destination (Join-Path $output.FullName 'install-evidence.json')
Copy-Item -LiteralPath (Join-Path $repository 'release/windows/identity.json') -Destination $output.FullName
Copy-Item -LiteralPath (Join-Path $repository 'release/windows/INSTALL.md') -Destination $output.FullName
syft.exe "file:$ApplicationFile" -o "spdx-json=$(Join-Path $sbom.FullName 'go-admin-plus-windows-x64.spdx.json')"
if ($LASTEXITCODE -ne 0) { throw 'Windows SBOM generation failed.' }
[ordered]@{
    schemaVersion = 1
    product = 'go-admin-plus'
    version = $Version
    sourceSha = $SourceSha
    platform = 'windows-x64'
    targetTriple = $identity.targetTriple
    releaseClass = 'private-release'
    installer = 'nsis'
    signed = $false
    timestamped = $false
    remotePublished = $false
    sbomFormat = 'SPDX JSON'
} | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath (Join-Path $output.FullName 'provenance.json') -Encoding utf8NoBOM

Get-ChildItem -LiteralPath $output.FullName -File -Recurse | Where-Object Name -ne 'SHA256SUMS' |
    Sort-Object FullName | ForEach-Object {
        $relative = [System.IO.Path]::GetRelativePath($output.FullName, $_.FullName).Replace('\', '/')
        $hash = (Get-FileHash -LiteralPath $_.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
        "$hash  ./$relative"
    } | Set-Content -LiteralPath (Join-Path $output.FullName 'SHA256SUMS') -Encoding ascii
Write-Host 'GO_ADMIN_WINDOWS_SUPPLY_CHAIN_PASS'
