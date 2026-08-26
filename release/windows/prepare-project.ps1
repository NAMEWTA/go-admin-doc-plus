[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$IdentityFile,

    [Parameter(Mandatory = $true)]
    [string]$ProjectFile,

    [Parameter(Mandatory = $true)]
    [string]$Version
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$identity = Get-Content -LiteralPath $IdentityFile -Raw | ConvertFrom-Json
if ($identity.schema_version -ne 1 -or $identity.release_status -ne 'approved') {
    throw 'Windows release identity is not approved.'
}
if ($identity.release_class -ne 'unsigned-self-use') {
    throw "Unsupported Windows release class: $($identity.release_class)"
}
if ($Version -notmatch '^\d+\.\d+\.\d+$' -or $Version -ne $identity.product_version) {
    throw "Release version $Version does not match identity version $($identity.product_version)."
}
if ($identity.install_scope -ne 'user') {
    throw 'Phase 1 Windows releases require user install scope.'
}
if ($identity.uninstall_key_name -ne "$($identity.company_name)$($identity.product_name)") {
    throw 'The stable uninstall key does not match the company and product identity.'
}

$project = Get-Content -LiteralPath $ProjectFile -Raw | ConvertFrom-Json
$project.name = $identity.project_name
$project.'outputfilename' = $identity.executable_basename
$project | Add-Member -NotePropertyName info -NotePropertyValue ([pscustomobject][ordered]@{
    companyName = $identity.company_name
    productName = $identity.product_name
    productVersion = $identity.product_version
    copyright = "Copyright NAMEWTA"
    comments = 'Unsigned self-use offline desktop release'
}) -Force
$project | ConvertTo-Json -Depth 20 | Set-Content -LiteralPath $ProjectFile -Encoding utf8NoBOM

$verified = Get-Content -LiteralPath $ProjectFile -Raw | ConvertFrom-Json
if ($verified.outputfilename -ne $identity.executable_basename -or
    $verified.info.productName -ne $identity.product_name -or
    $verified.info.productVersion -ne $identity.product_version) {
    throw 'Prepared Wails project metadata does not match the release identity.'
}

Write-Host 'GO_ADMIN_WINDOWS_PROJECT_PREPARE_PASS'
