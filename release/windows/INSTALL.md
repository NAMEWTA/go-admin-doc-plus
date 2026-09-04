# Install Go Admin Plus on Windows

Use the x64 NSIS installer only on Windows 10 version 2004 or newer. Before installation,
verify its Authenticode status and compare its SHA-256 digest with `SHA256SUMS`:

```powershell
Get-AuthenticodeSignature .\go-admin-plus-0.0.1-windows-x64-setup.exe
Get-FileHash -Algorithm SHA256 .\go-admin-plus-0.0.1-windows-x64-setup.exe
```

Stop if the signature is not `Valid`, the signer is unexpected, the timestamp is absent, or the
hash differs. Do not bypass SmartScreen, Defender, or managed-device policy.

The installer is current-user scoped and includes the WebView2 offline installer. Uninstalling
removes product files but intentionally preserves application data and the local credential so a
later reinstall can recover the same SQLite state.
