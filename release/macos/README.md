# macOS Apple Silicon Release

`identity.json` is the canonical bundle identity, minimum system version, target architecture, and
artifact evidence contract. Release artifacts target the latest Apple Silicon (`aarch64-apple-darwin`)
only; Intel macOS builds are not published.

The app and DMG are intentionally unsigned and not notarized for private self-use. macOS may show
an unsigned-app warning; on a machine you control, allow the app through the normal System Settings
security flow.

Drag `Go Admin Plus.app` from the DMG to a directory you choose. Runtime data is stored at
`<install-directory>/Go Admin Plus.app/data` and logs at `<install-directory>/Go Admin Plus.app/logs`.
Back up those directories before replacing the application bundle.
