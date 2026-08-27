[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)] [string] $OutputDirectory
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
if (-not $IsWindows -or $env:PROCESSOR_ARCHITECTURE -ne 'AMD64') {
    throw 'The Windows Generator runtime requires an AMD64 Windows host.'
}
if (Test-Path -LiteralPath $OutputDirectory) {
    throw 'Generator runtime output already exists.'
}

$repository = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
$identity = Get-Content -LiteralPath (Join-Path $repository 'release/windows/identity.json') -Raw | ConvertFrom-Json
$runtime = New-Item -ItemType Directory -Path $OutputDirectory
$staging = New-Item -ItemType Directory -Path (Join-Path $env:RUNNER_TEMP "go-admin-windows-runtime-$([guid]::NewGuid().ToString('N'))")

function Get-Archive {
    param([pscustomobject] $Archive, [string] $Url)
    $target = Join-Path $staging.FullName $Archive.name
    curl.exe --fail --location --proto '=https' --tlsv1.2 --retry 3 --output $target $Url
    if ($LASTEXITCODE -ne 0) { throw "Archive download failed: $($Archive.name)" }
    $actual = (Get-FileHash -LiteralPath $target -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actual -ne $Archive.sha256) { throw "Archive checksum mismatch: $($Archive.name)" }
    return $target
}

try {
    $toolchain = New-Item -ItemType Directory -Path (Join-Path $runtime.FullName 'toolchains/windows-amd64')
    New-Item -ItemType Directory -Path (Join-Path $toolchain.FullName 'bin') | Out-Null
    New-Item -ItemType Directory -Path (Join-Path $runtime.FullName 'repository') | Out-Null
    New-Item -ItemType Directory -Path (Join-Path $runtime.FullName 'go-mod') | Out-Null
    New-Item -ItemType Directory -Path (Join-Path $runtime.FullName 'pnpm-store') | Out-Null

    $archives = $identity.generatorRuntime.archives
    $goArchive = Get-Archive $archives.goWindowsAmd64 "https://go.dev/dl/$($archives.goWindowsAmd64.name)"
    $nodeArchive = Get-Archive $archives.nodeWindowsX64 "https://nodejs.org/dist/v$($identity.generatorRuntime.nodeVersion)/$($archives.nodeWindowsX64.name)"
    $gitArchive = Get-Archive $archives.gitWindowsX64 "https://github.com/git-for-windows/git/releases/download/v2.55.0.windows.5/$($archives.gitWindowsX64.name)"
    $pnpmArchive = Get-Archive $archives.pnpm "https://registry.npmjs.org/pnpm/-/$($archives.pnpm.name)"

    Expand-Archive -LiteralPath $goArchive -DestinationPath $toolchain.FullName
    Expand-Archive -LiteralPath $nodeArchive -DestinationPath $toolchain.FullName
    Move-Item -LiteralPath (Join-Path $toolchain.FullName "node-v$($identity.generatorRuntime.nodeVersion)-win-x64") -Destination (Join-Path $toolchain.FullName 'node')
    Expand-Archive -LiteralPath $gitArchive -DestinationPath (Join-Path $toolchain.FullName 'git')
    $pnpmExtract = New-Item -ItemType Directory -Path (Join-Path $staging.FullName 'pnpm')
    tar.exe -xzf $pnpmArchive -C $pnpmExtract.FullName
    if ($LASTEXITCODE -ne 0) { throw 'pnpm archive extraction failed.' }
    Move-Item -LiteralPath (Join-Path $pnpmExtract.FullName 'package') -Destination (Join-Path $runtime.FullName 'pnpm')
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'pnpm-offline.cmd') -Destination (Join-Path $toolchain.FullName 'bin/pnpm.cmd')

    $archive = Join-Path $staging.FullName 'repository.tar'
    $tracked = @(
        '.gitignore', 'contracts/openapi', 'scripts/contracts', 'go-admin-plus',
        'go-admin-plus-ui/.npmrc', 'go-admin-plus-ui/package.json', 'go-admin-plus-ui/pnpm-lock.yaml',
        'go-admin-plus-ui/pnpm-workspace.yaml', 'go-admin-plus-ui/apps', 'go-admin-plus-ui/packages',
        'go-admin-plus-ui/tests/shell'
    )
    git -C $repository archive --format=tar --output=$archive HEAD -- @tracked
    if ($LASTEXITCODE -ne 0) { throw 'Tracked Generator repository archive failed.' }
    tar.exe -xf $archive -C (Join-Path $runtime.FullName 'repository')
    if ($LASTEXITCODE -ne 0) { throw 'Tracked Generator repository extraction failed.' }
    $generatorRepository = Join-Path $runtime.FullName 'repository'
    git -C $generatorRepository init --quiet
    git -C $generatorRepository add -f .gitignore contracts scripts go-admin-plus go-admin-plus-ui
    if ($LASTEXITCODE -ne 0) { throw 'Tracked Generator repository initialization failed.' }

    $go = Join-Path $toolchain.FullName 'go/bin/go.exe'
    $node = Join-Path $toolchain.FullName 'node/node.exe'
    $pnpm = Join-Path $runtime.FullName 'pnpm/bin/pnpm.cjs'
    $workspace = Join-Path $generatorRepository 'go-admin-plus-ui'
    $workspaceConfig = Join-Path $workspace 'pnpm-workspace.yaml'
    $workspaceBackup = Join-Path $staging.FullName 'pnpm-workspace.yaml'
    Copy-Item -LiteralPath $workspaceConfig -Destination $workspaceBackup
    & $node $pnpm --dir $workspace config set --location=project --json supportedArchitectures '{"os":["win32"],"cpu":["x64"]}'
    if ($LASTEXITCODE -ne 0) { throw 'pnpm architecture policy failed.' }
    $env:GOENV = 'off'
    $env:GOTOOLCHAIN = 'local'
    $env:GOPROXY = 'https://proxy.golang.org'
    $env:GOSUMDB = 'sum.golang.org'
    $env:GOMODCACHE = Join-Path $runtime.FullName 'go-mod'
    & $go -C (Join-Path $generatorRepository 'go-admin-plus') mod download
    if ($LASTEXITCODE -ne 0) { throw 'Go module cache preparation failed.' }
    & $node $pnpm --store-dir (Join-Path $runtime.FullName 'pnpm-store') --dir $workspace fetch --frozen-lockfile
    if ($LASTEXITCODE -ne 0) { throw 'pnpm store preparation failed.' }
    & $node $pnpm --store-dir (Join-Path $runtime.FullName 'pnpm-store') --offline --dir $workspace install --frozen-lockfile --ignore-scripts
    if ($LASTEXITCODE -ne 0) { throw 'pnpm offline workspace installation failed.' }
    Copy-Item -LiteralPath $workspaceBackup -Destination $workspaceConfig -Force

    if ((git -C $generatorRepository remote) -or (git -C $generatorRepository diff -- .)) {
        throw 'Packaged Generator repository is not a clean remote-free skeleton.'
    }
    foreach ($required in @(
        'toolchains/windows-amd64/go/bin/go.exe', 'toolchains/windows-amd64/node/node.exe',
        'toolchains/windows-amd64/git/cmd/git.exe', 'toolchains/windows-amd64/bin/pnpm.cmd',
        'repository/.git', 'repository/scripts/contracts/cli.mjs', 'go-mod', 'pnpm-store'
    )) {
        if (-not (Test-Path -LiteralPath (Join-Path $runtime.FullName $required))) {
            throw "Packaged Generator runtime is incomplete: $required"
        }
    }
    Write-Host 'GO_ADMIN_WINDOWS_GENERATOR_RUNTIME_PASS'
} finally {
    Remove-Item -LiteralPath $staging.FullName -Recurse -Force -ErrorAction SilentlyContinue
}
