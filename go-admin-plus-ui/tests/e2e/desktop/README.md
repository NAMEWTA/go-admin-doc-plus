# Native desktop E2E

The harness starts the real Tauri host and supervised SQLite sidecar. On macOS, grant the terminal Accessibility permission, then run:

```sh
GO_ADMIN_DESKTOP_NATIVE_E2E=1 node go-admin-plus-ui/tests/e2e/desktop/run.mjs
```

Without the opt-in environment variable it compiles to a deterministic skip. Source worktrees do not claim this required native check as passed.
