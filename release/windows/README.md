# Windows x64 Release

`identity.json` is the canonical Windows release contract. The only release artifact is a
Tauri 2 x64 NSIS installer produced from one immutable source commit in the protected
`windows-production` environment.

The application, sidecar, and installer must have timestamped Authenticode signatures from Azure
Artifact Signing. The installer uses current-user scope, embeds the WebView2 offline installer,
and contains the complete Generator repository plus pinned Go, Node, pnpm, and MinGit runtimes.
Every payload signer must match the exact certificate thumbprint supplied by the protected environment.

Uninstall removes only installed product files. It preserves `%APPDATA%\com.goadmin.plus`, the
local log directory, and the Stronghold credential. The protected install tracer performs login,
Demo CRUD, restart/persistence, and uninstall-boundary checks before retaining release evidence.
