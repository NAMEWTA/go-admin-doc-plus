#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod demo_contract;
mod proxy;
mod vault;

use std::{
    fs,
    path::{Path, PathBuf},
    sync::{
        Arc, Mutex, RwLock,
        atomic::{AtomicBool, Ordering},
    },
    time::{Duration, Instant},
};

use base64::{Engine as _, engine::general_purpose::URL_SAFE_NO_PAD};
use proxy::{
    DesktopRequest, DesktopResponse, IdentityResult, LogoutResult, PublicMenu, PublicProfile,
    TransportProxy,
};
use serde::Serialize;
use tauri::{Manager, RunEvent, State};
use tauri_plugin_shell::{
    ShellExt,
    process::{CommandChild, CommandEvent},
};
use zeroize::Zeroizing;

const STARTUP_TIMEOUT: Duration = Duration::from_secs(30);
const SHUTDOWN_TIMEOUT: Duration = Duration::from_secs(5);
const MAX_DIAGNOSTIC_BYTES: usize = 4096;

struct HostState {
    proxy: RwLock<Option<Arc<TransportProxy>>>,
    child: Mutex<Option<CommandChild>>,
    child_exited: Arc<AtomicBool>,
    shutting_down: Arc<AtomicBool>,
    requests: Arc<tokio::sync::Semaphore>,
}

impl HostState {
    fn new() -> Self {
        Self {
            proxy: RwLock::new(None),
            child: Mutex::new(None),
            child_exited: Arc::new(AtomicBool::new(false)),
            shutting_down: Arc::new(AtomicBool::new(false)),
            requests: Arc::new(tokio::sync::Semaphore::new(8)),
        }
    }

    fn proxy(&self) -> Result<Arc<TransportProxy>, &'static str> {
        self.proxy
            .read()
            .map_err(|_| "desktop runtime unavailable")?
            .clone()
            .ok_or("desktop runtime unavailable")
    }

    fn shutdown(&self) {
        self.shutting_down.store(true, Ordering::Release);
        if let Ok(proxy) = self.proxy() {
            proxy.shutdown();
        }
        let deadline = Instant::now() + SHUTDOWN_TIMEOUT;
        while !self.child_exited.load(Ordering::Acquire) && Instant::now() < deadline {
            std::thread::sleep(Duration::from_millis(25));
        }
        if !self.child_exited.load(Ordering::Acquire)
            && let Ok(mut owner) = self.child.lock()
            && let Some(child) = owner.take()
        {
            let _ = child.kill();
        }
        if let Ok(mut owner) = self.child.lock() {
            owner.take();
        }
        if let Ok(mut owner) = self.proxy.write() {
            owner.take();
        }
    }
}

#[tauri::command]
async fn desktop_request(
    state: State<'_, Arc<HostState>>,
    request: DesktopRequest,
) -> Result<DesktopResponse, &'static str> {
    let proxy = state.proxy()?;
    let permit = state
        .requests
        .clone()
        .try_acquire_owned()
        .map_err(|_| "desktop request busy")?;
    tauri::async_runtime::spawn_blocking(move || {
        let _permit = permit;
        proxy.business(request)
    })
    .await
    .map_err(|_| "desktop request failed")?
}

#[tauri::command]
async fn desktop_identity(
    state: State<'_, Arc<HostState>>,
) -> Result<IdentityResult, &'static str> {
    let proxy = state.proxy()?;
    let permit = state
        .requests
        .clone()
        .try_acquire_owned()
        .map_err(|_| "desktop request busy")?;
    tauri::async_runtime::spawn_blocking(move || {
        let _permit = permit;
        proxy.identity()
    })
    .await
    .map_err(|_| "desktop identity failed")?
}

#[tauri::command]
async fn desktop_navigation(
    state: State<'_, Arc<HostState>>,
) -> Result<Vec<PublicMenu>, &'static str> {
    let proxy = state.proxy()?;
    let permit = state
        .requests
        .clone()
        .try_acquire_owned()
        .map_err(|_| "desktop request busy")?;
    tauri::async_runtime::spawn_blocking(move || {
        let _permit = permit;
        proxy.navigation()
    })
    .await
    .map_err(|_| "desktop navigation failed")?
}

#[tauri::command]
async fn desktop_login(
    state: State<'_, Arc<HostState>>,
    username: String,
    password: String,
) -> Result<PublicProfile, &'static str> {
    let proxy = state.proxy()?;
    let permit = state
        .requests
        .clone()
        .try_acquire_owned()
        .map_err(|_| "desktop request busy")?;
    let username = Zeroizing::new(username);
    let password = Zeroizing::new(password);
    tauri::async_runtime::spawn_blocking(move || {
        let _permit = permit;
        proxy.login(username, password)
    })
    .await
    .map_err(|_| "desktop login failed")?
}

#[tauri::command]
async fn desktop_logout(state: State<'_, Arc<HostState>>) -> Result<LogoutResult, &'static str> {
    let proxy = state.proxy()?;
    let permit = state
        .requests
        .clone()
        .try_acquire_owned()
        .map_err(|_| "desktop request busy")?;
    tauri::async_runtime::spawn_blocking(move || {
        let _permit = permit;
        proxy.logout()
    })
    .await
    .map_err(|_| "desktop logout failed")?
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
struct LaunchWire<'a> {
    data_directory: &'a Path,
    log_directory: &'a Path,
    loopback_port: u16,
    readiness_nonce: &'a str,
    control_token: &'a str,
}

async fn start_runtime(app: tauri::AppHandle, state: Arc<HostState>) -> Result<(), &'static str> {
    let (data_path, log_path) = runtime_paths(&app)?;
    let data_root = prepare_root(data_path)?;
    let log_root = prepare_root(log_path)?;
    reject_overlap(&data_root, &log_root)?;
    let vault = vault::SessionVault::open(&data_root)?;
    let readiness_nonce = Zeroizing::new(random_secret()?);
    let control_token = Zeroizing::new(random_secret()?);
    let mut launch = serde_json::to_vec(&LaunchWire {
        data_directory: &data_root,
        log_directory: &log_root,
        loopback_port: 0,
        readiness_nonce: &readiness_nonce,
        control_token: &control_token,
    })
    .map_err(|_| "desktop launch material encode failed")?;
    launch.push(b'\n');

    let command = app
        .shell()
        .sidecar("go-admin-sidecar")
        .map_err(|_| "desktop sidecar command unavailable")?
        .env_clear();
    let (mut events, mut child) = command
        .spawn()
        .map_err(|_| "desktop sidecar spawn failed")?;
    child
        .write(&launch)
        .map_err(|_| "desktop sidecar input failed")?;
    launch.fill(0);

    let port = match wait_for_listening(&mut events).await {
        Ok(port) => port,
        Err(error) => {
            let _ = child.kill();
            return Err(error);
        }
    };
    let origin = format!("http://127.0.0.1:{port}");
    if let Err(error) = readiness_handshake(&origin, readiness_nonce).await {
        let _ = child.kill();
        return Err(error);
    }
    let proxy = match TransportProxy::new(origin, control_token, vault) {
        Ok(proxy) => Arc::new(proxy),
        Err(error) => {
            let _ = child.kill();
            return Err(error);
        }
    };
    *state
        .proxy
        .write()
        .map_err(|_| "desktop runtime unavailable")? = Some(proxy);
    *state
        .child
        .lock()
        .map_err(|_| "desktop runtime unavailable")? = Some(child);

    let exited = Arc::clone(&state.child_exited);
    let shutting_down = Arc::clone(&state.shutting_down);
    let monitor_state = Arc::clone(&state);
    let monitor_app = app.clone();
    tauri::async_runtime::spawn(async move {
        let mut observed = 0usize;
        let mut output_failed = false;
        while let Some(event) = events.recv().await {
            match event {
                CommandEvent::Stdout(bytes) | CommandEvent::Stderr(bytes) => {
                    observed = observed
                        .saturating_add(bytes.len())
                        .min(MAX_DIAGNOSTIC_BYTES + 1);
                    if observed > MAX_DIAGNOSTIC_BYTES {
                        output_failed = true;
                        if let Ok(mut owner) = monitor_state.child.lock()
                            && let Some(child) = owner.take()
                        {
                            let _ = child.kill();
                        }
                    }
                }
                CommandEvent::Terminated(_) => {
                    exited.store(true, Ordering::Release);
                    if !shutting_down.load(Ordering::Acquire) {
                        monitor_app.exit(1);
                    }
                    return;
                }
                _ => {}
            }
        }
        if let Ok(mut owner) = monitor_state.child.lock()
            && let Some(child) = owner.take()
        {
            let _ = child.kill();
        }
        exited.store(true, Ordering::Release);
        if output_failed || !shutting_down.load(Ordering::Acquire) {
            monitor_app.exit(1);
        }
    });

    app.get_webview_window("main")
        .ok_or("desktop main window unavailable")?
        .show()
        .map_err(|_| "desktop main window failed")
}

fn runtime_paths(app: &tauri::AppHandle) -> Result<(PathBuf, PathBuf), &'static str> {
    #[cfg(feature = "native-e2e")]
    if let Some(root) = std::env::var_os("GO_ADMIN_DESKTOP_E2E_ROOT") {
        let root = PathBuf::from(root);
        if !root.is_absolute() {
            return Err("desktop e2e root invalid");
        }
        let root = fs::canonicalize(root).map_err(|_| "desktop e2e root invalid")?;
        return Ok((root.join("data"), root.join("logs")));
    }
    Ok((
        app.path()
            .app_data_dir()
            .map_err(|_| "desktop data path unavailable")?,
        app.path()
            .app_log_dir()
            .map_err(|_| "desktop log path unavailable")?,
    ))
}

async fn wait_for_listening(
    events: &mut tokio::sync::mpsc::Receiver<CommandEvent>,
) -> Result<u16, &'static str> {
    let deadline = tokio::time::Instant::now() + STARTUP_TIMEOUT;
    let mut stdout = Vec::new();
    loop {
        let event = tokio::time::timeout_at(deadline, events.recv())
            .await
            .map_err(|_| "desktop sidecar startup timed out")?
            .ok_or("desktop sidecar stopped during startup")?;
        match event {
            CommandEvent::Stdout(bytes) => {
                if stdout.len().saturating_add(bytes.len()) > MAX_DIAGNOSTIC_BYTES {
                    return Err("desktop sidecar startup output rejected");
                }
                stdout.extend_from_slice(&bytes);
                if let Some(position) = stdout.iter().position(|byte| *byte == b'\n') {
                    #[derive(serde::Deserialize)]
                    #[serde(deny_unknown_fields)]
                    struct Listening {
                        state: String,
                        port: u16,
                    }
                    let value: Listening = serde_json::from_slice(&stdout[..position])
                        .map_err(|_| "desktop sidecar startup response invalid")?;
                    stdout.fill(0);
                    if value.state == "listening" && value.port > 0 {
                        return Ok(value.port);
                    }
                    return Err("desktop sidecar startup response invalid");
                }
            }
            CommandEvent::Stderr(bytes) if bytes.len() > MAX_DIAGNOSTIC_BYTES => {
                return Err("desktop sidecar startup output rejected");
            }
            CommandEvent::Terminated(_) => return Err("desktop sidecar stopped during startup"),
            _ => {}
        }
    }
}

async fn readiness_handshake(origin: &str, nonce: Zeroizing<String>) -> Result<(), &'static str> {
    let origin = origin.to_owned();
    tauri::async_runtime::spawn_blocking(move || {
        let deadline = Instant::now() + STARTUP_TIMEOUT;
        while Instant::now() < deadline {
            let result = ureq::get(format!("{origin}/__desktop/ready"))
                .header("X-Go-Admin-Desktop-Nonce", nonce.as_str())
                .config()
                .timeout_global(Some(Duration::from_secs(1)))
                .build()
                .call();
            if let Ok(mut response) = result
                && response.status() == 200
            {
                let value: serde_json::Value = response
                    .body_mut()
                    .read_json()
                    .map_err(|_| "desktop readiness response invalid")?;
                if value == serde_json::json!({"state":"ready"}) {
                    return Ok(());
                }
                return Err("desktop readiness response invalid");
            }
            std::thread::sleep(Duration::from_millis(50));
        }
        Err("desktop readiness timed out")
    })
    .await
    .map_err(|_| "desktop readiness failed")?
}

fn prepare_root(path: PathBuf) -> Result<PathBuf, &'static str> {
    validate_existing_ancestors(&path)?;
    fs::create_dir_all(&path).map_err(|_| "desktop runtime directory unavailable")?;
    let canonical = fs::canonicalize(&path).map_err(|_| "desktop runtime directory unavailable")?;
    let info = fs::symlink_metadata(&path).map_err(|_| "desktop runtime directory unavailable")?;
    if !info.is_dir() || info.file_type().is_symlink() || canonical != path {
        return Err("desktop runtime directory is not canonical");
    }
    set_private_directory(&path)?;
    Ok(canonical)
}

fn validate_existing_ancestors(path: &Path) -> Result<(), &'static str> {
    use std::path::Component;

    if !path.is_absolute()
        || path
            .components()
            .any(|part| matches!(part, Component::CurDir | Component::ParentDir))
    {
        return Err("desktop runtime directory path is invalid");
    }
    let mut ancestors: Vec<_> = path.ancestors().collect();
    ancestors.reverse();
    for ancestor in ancestors {
        match fs::symlink_metadata(ancestor) {
            Ok(info) if info.is_dir() && !info.file_type().is_symlink() => {}
            Ok(_) => return Err("desktop runtime directory ancestor is unsafe"),
            Err(error) if error.kind() == std::io::ErrorKind::NotFound => {}
            Err(_) => return Err("desktop runtime directory unavailable"),
        }
    }
    Ok(())
}

#[cfg(unix)]
fn set_private_directory(path: &Path) -> Result<(), &'static str> {
    use std::os::unix::fs::PermissionsExt;
    fs::set_permissions(path, fs::Permissions::from_mode(0o700))
        .map_err(|_| "desktop runtime directory cannot be secured")
}
#[cfg(not(unix))]
fn set_private_directory(_path: &Path) -> Result<(), &'static str> {
    Ok(())
}

fn reject_overlap(first: &Path, second: &Path) -> Result<(), &'static str> {
    if first == second || first.starts_with(second) || second.starts_with(first) {
        return Err("desktop runtime directories overlap");
    }
    Ok(())
}

fn random_secret() -> Result<String, &'static str> {
    let mut value = [0_u8; 32];
    getrandom::fill(&mut value).map_err(|_| "desktop launch secret generation failed")?;
    let encoded = URL_SAFE_NO_PAD.encode(value);
    value.fill(0);
    Ok(encoded)
}

fn main() {
    let state = Arc::new(HostState::new());
    let managed = Arc::clone(&state);
    let builder = tauri::Builder::default()
        .plugin(tauri_plugin_single_instance::init(|app, _, _| {
            if let Some(window) = app.get_webview_window("main") {
                let _ = window.set_focus();
            }
        }))
        .plugin(tauri_plugin_shell::init())
        .manage(managed)
        .invoke_handler(tauri::generate_handler![
            desktop_request,
            desktop_identity,
            desktop_navigation,
            desktop_login,
            desktop_logout
        ])
        .setup({
            let state = Arc::clone(&state);
            move |app| {
                let handle = app.handle().clone();
                tauri::async_runtime::spawn(async move {
                    if start_runtime(handle.clone(), state).await.is_err() {
                        handle.exit(1);
                    }
                });
                Ok(())
            }
        });
    let app = builder
        .build(tauri::generate_context!())
        .expect("desktop host initialization failed");
    app.run(move |_handle, event| {
        if let RunEvent::Exit = event {
            state.shutdown();
        }
    });
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn random_material_is_url_safe_and_independent() {
        let first = random_secret().unwrap();
        let second = random_secret().unwrap();
        assert_eq!(first.len(), 43);
        assert_ne!(first, second);
        assert!(
            first
                .bytes()
                .all(|byte| byte.is_ascii_alphanumeric() || byte == b'-' || byte == b'_')
        );
    }

    #[test]
    fn nested_runtime_roots_are_rejected() {
        assert!(reject_overlap(Path::new("/data"), Path::new("/data/logs")).is_err());
        assert!(reject_overlap(Path::new("/data"), Path::new("/logs")).is_ok());
    }

    #[cfg(unix)]
    #[test]
    fn symlinked_parent_is_rejected_before_creating_child() {
        use std::os::unix::fs::symlink;

        let unique = format!("go-admin-desktop-root-{}", random_secret().unwrap());
        let temporary = fs::canonicalize(std::env::temp_dir()).unwrap().join(unique);
        let outside = temporary.join("outside");
        let linked = temporary.join("linked");
        fs::create_dir_all(&outside).unwrap();
        symlink(&outside, &linked).unwrap();
        let requested = linked.join("missing");
        assert!(prepare_root(requested).is_err());
        assert!(!outside.join("missing").exists());
        fs::remove_dir_all(&temporary).unwrap();
    }
}
