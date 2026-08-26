# Install Go Admin Plus on macOS

This ARM64 package is for self-use. It has an ad-hoc integrity signature, but it
is not signed with Apple Developer ID and has not been notarized by Apple.
macOS cannot verify its publisher or confirm that Apple checked it for malware.

Only continue when you obtained the artifact from the expected release source.
From the directory containing `SHA256SUMS`, verify every distributed file first:

```sh
shasum -a 256 -c SHA256SUMS
```

Do not continue if any checksum fails. Open the DMG and drag `Go Admin Plus.app`
to Applications. Try to open the app once. If macOS blocks it, open System
Settings, select Privacy & Security, scroll to Security, choose Open Anyway for
Go Admin Plus, then confirm Open. macOS saves an exception for this app.

If Open Anyway is unavailable and you have independently verified all checksums,
the scoped command below removes quarantine only from this installed app:

```sh
xattr -dr com.apple.quarantine "/Applications/Go Admin Plus.app"
```

Never use `spctl --master-disable` or disable Gatekeeper globally. A managed Mac
may prevent local exceptions; this self-use package does not bypass management
policy. Application data is stored outside the app bundle, so replacing the app
during an upgrade preserves the existing data and migration backups.
