# Install Go Admin Plus on Windows

This Windows 10/11 x64 package is for self-use. The Go Admin Plus installer and
application are not Authenticode signed. Windows cannot verify their publisher,
and Microsoft Defender SmartScreen can show "Windows protected your PC".

Only continue when the artifact came from the expected release source. In
PowerShell, calculate the installer hash from the directory containing
`SHA256SUMS`:

```powershell
Get-FileHash -Algorithm SHA256 .\go-admin-plus-0.1.0-windows-amd64-unsigned-self-use-setup.exe
```

Compare the complete value with the matching line in `SHA256SUMS`. Stop if it
does not match. If SmartScreen offers a self-use exception, select **More info**
and then **Run anyway**. The installer contains the complete Microsoft WebView2
Evergreen Standalone x64 runtime and does not need an internet connection.

Do not disable Microsoft Defender, SmartScreen or Smart App Control. Smart App
Control and managed-device policy can block an unsigned installer without a
Run anyway option; that environment is unsupported by this Phase 1 package.
Application data remains in `%LOCALAPPDATA%\go-admin-plus` when the application
is upgraded or uninstalled.
