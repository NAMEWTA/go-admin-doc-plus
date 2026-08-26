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

Add-Type -TypeDefinition @'
using System;
using System.ComponentModel;
using System.Runtime.InteropServices;

public static class VersionResource
{
    [DllImport("version.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern uint GetFileVersionInfoSize(string filename, out uint handle);

    [DllImport("version.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern bool GetFileVersionInfo(string filename, uint handle, uint length, byte[] data);

    [DllImport("version.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern bool VerQueryValue(IntPtr data, string subBlock, out IntPtr value, out uint length);

    public static string GetString(string filename, string key)
    {
        uint handle;
        uint size = GetFileVersionInfoSize(filename, out handle);
        if (size == 0)
            throw new Win32Exception(Marshal.GetLastWin32Error(), "The PE file has no readable version resource.");

        byte[] data = new byte[size];
        if (!GetFileVersionInfo(filename, handle, size, data))
            throw new Win32Exception(Marshal.GetLastWin32Error(), "The PE version resource could not be loaded.");

        GCHandle pinnedData = GCHandle.Alloc(data, GCHandleType.Pinned);
        try
        {
            IntPtr translation;
            uint translationLength;
            if (!VerQueryValue(pinnedData.AddrOfPinnedObject(), @"\VarFileInfo\Translation", out translation, out translationLength) || translationLength < 4)
                throw new InvalidOperationException("The PE version resource has no language/code-page translation.");

            ushort language = unchecked((ushort)Marshal.ReadInt16(translation, 0));
            ushort codePage = unchecked((ushort)Marshal.ReadInt16(translation, 2));
            string query = string.Format(@"\StringFileInfo\{0:X4}{1:X4}\{2}", language, codePage, key);
            IntPtr value;
            uint valueLength;
            if (!VerQueryValue(pinnedData.AddrOfPinnedObject(), query, out value, out valueLength) || valueLength == 0)
                throw new InvalidOperationException("The PE version resource does not contain " + key + " for " + language.ToString("X4") + codePage.ToString("X4") + ".");

            return Marshal.PtrToStringUni(value, checked((int)valueLength)).TrimEnd('\0');
        }
        finally
        {
            pinnedData.Free();
        }
    }
}
'@

$identity = Get-Content -LiteralPath $IdentityFile -Raw | ConvertFrom-Json
$installer = Get-Item -LiteralPath $InstallerFile
$application = Get-Item -LiteralPath $ApplicationFile
$expectedInstaller = "$($identity.artifact_basename)-$($identity.product_version)-windows-amd64-$($identity.release_class)-setup.exe"
if ($installer.Name -ne $expectedInstaller) {
    throw "Installer name $($installer.Name) does not match $expectedInstaller."
}
if ($installer.Length -le $identity.webview2.size_bytes) {
    throw 'Installer is too small to contain the pinned standalone WebView2 payload and application.'
}

foreach ($file in @($installer, $application)) {
    $signature = Get-AuthenticodeSignature -LiteralPath $file.FullName
    if ($signature.Status -ne 'NotSigned') {
        throw "$($file.Name) Authenticode status is $($signature.Status), expected NotSigned."
    }
}
$applicationProductName = [VersionResource]::GetString($application.FullName, 'ProductName')
$applicationProductVersion = [VersionResource]::GetString($application.FullName, 'ProductVersion')
$windowsProductVersion = "$($identity.product_version).0"
Write-Host "Application native ProductName=$applicationProductName ProductVersion=$applicationProductVersion"
Write-Host "Application .NET ProductName=$($application.VersionInfo.ProductName) ProductVersion=$($application.VersionInfo.ProductVersion)"
if ($applicationProductName -ne $identity.product_name) {
    throw "Application product name $applicationProductName does not match $($identity.product_name)."
}
if ($applicationProductVersion -notin @($identity.product_version, $windowsProductVersion)) {
    throw "Application product version $applicationProductVersion does not match $($identity.product_version) or its Windows four-part form $windowsProductVersion."
}

$goMetadata = (& go version -m $application.FullName 2>&1) -join "`n"
if ($LASTEXITCODE -ne 0 -or $goMetadata -notmatch 'GOOS=windows' -or $goMetadata -notmatch 'GOARCH=amd64') {
    throw "Application is not a Go windows/amd64 executable:`n$goMetadata"
}
$archiveListing = (& 7z l $installer.FullName 2>&1) -join "`n"
if ($LASTEXITCODE -ne 0 -or
    $archiveListing -notmatch [regex]::Escape($identity.webview2.filename) -or
    $archiveListing -notmatch [regex]::Escape("$($identity.executable_basename).exe")) {
    throw 'NSIS archive does not contain the expected application and WebView2 payload.'
}

Write-Host 'GO_ADMIN_WINDOWS_INSTALLER_VERIFY_PASS'
