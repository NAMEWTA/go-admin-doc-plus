# Product release manifest

`product-release.mjs` validates the immutable source contract used by the automatic tag workflow.
It binds the product version, source SHA, OpenAPI digest, migration version, supported platform
targets, and the explicit unsigned self-use policy.

```bash
node release/manifest/product-release.mjs preflight --version 0.0.2 --root-ref "$(git rev-parse HEAD)"
node --test release/manifest/product-release.test.mjs
```

The manifest does not sign, notarize, publish, or deploy anything. GitHub Actions creates the public
Release after the Linux service, macOS ARM64, and Windows x64 jobs have all completed successfully.
