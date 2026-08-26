# Release Engineering

`release/` owns product packaging, platform identity, manifests, SBOM and provenance checks.
The public `task release VERSION=x.y.z` command performs local preflight only. Remote workflow
dispatch remains an explicit internal operation and requires the GitHub CLI plus an exact
`origin/main` match.

Platform-specific scripts must fail when signing, notarization, native packaging or other
required tools are unavailable. They must never print release credentials or silently replace
a protected release check with an unsigned success.

Release assets below either legacy subproject are frozen migration inputs and are tracked for
atomic deletion by `scripts/go-admin-plus/legacy-governance-t21.txt`.
