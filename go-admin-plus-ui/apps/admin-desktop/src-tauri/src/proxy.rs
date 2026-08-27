use std::{collections::BTreeSet, sync::Mutex, time::Duration};

use base64::{Engine as _, engine::general_purpose::STANDARD};
use serde::{Deserialize, Serialize};
use serde_json::Value;
use zeroize::{Zeroize, Zeroizing};

use crate::product_contract;
use crate::vault::{SessionSecrets, SessionVault};

const SESSION_COOKIE: &str = "__Host-go-admin-session";
const CONTROL_HEADER: &str = "X-Go-Admin-Desktop-Control";
const MAX_REQUEST_BYTES: usize = 1024 * 1024;
const MAX_RESPONSE_BYTES: u64 = 11 * 1024 * 1024;

#[derive(Clone, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct DesktopRequest {
    pub path: String,
    pub method: String,
    pub body: Option<Value>,
}

#[derive(Serialize)]
pub struct DesktopResponse {
    pub status: u16,
    pub body: Value,
}

#[derive(Default)]
struct Rotation {
    cookie: Option<String>,
    csrf: Option<String>,
}

struct WireResponse {
    response: DesktopResponse,
    rotation: Rotation,
    protected_values: Vec<Zeroizing<String>>,
}

#[derive(Clone, Deserialize, Serialize)]
#[serde(deny_unknown_fields, rename_all = "camelCase")]
pub struct PublicProfile {
    pub id: String,
    pub username: String,
    pub display_name: String,
    pub email: String,
    pub avatar_ref: Option<String>,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields, rename_all = "camelCase")]
struct SessionWire {
    profile: PublicProfile,
    csrf_token: String,
}

#[derive(Clone, Deserialize, Serialize)]
#[serde(deny_unknown_fields, rename_all = "camelCase")]
pub struct PublicMenu {
    pub key: String,
    pub label: String,
    pub path: String,
    pub permission_code: String,
    pub sort_order: i64,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields, rename_all = "camelCase")]
struct ManifestWire {
    permission_codes: Vec<String>,
    menus: Vec<PublicMenu>,
    data_scope: String,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase", tag = "kind")]
pub enum IdentityResult {
    #[serde(rename = "unauthenticated")]
    Unauthenticated,
    #[serde(rename = "authenticated")]
    Authenticated {
        profile: PublicProfile,
        permissions: Vec<String>,
        #[serde(rename = "dataScope")]
        data_scope: String,
    },
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct LogoutResult {
    pub local_cleared: bool,
    pub remote_revoked: bool,
}

pub struct TransportProxy {
    origin: String,
    control: Zeroizing<String>,
    vault: Mutex<SessionVault>,
    agent: ureq::Agent,
}

impl TransportProxy {
    pub fn new(
        origin: String,
        control: Zeroizing<String>,
        vault: SessionVault,
    ) -> Result<Self, &'static str> {
        if !valid_origin(&origin) || !valid_secret(&control) {
            return Err("desktop proxy configuration invalid");
        }
        let config = ureq::Agent::config_builder()
            .timeout_global(Some(Duration::from_secs(15)))
            .max_redirects(0)
            .http_status_as_error(false)
            .build();
        Ok(Self {
            origin,
            control,
            vault: Mutex::new(vault),
            agent: config.into(),
        })
    }

    pub fn identity(&self) -> Result<IdentityResult, &'static str> {
        let current = self.send("GET", "/iam/session/current", None, true)?;
        if current.response.status == 401 {
            if contains_secret_material(&current.response.body, &current.protected_values) {
                return Err("desktop identity response invalid");
            }
            self.commit_rotation(current.rotation, None)?;
            return Ok(IdentityResult::Unauthenticated);
        }
        if current.response.status != 200 {
            return Err("desktop identity request failed");
        }
        let session: SessionWire = serde_json::from_value(current.response.body)
            .map_err(|_| "desktop identity response invalid")?;
        validate_profile(&session.profile)?;
        if profile_contains_secret(&session.profile, &current.protected_values) {
            return Err("desktop identity response invalid");
        }
        self.commit_rotation(current.rotation, Some(session.csrf_token))?;
        let manifest = self.manifest()?;
        Ok(IdentityResult::Authenticated {
            profile: session.profile,
            permissions: manifest.permission_codes,
            data_scope: manifest.data_scope,
        })
    }

    pub fn navigation(&self) -> Result<Vec<PublicMenu>, &'static str> {
        let manifest = self.manifest()?;
        Ok(manifest.menus)
    }

    pub fn login(
        &self,
        username: Zeroizing<String>,
        password: Zeroizing<String>,
    ) -> Result<PublicProfile, &'static str> {
        if username.len() < 3 || username.len() > 64 || password.len() < 12 || password.len() > 128
        {
            return Err("desktop login input invalid");
        }
        let mut credentials =
            Some(serde_json::json!({"username": username.as_str(), "password": password.as_str()}));
        let response = self.send("POST", "/iam/session/login", credentials.as_ref(), false);
        if let Some(value) = credentials.as_mut() {
            scrub_json(value);
        }
        let response = response?;
        if response.response.status != 200 {
            return Err("desktop login rejected");
        }
        let session: SessionWire = serde_json::from_value(response.response.body)
            .map_err(|_| "desktop login response invalid")?;
        validate_profile(&session.profile)?;
        if profile_contains_secret(&session.profile, &response.protected_values) {
            return Err("desktop login response invalid");
        }
        self.commit_rotation(response.rotation, Some(session.csrf_token))?;
        Ok(session.profile)
    }

    pub fn logout(&self) -> Result<LogoutResult, &'static str> {
        let response = self.send("POST", "/iam/session/logout", None, true)?;
        if !logout_status_allows_clear(response.response.status)
            || contains_secret_material(&response.response.body, &response.protected_values)
        {
            return Err("desktop logout request failed");
        }
        finish_logout(response.response.status, || {
            self.vault
                .lock()
                .map_err(|_| "desktop vault unavailable")?
                .clear()
        })
    }

    pub fn business(&self, request: DesktopRequest) -> Result<DesktopResponse, &'static str> {
        #[cfg(feature = "native-e2e")]
        if request.path == "/__desktop/test-control" {
            return self.native_e2e_control(request);
        }
        let mut request = product_contract::validate_request(request)?;
        if request.upload.is_none()
            && request.body.as_ref().is_some_and(|body| {
                serde_json::to_vec(body).map_or(true, |encoded| encoded.len() > MAX_REQUEST_BYTES)
                    || contains_secret_material(body, &[])
            })
        {
            return Err("desktop request body rejected");
        }
        let response = self.send(&request.method, &request.path, request.body.as_ref(), true);
        if let Some(body) = request.body.as_mut() {
            scrub_json(body);
        }
        let response = response?;
        if contains_secret_material(&response.response.body, &response.protected_values) {
            return Err("desktop response body rejected");
        }
        let mut public = product_contract::validate_response(&request, response.response)?;
        self.commit_rotation(response.rotation, None)?;
        if contains_secret_material(&public.body, &[]) {
            scrub_json(&mut public.body);
            return Err("desktop response body rejected");
        }
        Ok(public)
    }

    #[cfg(feature = "native-e2e")]
    fn native_e2e_control(&self, request: DesktopRequest) -> Result<DesktopResponse, &'static str> {
        if request.method != "POST" {
            return Err("desktop test control rejected");
        }
        let body = request.body.ok_or("desktop test control rejected")?;
        let values = body
            .as_object()
            .filter(|values| values.len() == 1)
            .ok_or("desktop test control rejected")?;
        let action = values
            .get("action")
            .and_then(Value::as_str)
            .filter(|value| {
                matches!(
                    *value,
                    "scope-self"
                        | "scope-all"
                        | "permissions-off"
                        | "permissions-on"
                        | "session-revoke"
                )
            })
            .ok_or("desktop test control rejected")?;
        let body = serde_json::json!({"action": action});
        let response = self.send("POST", "/__desktop/test-control", Some(&body), false)?;
        if response.response.status != 204
            || response.response.body != Value::Null
            || response.rotation.cookie.is_some()
            || response.rotation.csrf.is_some()
            || contains_secret_material(&response.response.body, &response.protected_values)
        {
            return Err("desktop test control failed");
        }
        Ok(response.response)
    }

    pub fn shutdown(&self) {
        let _ = self.send("POST", "/__desktop/shutdown", None, false);
    }

    fn manifest(&self) -> Result<ManifestWire, &'static str> {
        let response = self.send("GET", "/iam/administration/manifest", None, true)?;
        if response.response.status != 200 {
            return Err("desktop navigation request failed");
        }
        if contains_secret_material(&response.response.body, &response.protected_values) {
            return Err("desktop navigation response invalid");
        }
        let manifest = decode_manifest(response.response.body)?;
        self.commit_rotation(response.rotation, None)?;
        Ok(manifest)
    }

    fn commit_rotation(
        &self,
        rotation: Rotation,
        response_csrf: Option<String>,
    ) -> Result<(), &'static str> {
        if rotation.cookie.is_none() && rotation.csrf.is_none() && response_csrf.is_none() {
            return Ok(());
        }
        let header_csrf = rotation.csrf;
        if header_csrf
            .as_ref()
            .is_some_and(|value| !valid_secret(value))
            || response_csrf
                .as_ref()
                .is_some_and(|value| !valid_secret(value))
            || header_csrf
                .as_ref()
                .zip(response_csrf.as_ref())
                .is_some_and(|(left, right)| left != right)
        {
            return Err("desktop session response invalid");
        }
        let vault = self.vault.lock().map_err(|_| "desktop vault unavailable")?;
        let mut current = vault.read()?.unwrap_or(SessionSecrets {
            token: String::new(),
            csrf: String::new(),
        });
        if let Some(cookie) = rotation.cookie {
            match parse_session_cookie(&cookie)? {
                Some(token) => current.token = token,
                None => return vault.clear(),
            }
        }
        if let Some(csrf) = response_csrf.or(header_csrf) {
            current.csrf.zeroize();
            current.csrf = csrf;
        }
        if !valid_secret(&current.token) || !valid_secret(&current.csrf) {
            return Err("desktop session rotation invalid");
        };
        vault.write(current)
    }

    fn send(
        &self,
        method: &str,
        path: &str,
        body: Option<&Value>,
        authenticated: bool,
    ) -> Result<WireResponse, &'static str> {
        if !valid_path(path) {
            return Err("desktop proxy path invalid");
        }
        let url = format!("{}{}", self.origin, path);
        let secrets = if authenticated {
            self.vault
                .lock()
                .map_err(|_| "desktop vault unavailable")?
                .read()?
        } else {
            None
        };

        let cookie = secrets
            .as_ref()
            .map(|values| Zeroizing::new(format!("{}={}", SESSION_COOKIE, values.token)));
        let mut protected_values = Vec::new();
        if let Some(values) = secrets.as_ref() {
            protected_values.push(Zeroizing::new(values.token.clone()));
            protected_values.push(Zeroizing::new(values.csrf.clone()));
            protected_values.push(Zeroizing::new(format!("Bearer {}", values.token)));
        }
        if let Some(value) = cookie.as_ref() {
            protected_values.push(Zeroizing::new(value.to_string()));
        }
        let mut multipart = if method == "POST" && path == "/files/objects" {
            Some(multipart_upload(
                body.ok_or("desktop upload body invalid")?,
            )?)
        } else {
            None
        };
        let result = match (method, body) {
            ("GET", None) => {
                let mut builder = self
                    .agent
                    .get(&url)
                    .header(CONTROL_HEADER, self.control.as_str())
                    .header("Accept", "application/json");
                if let Some(value) = cookie.as_ref() {
                    builder = builder.header("Cookie", value.as_str());
                }
                builder.call()
            }
            ("POST", value) => {
                let mut builder = self
                    .agent
                    .post(&url)
                    .header(CONTROL_HEADER, self.control.as_str())
                    .header("Accept", "application/json");
                if let Some(value) = cookie.as_ref() {
                    builder = builder.header("Cookie", value.as_str());
                }
                if let Some(values) = secrets.as_ref() {
                    builder = builder.header("X-CSRF-Token", &values.csrf);
                }
                match multipart.as_ref() {
                    Some((content_type, bytes)) => builder
                        .header("Content-Type", content_type)
                        .send(bytes.as_slice()),
                    None => match value {
                        Some(value) => builder.send_json(value),
                        None => builder.send_empty(),
                    },
                }
            }
            ("PUT", Some(value)) => {
                let mut builder = self
                    .agent
                    .put(&url)
                    .header(CONTROL_HEADER, self.control.as_str())
                    .header("Accept", "application/json");
                if let Some(value) = cookie.as_ref() {
                    builder = builder.header("Cookie", value.as_str());
                }
                if let Some(values) = secrets.as_ref() {
                    builder = builder.header("X-CSRF-Token", &values.csrf);
                }
                builder.send_json(value)
            }
            ("PATCH", Some(value)) => {
                let mut builder = self
                    .agent
                    .patch(&url)
                    .header(CONTROL_HEADER, self.control.as_str())
                    .header("Accept", "application/json");
                if let Some(value) = cookie.as_ref() {
                    builder = builder.header("Cookie", value.as_str());
                }
                if let Some(values) = secrets.as_ref() {
                    builder = builder.header("X-CSRF-Token", &values.csrf);
                }
                builder.send_json(value)
            }
            ("DELETE", None) => {
                let mut builder = self
                    .agent
                    .delete(&url)
                    .header(CONTROL_HEADER, self.control.as_str())
                    .header("Accept", "application/json");
                if let Some(value) = cookie.as_ref() {
                    builder = builder.header("Cookie", value.as_str());
                }
                if let Some(values) = secrets.as_ref() {
                    builder = builder.header("X-CSRF-Token", &values.csrf);
                }
                builder.call()
            }
            _ => return Err("desktop proxy method invalid"),
        };
        if let Some((_, bytes)) = multipart.as_mut() {
            bytes.zeroize();
        }
        let mut response = result.map_err(|_| "desktop sidecar request failed")?;
        let status = response.status().as_u16();
        let replacement = unique_header(response.headers(), "set-cookie")?;
        let rotated_csrf = unique_header(response.headers(), "x-csrf-token")?;
        if let Some(header) = replacement.as_ref() {
            protected_values.push(Zeroizing::new(header.clone()));
            if let Some(token) = parse_session_cookie(header)? {
                protected_values.push(Zeroizing::new(format!("Bearer {token}")));
                protected_values.push(Zeroizing::new(token));
            }
        }
        if let Some(csrf) = rotated_csrf.as_ref() {
            if !valid_secret(csrf) {
                return Err("desktop session response invalid");
            }
            protected_values.push(Zeroizing::new(csrf.clone()));
        }
        let mut bytes = response
            .body_mut()
            .with_config()
            .limit(MAX_RESPONSE_BYTES)
            .read_to_vec()
            .map_err(|_| "desktop sidecar response invalid")?;
        let binary_response =
            path.starts_with("/files/objects/") && path.ends_with("/content") && status < 300;
        let parsed = if binary_response {
            let media_type = unique_header(response.headers(), "content-type")?
                .ok_or("desktop binary response invalid")?;
            if !matches!(
                media_type.as_str(),
                "image/png"
                    | "image/jpeg"
                    | "application/pdf"
                    | "text/plain; charset=utf-8"
                    | "text/plain"
            ) {
                return Err("desktop binary response invalid");
            }
            Ok(
                serde_json::json!({"encoding":"base64", "mediaType":media_type, "data":STANDARD.encode(&bytes)}),
            )
        } else if bytes.is_empty() {
            Ok(Value::Null)
        } else {
            serde_json::from_slice(&bytes).map_err(|_| "desktop sidecar response invalid")
        };
        bytes.zeroize();
        let body = parsed?;
        Ok(WireResponse {
            response: DesktopResponse { status, body },
            rotation: Rotation {
                cookie: replacement,
                csrf: rotated_csrf,
            },
            protected_values,
        })
    }
}

fn multipart_upload(body: &Value) -> Result<(String, Vec<u8>), &'static str> {
    let upload: crate::product_contract::UploadBody =
        serde_json::from_value(body.clone()).map_err(|_| "desktop upload body invalid")?;
    let data = STANDARD
        .decode(upload.data.as_bytes())
        .map_err(|_| "desktop upload body invalid")?;
    if data.is_empty() || data.len() > 10 * 1024 * 1024 {
        return Err("desktop upload body invalid");
    }
    let boundary = "go-admin-plus-desktop-boundary-7f6d8e3a";
    let mut bytes = Vec::with_capacity(data.len() + 512);
    bytes.extend_from_slice(format!("--{boundary}\r\nContent-Disposition: form-data; name=\"file\"; filename=\"{}\"\r\nContent-Type: {}\r\n\r\n", upload.name, upload.media_type).as_bytes());
    bytes.extend_from_slice(&data);
    bytes.extend_from_slice(format!("\r\n--{boundary}--\r\n").as_bytes());
    Ok((format!("multipart/form-data; boundary={boundary}"), bytes))
}

fn decode_manifest(value: Value) -> Result<ManifestWire, &'static str> {
    let manifest: ManifestWire =
        serde_json::from_value(value).map_err(|_| "desktop navigation response invalid")?;
    let permissions: BTreeSet<_> = manifest.permission_codes.iter().collect();
    let menu_keys: BTreeSet<_> = manifest.menus.iter().map(|menu| &menu.key).collect();
    let menu_paths: BTreeSet<_> = manifest.menus.iter().map(|menu| &menu.path).collect();
    if !matches!(manifest.data_scope.as_str(), "self" | "all")
        || permissions.len() != manifest.permission_codes.len()
        || menu_keys.len() != manifest.menus.len()
        || menu_paths.len() != manifest.menus.len()
        || manifest
            .permission_codes
            .iter()
            .any(|value| !valid_permission(value))
        || manifest.menus.iter().any(|menu| {
            menu.key.is_empty()
                || menu.key.len() > 64
                || menu.label.is_empty()
                || menu.label.chars().count() > 64
                || menu.sort_order < 0
                || !valid_path(&menu.path)
                || menu.path.contains('?')
                || !valid_permission(&menu.permission_code)
                || !permissions.contains(&menu.permission_code)
        })
    {
        return Err("desktop navigation response invalid");
    }
    Ok(manifest)
}

fn parse_session_cookie(header: &str) -> Result<Option<String>, &'static str> {
    if header.contains(',') || header.len() > 512 {
        return Err("desktop session cookie invalid");
    }
    let mut parts = header.split(';').map(str::trim);
    let pair = parts.next().ok_or("desktop session cookie invalid")?;
    let prefix = format!("{}=", SESSION_COOKIE);
    let value = pair
        .strip_prefix(&prefix)
        .ok_or("desktop session cookie invalid")?;
    if !value.is_empty() && !valid_secret(value) {
        return Err("desktop session cookie invalid");
    }
    let mut secure = false;
    let mut http_only = false;
    let mut path = false;
    let mut same_site = false;
    let mut deletion = false;
    for part in parts {
        let lower = part.to_ascii_lowercase();
        match lower.as_str() {
            "secure" if !secure => secure = true,
            "httponly" if !http_only => http_only = true,
            "path=/" if !path => path = true,
            "samesite=strict" if !same_site => same_site = true,
            "max-age=0" if !deletion => deletion = true,
            _ => return Err("desktop session cookie invalid"),
        }
    }
    if !secure || !http_only || !path || !same_site || deletion != value.is_empty() {
        return Err("desktop session cookie invalid");
    }
    Ok((!value.is_empty()).then(|| value.to_owned()))
}

fn unique_header(
    headers: &ureq::http::HeaderMap,
    name: &str,
) -> Result<Option<String>, &'static str> {
    let mut values = headers.get_all(name).iter();
    let first = values
        .next()
        .map(|value| value.to_str().map(str::to_owned))
        .transpose()
        .map_err(|_| "desktop sidecar response headers invalid")?;
    if values.next().is_some() {
        return Err("desktop sidecar response headers invalid");
    }
    Ok(first)
}

fn valid_secret(value: &str) -> bool {
    value.len() == 43
        && value
            .bytes()
            .all(|byte| byte.is_ascii_alphanumeric() || byte == b'_' || byte == b'-')
}

fn valid_permission(value: &str) -> bool {
    let segments: Vec<_> = value.split('.').collect();
    (2..=3).contains(&segments.len())
        && segments.iter().all(|segment| {
            !segment.is_empty()
                && segment.len() <= 64
                && segment.as_bytes()[0].is_ascii_lowercase()
                && segment
                    .bytes()
                    .all(|byte| byte.is_ascii_lowercase() || byte.is_ascii_digit() || byte == b'-')
        })
}

fn validate_profile(profile: &PublicProfile) -> Result<(), &'static str> {
    if profile.id.is_empty()
        || profile.id.len() > 128
        || profile.username.is_empty()
        || profile.username.len() > 64
        || profile.display_name.is_empty()
        || profile.display_name.chars().count() > 120
        || profile.email.is_empty()
        || profile.email.len() > 254
        || profile
            .avatar_ref
            .as_ref()
            .is_some_and(|value| value.len() > 512)
    {
        return Err("desktop profile response invalid");
    }
    Ok(())
}

fn valid_origin(origin: &str) -> bool {
    let Some(port) = origin.strip_prefix("http://127.0.0.1:") else {
        return false;
    };
    !port.is_empty()
        && port.bytes().all(|byte| byte.is_ascii_digit())
        && port.parse::<u16>().is_ok_and(|value| value > 0)
}

fn valid_path(path: &str) -> bool {
    path.starts_with('/')
        && !path.starts_with("//")
        && path.len() <= 2048
        && !path.contains('#')
        && !path.contains('\\')
        && !path.contains("..")
        && !path.bytes().any(|byte| byte < 0x20 || byte == 0x7f)
        && !path.to_ascii_lowercase().contains("%2f")
        && !path.to_ascii_lowercase().contains("%5c")
}

fn sensitive_key(key: &str) -> bool {
    let normalized: String = key
        .chars()
        .filter(|value| value.is_ascii_alphanumeric())
        .map(|value| value.to_ascii_lowercase())
        .collect();
    matches!(
        normalized.as_str(),
        "csrftoken"
            | "csrf"
            | "token"
            | "accesstoken"
            | "refreshtoken"
            | "sessiontoken"
            | "password"
            | "passwordhash"
            | "cookie"
            | "authorization"
            | "session"
            | "credential"
            | "credentials"
            | "secret"
            | "secretkey"
            | "privatekey"
            | "clientsecret"
            | "dsn"
            | "databaseurl"
    )
}

#[cfg(test)]
fn contains_secret_key(value: &Value) -> bool {
    match value {
        Value::Object(values) => values
            .iter()
            .any(|(key, value)| sensitive_key(key) || contains_secret_key(value)),
        Value::Array(values) => values.iter().any(contains_secret_key),
        _ => false,
    }
}

fn secret_like_text(value: &str) -> bool {
    let lower = value.to_ascii_lowercase();
    if lower.contains("bearer ")
        || lower.contains("-----begin private key-----")
        || lower.contains("-----begin ec private key-----")
        || lower.contains("postgres://")
        || lower.contains("postgresql://")
        || lower.contains("mysql://")
        || lower.contains("jdbc:")
        || lower.contains("$argon2")
        || lower.contains("$2a$")
        || lower.contains("$2b$")
        || lower.contains("$2y$")
    {
        return true;
    }
    if let Some(scheme) = lower.find("://") {
        let authority = &value[scheme + 3..];
        let end = authority.find('/').unwrap_or(authority.len());
        let user_info = &authority[..end];
        if user_info.contains('@')
            && user_info
                .split('@')
                .next()
                .is_some_and(|part| part.contains(':'))
        {
            return true;
        }
    }
    if value
        .split(|character: char| {
            !(character.is_ascii_alphanumeric() || character == '_' || character == '-')
        })
        .any(looks_like_opaque_secret)
    {
        return true;
    }
    value.split_whitespace().any(|part| {
        let segments: Vec<_> = part.split('.').collect();
        segments.len() == 3
            && segments.iter().all(|segment| {
                segment.len() >= 8
                    && segment
                        .bytes()
                        .all(|byte| byte.is_ascii_alphanumeric() || byte == b'_' || byte == b'-')
            })
    })
}

fn looks_like_opaque_secret(value: &str) -> bool {
    if value.len() != 43 {
        return false;
    }
    let classes = [
        value.bytes().any(|byte| byte.is_ascii_lowercase()),
        value.bytes().any(|byte| byte.is_ascii_uppercase()),
        value.bytes().any(|byte| byte.is_ascii_digit()),
        value.bytes().any(|byte| matches!(byte, b'-' | b'_')),
    ]
    .into_iter()
    .filter(|present| *present)
    .count();
    let distinct: BTreeSet<_> = value.bytes().collect();
    classes >= 2 && distinct.len() >= 12
}

fn contains_secret_material(value: &Value, protected: &[Zeroizing<String>]) -> bool {
    match value {
        Value::String(value) => {
            secret_like_text(value)
                || protected
                    .iter()
                    .any(|secret| !secret.is_empty() && value.contains(secret.as_str()))
        }
        Value::Object(values) => values
            .iter()
            .any(|(key, value)| sensitive_key(key) || contains_secret_material(value, protected)),
        Value::Array(values) => values
            .iter()
            .any(|value| contains_secret_material(value, protected)),
        _ => false,
    }
}

fn profile_contains_secret(profile: &PublicProfile, protected: &[Zeroizing<String>]) -> bool {
    [
        Some(profile.id.as_str()),
        Some(profile.username.as_str()),
        Some(profile.display_name.as_str()),
        Some(profile.email.as_str()),
        profile.avatar_ref.as_deref(),
    ]
    .into_iter()
    .flatten()
    .any(|value| {
        secret_like_text(value)
            || protected
                .iter()
                .any(|secret| !secret.is_empty() && value.contains(secret.as_str()))
    })
}

fn scrub_json(value: &mut Value) {
    match value {
        Value::Object(values) => {
            for value in values.values_mut() {
                scrub_json(value);
            }
            values.clear();
        }
        Value::Array(values) => {
            for value in values.iter_mut() {
                scrub_json(value);
            }
            values.clear();
        }
        Value::String(value) => value.zeroize(),
        other => *other = Value::Null,
    }
}

fn logout_status_allows_clear(status: u16) -> bool {
    matches!(status, 204 | 401)
}

fn finish_logout(
    status: u16,
    clear: impl FnOnce() -> Result<(), &'static str>,
) -> Result<LogoutResult, &'static str> {
    if !logout_status_allows_clear(status) {
        return Err("desktop logout request failed");
    }
    clear()?;
    Ok(LogoutResult {
        local_cleared: true,
        remote_revoked: true,
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rejects_session_and_traversal_from_generic_bridge() {
        assert!(!"/iam/session/current".starts_with("/demo/"));
        for path in [
            "//demo/products",
            "/demo/../iam/session/current",
            "/demo/%2fsecret",
            "/demo/x#fragment",
        ] {
            assert!(!valid_path(path), "accepted {path}");
        }
    }

    #[test]
    fn detects_nested_credential_keys() {
        for key in [
            "csrfToken",
            "session_token",
            "access-token",
            "password_hash",
        ] {
            let mut nested = serde_json::Map::new();
            nested.insert(key.to_owned(), Value::String("hidden".to_owned()));
            assert!(contains_secret_key(&serde_json::json!({"items":[nested]})));
        }
        assert!(!contains_secret_key(
            &serde_json::json!({"sessionTimeout":60,"tokenCount":2})
        ));
    }

    #[test]
    fn rejects_sensitive_material_in_nested_public_strings() {
        let secret = Zeroizing::new("abcdefghijklmnopqrstuvwxyzABCDEFGH123456789".to_owned());
        assert!(contains_secret_material(
            &serde_json::json!({"detail": ["prefix-abcdefghijklmnopqrstuvwxyzABCDEFGH123456789-suffix"]}),
            &[secret]
        ));
        assert!(!contains_secret_material(
            &serde_json::json!({"sessionTimeout": "tokenized-mode", "tokenCount": 2}),
            &[Zeroizing::new(
                "abcdefghijklmnopqrstuvwxyzABCDEFGH123456789".to_owned()
            )]
        ));
        for value in [
            "aaaaaaaa.bbbbbbbb.cccccccc",
            "-----BEGIN PRIVATE KEY-----",
            "postgresql://user:password@example.test/db",
            "$argon2id$v=19$m=65536,t=3,p=1$hash",
        ] {
            assert!(contains_secret_material(
                &serde_json::json!({"violations":[{"detail":value}]}),
                &[]
            ));
        }
        assert!(!contains_secret_material(
            &serde_json::json!({
                "name": "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopq",
                "description": "a".repeat(120)
            }),
            &[]
        ));
        assert!(contains_secret_material(
            &serde_json::json!({"detail":"abcdefghijklmnopqrstuvwxyzABCDEFGH123456789"}),
            &[]
        ));
    }

    #[test]
    fn public_identity_contains_scope_but_no_session_material() {
        let encoded = serde_json::to_value(IdentityResult::Authenticated {
            profile: PublicProfile {
                id: "account-1".to_owned(),
                username: "admin".to_owned(),
                display_name: "Administrator".to_owned(),
                email: "admin@example.test".to_owned(),
                avatar_ref: None,
            },
            permissions: vec!["demo.products.read".to_owned()],
            data_scope: "all".to_owned(),
        })
        .unwrap();
        assert_eq!(encoded["dataScope"], "all");
        assert!(!contains_secret_key(&encoded));
    }

    #[test]
    fn logout_only_clears_after_a_terminal_remote_status() {
        use std::cell::Cell;

        assert!(logout_status_allows_clear(204));
        assert!(logout_status_allows_clear(401));
        assert!(!logout_status_allows_clear(403));
        assert!(!logout_status_allows_clear(500));
        let cleared = Cell::new(false);
        assert!(
            finish_logout(500, || {
                cleared.set(true);
                Ok(())
            })
            .is_err()
        );
        assert!(!cleared.get());
        assert!(finish_logout(204, || Err("injected vault fault")).is_err());
    }

    #[test]
    fn session_cookie_parser_is_exact() {
        let token = "abcdefghijklmnopqrstuvwxyzABCDEFGH123456789";
        assert_eq!(
            parse_session_cookie(&format!(
                "{SESSION_COOKIE}={token}; Path=/; HttpOnly; Secure; SameSite=Strict"
            ))
            .unwrap()
            .as_deref(),
            Some(token)
        );
        assert!(parse_session_cookie(&format!("other={token}")).is_err());
        assert!(
            parse_session_cookie(&format!("{SESSION_COOKIE}={token}; Path=/; HttpOnly")).is_err()
        );
        assert_eq!(
            parse_session_cookie(&format!(
                "{SESSION_COOKIE}=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Strict"
            ))
            .unwrap(),
            None
        );
        assert!(
            parse_session_cookie(&format!(
                "{SESSION_COOKIE}=; Path=/; HttpOnly; Secure; SameSite=Strict"
            ))
            .is_err()
        );
    }

    #[test]
    fn duplicate_sensitive_headers_and_malformed_csrf_are_rejected() {
        let mut headers = ureq::http::HeaderMap::new();
        headers.append("set-cookie", "first=1".parse().unwrap());
        headers.append("set-cookie", "second=2".parse().unwrap());
        assert!(unique_header(&headers, "set-cookie").is_err());
        assert!(!valid_secret("abcdefghijklmnopqrstuvwxyzABCDEFGH12345678+"));
        assert!(valid_secret("abcdefghijklmnopqrstuvwxyzABCDEFGH123456789"));
    }
}
