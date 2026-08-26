# Change Decision Log

## LOG-001: Phase 1 desktop distribution uses unsigned self-use artifacts

- **Status:** confirmed
- **Date:** 2026-08-25
- **Decision owner:** user
- **Deviation:** `DEV-001` / spec + release
- **Decision:** macOS ARM64 and Windows AMD64 Phase 1 desktop packages prioritize self-use installability without paid platform signing. Artifacts must be labelled `unsigned-self-use`, include SHA-256, SBOM, source provenance and trust-state metadata, and document the platform's per-app user authorization path.
- **Safety boundary:** do not prescribe global Gatekeeper, SmartScreen, Defender or Smart App Control disablement. macOS may use Privacy & Security Open Anyway, with quarantine removal limited to the verified app as fallback. Windows may use SmartScreen More info / Run anyway; Smart App Control or managed policy denial is reported as unsupported.
- **Future:** Developer ID/notarization, Authenticode and Store/public trusted distribution require a new release deviation. External publish/deploy remains separately authorized.
- **Sources:** `USER-DECISION:2026-08-25`; `RESEARCH:<Url>https://support.apple.com/guide/mac-help/open-a-mac-app-from-an-unknown-developer-mh40616/mac</Url>`; `RESEARCH:<Url>https://learn.microsoft.com/en-us/windows/apps/package-and-deploy/smartscreen-reputation</Url>`.
