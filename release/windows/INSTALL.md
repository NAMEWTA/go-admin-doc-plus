# Install Go Admin Plus on Windows

Download the x64 NSIS installer and verify its SHA-256 digest:

```powershell
Get-FileHash -Algorithm SHA256 .\go-admin-plus-0.0.2-windows-x64-setup.exe
```

The installer is intentionally unsigned for private self-use. Windows SmartScreen or Defender may
show a warning; continue only on a machine you control and follow the normal Windows confirmation
flow. Do not disable system security controls globally.

The NSIS setup is current-user scoped and presents a destination chooser. Runtime data is kept under
the selected installation directory: `data\\` contains SQLite, files, backups, and the Stronghold
snapshot; `logs\\` contains runtime logs. Back up those directories before replacing the app. An
uninstall removes installed binaries but leaves runtime-created data for recovery.

Release verification scripts on Windows require a POSIX shell. Install Git for Windows (Git Bash)
or MSYS2 and ensure `sh.exe` is on `PATH`, or set `GO_ADMIN_POSIX_SHELL` to the absolute path of a
compatible `sh.exe` before invoking the scripts.
