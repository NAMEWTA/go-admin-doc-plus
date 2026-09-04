# Install Go Admin Plus on macOS

Download the Apple Silicon (ARM64) DMG. From the directory containing `SHA256SUMS`, verify every
release file first:

```bash
shasum -a 256 -c SHA256SUMS
```

Stop on any checksum failure. Open the DMG and drag `Go Admin Plus.app` to the directory you choose.
The app is intentionally unsigned for private use, so macOS may require you to approve the first
launch in System Settings. Runtime data is stored inside the selected app bundle under `data/`, and
logs under `logs/`; back up both directories before replacing the app.
