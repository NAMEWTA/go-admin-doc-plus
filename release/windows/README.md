# Windows AMD64 self-use release

`identity.json` is the versioned source for the Windows product, installer and
WebView2 payload identity. Phase 1 only permits the `unsigned-self-use` release
class. The application and NSIS installer are intentionally not Authenticode
signed and must not be described as Microsoft-trusted artifacts.

The installer embeds the complete Microsoft WebView2 Evergreen Standalone x64
payload. The workflow verifies its exact byte size, SHA-256 and Microsoft
Authenticode signature before packaging. The Wails application is built with
the `error` WebView2 strategy, so it never falls back to an online bootstrapper.

The user-scope installer preserves `%LOCALAPPDATA%\go-admin-plus` across
install, upgrade and uninstall. Read `INSTALL.md` before running the artifact.
No workflow creates a GitHub Release or publishes to an external channel.
