# Product release contract

This directory owns product-level convergence. Platform workflows remain the
builders and verifiers for Linux AMD64 Compose, macOS ARM64 desktop, and Windows
AMD64 desktop artifacts. `product-release.mjs` binds successful platform runs
to one monorepo/OpenAPI/migration provenance record. The root, backend, and
frontend provenance fields intentionally carry the same monorepo commit SHA.

The generated manifest is a release candidate, not a publish instruction. Its
policy always records `external_publish_authorized: false`, and the workflow
only uploads the manifest as a short-lived GitHub Actions artifact.

Run the local checks with:

```sh
node --test release/manifest/product-release.test.mjs
node release/manifest/scan-compatibility.mjs
node release/manifest/product-release.mjs preflight --version 0.1.0
```

Use `task release:dispatch VERSION=0.1.0` to start the three platform gates at
the exact current monorepo commit. After all three succeed, pass their run and artifact
IDs to `task release:collect`. No command in this contract creates a GitHub
Release, deploys Compose, changes global platform security, or signs binaries.
