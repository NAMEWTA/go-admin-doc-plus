# macOS ARM64 release

`identity.json` is the versioned source for the macOS product name, bundle ID,
minimum OS and application-data identity. `self_use_release_status=approved`
permits the Phase 1 `unsigned-self-use` artifact. It is ad-hoc code signed for
bundle integrity, but it is neither Developer ID signed nor notarized by Apple.
`identity_status=candidate` remains the independent future signing gate; a
signed production run fails until that identity is audited as `approved`.

The backend `release-macos.yml` workflow checks out exact root, backend and
frontend commits. It builds and runs a tagged native tracer twice against one
data root, then rebuilds the untagged production app. The default self-use mode
ad-hoc signs the app, verifies a scoped quarantine removal flow, and packages
the DMG with `INSTALL.md`, provenance, checksum and SPDX SBOM. The optional
future production mode retains the Developer ID/notary/staple implementation
without making those credentials a Phase 1 requirement.

Future signed-production secrets:

- `APPLE_DEVELOPER_ID_P12_BASE64`
- `APPLE_DEVELOPER_ID_P12_PASSWORD`
- `APPLE_SIGNING_IDENTITY`
- `APPLE_NOTARY_KEY_P8_BASE64`
- `APPLE_NOTARY_KEY_ID`
- `APPLE_NOTARY_ISSUER_ID`

Read `INSTALL.md` before running the self-use artifact. It requires verifying
SHA-256 first and never recommends globally disabling Gatekeeper. No workflow
publishes a GitHub Release or external distribution.
