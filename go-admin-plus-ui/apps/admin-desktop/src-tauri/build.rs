fn main() {
    let manifest = tauri_build::AppManifest::new().commands(&[
        "desktop_request",
        "desktop_identity",
        "desktop_navigation",
        "desktop_login",
        "desktop_logout",
        "desktop_pick_file",
        "desktop_save_file",
        "desktop_notify",
        "desktop_write_clipboard",
    ]);
    tauri_build::try_build(tauri_build::Attributes::new().app_manifest(manifest))
        .expect("failed to build Tauri application manifest");
}
