# Windows x64 Release

`identity.json` is the canonical Windows release contract. The only release artifact is a
Tauri 2 x64 NSIS installer produced automatically when an exact semver tag is pushed.

The installer is intentionally unsigned for private self-use. Windows SmartScreen or Defender may
show a warning; continue only on a machine you control. The NSIS installer is current-user scoped,
shows the normal destination chooser, and embeds the WebView2 offline installer.

Choose the installation directory during setup. The application stores its SQLite database, files,
backups, logs, and desktop runtime state below `<install-directory>\\data` and
`<install-directory>\\logs`; moving or replacing the directory without backing it up can lose data.
Uninstall removes installed binaries and leaves runtime-created data for recovery. The release
workflow smoke-tests a non-default installation path, first launch, login, CRUD, restart, and data
persistence.
