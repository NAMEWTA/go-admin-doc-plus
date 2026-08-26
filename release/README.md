# Release Engineering

`release/` owns product packaging, platform identity, manifests, SBOM and provenance checks.
The public `task release VERSION=x.y.z` command performs local preflight only. Remote workflow
dispatch is intentionally absent from the root command plane; the platform release Tickets own
their idempotent orchestration and approval contracts.

Platform-specific scripts must fail when signing, notarization, native packaging or other
required tools are unavailable. They must never print release credentials or silently replace
a protected release check with an unsigned success.

Release assets below either legacy subproject are frozen migration inputs and are tracked for
atomic deletion by `scripts/go-admin-plus/legacy-governance-t21.txt`.
