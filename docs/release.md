# Release Guide

The release workflow publishes only the supported product targets:

| Target | Published artifact |
| --- | --- |
| Linux | `linux/amd64` and `linux/arm64` server service archives with systemd unit and deployment docs |
| macOS | Apple Silicon (`arm64`) Tauri 2 app and DMG, unsigned private-use build |
| Windows | x64 Tauri 2 NSIS installer, unsigned private-use build |

Push an exact `major.minor.patch` tag, for example `0.0.2`, to start
[`.github/workflows/release.yml`](../.github/workflows/release.yml). The workflow validates the tag,
builds the three targets in parallel, runs platform smoke tests, and creates a published GitHub Release
with the generated notes and checksums. It uses the repository `GITHUB_TOKEN`; no signing certificates,
notarization keys, or other release secrets are required.

## Linux service deployment

Download the matching Linux service archive and read its included `SERVER-INSTALL.md`. The archive
contains the `go-admin-plus-server` binary, profile examples, and systemd units; it does not contain
Docker Compose files or images. Use the repository's `deploy/compose/` definitions for container
deployment. Choose either `server-sqlite` for a single host or `server-postgres` for PostgreSQL. Keep
the database DSN and bootstrap password in permission-restricted files or environment variables, never
in Git. Run `migrate`, `bootstrap`, `doctor`, then enable the service.

## Desktop installation

Windows NSIS setup presents the normal destination chooser and installs per user. macOS users choose a
destination by dragging the ARM64 app from the DMG. Both desktop targets store runtime state below the
selected installation directory: `data/` contains SQLite, files, backups, and the Stronghold snapshot;
`logs/` contains runtime logs. Back up those directories before replacing an installation.

The artifacts are intentionally unsigned and unnotarized for personal use. Follow the operating system's
normal confirmation flow on a machine you control; do not disable managed security controls globally.

## Local checks

The three-profile clean-room covers `server-sqlite`, `server-postgres`, and `desktop-sqlite`. For
personal releases, signing and notarization are explicitly `not-required`.

```bash
task release VERSION=0.0.2
task release:verify
```

These commands perform local preflight and policy checks only. They do not push, publish, deploy, or
migrate a production database.
