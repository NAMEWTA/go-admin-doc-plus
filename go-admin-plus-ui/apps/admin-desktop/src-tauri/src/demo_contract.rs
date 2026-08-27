use std::collections::{BTreeMap, BTreeSet};

use serde::{Deserialize, Serialize};
use serde_json::Value;
use time::{OffsetDateTime, format_description::well_known::Rfc3339};
use uuid::Uuid;

use crate::proxy::{DesktopRequest, DesktopResponse};

#[derive(Clone, Copy)]
enum SuccessShape {
    Page,
    Product(u16),
    Empty,
}

pub struct ValidatedRequest {
    pub path: String,
    pub method: &'static str,
    pub body: Option<Value>,
    success: SuccessShape,
    operation: Operation,
}

impl ValidatedRequest {
    pub fn required_permission(&self) -> &'static str {
        match self.operation {
            Operation::List | Operation::Get => "demo.products.read",
            Operation::Create | Operation::Update => "demo.products.write",
            Operation::Delete => "demo.products.delete",
        }
    }
}

pub fn authorization_failure(authenticated: bool) -> DesktopResponse {
    let (status, category, code, title) = if authenticated {
        (
            403,
            "authorization",
            "PERMISSION_DENIED",
            "Permission denied",
        )
    } else {
        (
            401,
            "authentication",
            "SESSION_REQUIRED",
            "Session required",
        )
    };
    DesktopResponse {
        status,
        body: serde_json::json!({
            "type": format!("urn:go-admin-plus:problem:{}", code.to_ascii_lowercase().replace('_', "-")),
            "title": title,
            "status": status,
            "category": category,
            "code": code,
            "traceId": "00000000000000000000000000000000"
        }),
    }
}

#[derive(Clone, Copy)]
enum Operation {
    List,
    Create,
    Get,
    Update,
    Delete,
}

#[derive(Deserialize, Serialize)]
#[serde(deny_unknown_fields, rename_all = "camelCase")]
struct ProductInput {
    sku: String,
    name: String,
    description: String,
    price_cents: i64,
    status: String,
}

#[derive(Deserialize, Serialize)]
#[serde(deny_unknown_fields, rename_all = "camelCase")]
struct UpdateProductInput {
    sku: String,
    name: String,
    description: String,
    price_cents: i64,
    status: String,
    revision: i64,
}

#[derive(Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct DeleteProductsInput {
    products: Vec<DeleteTarget>,
}

#[derive(Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct DeleteTarget {
    id: String,
    revision: i64,
}

#[derive(Deserialize, Serialize)]
#[serde(deny_unknown_fields, rename_all = "camelCase")]
struct Product {
    id: String,
    sku: String,
    name: String,
    description: String,
    price_cents: i64,
    status: String,
    revision: i64,
    created_at: String,
    updated_at: String,
}

#[derive(Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct ProductPage {
    rows: Vec<Product>,
    total: i64,
}

#[derive(Deserialize, Serialize)]
#[serde(deny_unknown_fields, rename_all = "camelCase")]
struct Problem {
    r#type: String,
    title: String,
    status: u16,
    category: String,
    code: String,
    trace_id: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    detail: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    instance: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    violations: Option<Vec<Violation>>,
}

#[derive(Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct Violation {
    field: String,
    rule: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    message: Option<String>,
}

pub fn validate_request(request: DesktopRequest) -> Result<ValidatedRequest, &'static str> {
    let DesktopRequest { path, method, body } = request;
    if path == "/demo/products" || path.starts_with("/demo/products?") {
        return match method.as_str() {
            "GET" if body.is_none() => {
                validate_list_query(&path)?;
                Ok(ValidatedRequest {
                    path,
                    method: "GET",
                    body: None,
                    success: SuccessShape::Page,
                    operation: Operation::List,
                })
            }
            "POST" if path == "/demo/products" => {
                let value: ProductInput = decode_body(body)?;
                validate_product_input(&value)?;
                Ok(ValidatedRequest {
                    path,
                    method: "POST",
                    body: Some(
                        serde_json::to_value(value).map_err(|_| "desktop request body invalid")?,
                    ),
                    success: SuccessShape::Product(201),
                    operation: Operation::Create,
                })
            }
            _ => Err("desktop request is not allowed"),
        };
    }
    if path == "/demo/products/batch-delete" {
        if method != "POST" {
            return Err("desktop request is not allowed");
        }
        let value: DeleteProductsInput = decode_body(body)?;
        if value.products.is_empty() || value.products.len() > 100 {
            return Err("desktop request body invalid");
        }
        let mut ids = BTreeSet::new();
        for target in &value.products {
            validate_id(&target.id)?;
            if target.revision < 1 || !ids.insert(&target.id) {
                return Err("desktop request body invalid");
            }
        }
        return Ok(ValidatedRequest {
            path,
            method: "POST",
            body: Some(serde_json::to_value(value).map_err(|_| "desktop request body invalid")?),
            success: SuccessShape::Empty,
            operation: Operation::Delete,
        });
    }
    let Some(id) = path.strip_prefix("/demo/products/") else {
        return Err("desktop request is not allowed");
    };
    if id.contains('?') || id.contains('/') {
        return Err("desktop request is not allowed");
    }
    validate_id(id)?;
    match method.as_str() {
        "GET" if body.is_none() => Ok(ValidatedRequest {
            path,
            method: "GET",
            body: None,
            success: SuccessShape::Product(200),
            operation: Operation::Get,
        }),
        "PATCH" => {
            let value: UpdateProductInput = decode_body(body)?;
            validate_product_fields(
                &value.sku,
                &value.name,
                &value.description,
                value.price_cents,
                &value.status,
            )?;
            if value.revision < 1 {
                return Err("desktop request body invalid");
            }
            Ok(ValidatedRequest {
                path,
                method: "PATCH",
                body: Some(
                    serde_json::to_value(value).map_err(|_| "desktop request body invalid")?,
                ),
                success: SuccessShape::Product(200),
                operation: Operation::Update,
            })
        }
        _ => Err("desktop request is not allowed"),
    }
}

pub fn validate_response(
    request: &ValidatedRequest,
    response: DesktopResponse,
) -> Result<DesktopResponse, &'static str> {
    let status = response.status;
    let body = match request.success {
        SuccessShape::Page if status == 200 => {
            let page: ProductPage = serde_json::from_value(response.body)
                .map_err(|_| "desktop response body invalid")?;
            if page.total < 0 || page.rows.len() > 100 {
                return Err("desktop response body invalid");
            }
            for product in &page.rows {
                validate_product(product)?;
            }
            serde_json::to_value(page).map_err(|_| "desktop response body invalid")?
        }
        SuccessShape::Product(expected) if status == expected => {
            let product: Product = serde_json::from_value(response.body)
                .map_err(|_| "desktop response body invalid")?;
            validate_product(&product)?;
            serde_json::to_value(product).map_err(|_| "desktop response body invalid")?
        }
        SuccessShape::Empty if status == 204 && response.body.is_null() => Value::Null,
        _ if allowed_error(request.operation, status) => {
            let problem: Problem = serde_json::from_value(response.body)
                .map_err(|_| "desktop response body invalid")?;
            validate_problem(status, &problem)?;
            serde_json::to_value(problem).map_err(|_| "desktop response body invalid")?
        }
        _ => return Err("desktop response status invalid"),
    };
    Ok(DesktopResponse { status, body })
}

fn decode_body<T: for<'de> Deserialize<'de>>(body: Option<Value>) -> Result<T, &'static str> {
    serde_json::from_value(body.ok_or("desktop request body invalid")?)
        .map_err(|_| "desktop request body invalid")
}

fn validate_list_query(path: &str) -> Result<(), &'static str> {
    let query = path.split_once('?').map_or("", |(_, value)| value);
    let mut values = BTreeMap::new();
    for (key, value) in url::form_urlencoded::parse(query.as_bytes()) {
        if !matches!(
            key.as_ref(),
            "search" | "page" | "pageSize" | "sort" | "direction"
        ) || values
            .insert(key.into_owned(), value.into_owned())
            .is_some()
        {
            return Err("desktop request query invalid");
        }
    }
    if values
        .get("search")
        .is_some_and(|value| value.chars().count() > 100)
        || !number_in_range(values.get("page"), 1, 1_000_000)
        || !number_in_range(values.get("pageSize"), 1, 100)
        || values.get("sort").is_some_and(|value| {
            !matches!(value.as_str(), "sku" | "name" | "priceCents" | "updatedAt")
        })
        || values
            .get("direction")
            .is_some_and(|value| !matches!(value.as_str(), "ascending" | "descending"))
    {
        return Err("desktop request query invalid");
    }
    Ok(())
}

fn number_in_range(value: Option<&String>, minimum: u64, maximum: u64) -> bool {
    value.is_none_or(|value| {
        value
            .parse::<u64>()
            .is_ok_and(|number| (minimum..=maximum).contains(&number))
    })
}

fn validate_product_input(value: &ProductInput) -> Result<(), &'static str> {
    validate_product_fields(
        &value.sku,
        &value.name,
        &value.description,
        value.price_cents,
        &value.status,
    )
}

fn validate_product(value: &Product) -> Result<(), &'static str> {
    validate_id(&value.id)?;
    validate_product_fields(
        &value.sku,
        &value.name,
        &value.description,
        value.price_cents,
        &value.status,
    )?;
    if value.revision < 1
        || OffsetDateTime::parse(&value.created_at, &Rfc3339).is_err()
        || OffsetDateTime::parse(&value.updated_at, &Rfc3339).is_err()
    {
        return Err("desktop response body invalid");
    }
    Ok(())
}

fn validate_product_fields(
    sku: &str,
    name: &str,
    description: &str,
    price: i64,
    status: &str,
) -> Result<(), &'static str> {
    let sku_bytes = sku.as_bytes();
    if !(3..=32).contains(&sku_bytes.len())
        || !sku_bytes[0].is_ascii_uppercase() && !sku_bytes[0].is_ascii_digit()
        || !sku_bytes.iter().all(|byte| {
            byte.is_ascii_uppercase() || byte.is_ascii_digit() || *byte == b'_' || *byte == b'-'
        })
        || !(3..=120).contains(&name.chars().count())
        || description.chars().count() > 500
        || !(0..=100_000_000).contains(&price)
        || !matches!(status, "active" | "inactive")
    {
        return Err("desktop request body invalid");
    }
    Ok(())
}

fn validate_id(value: &str) -> Result<(), &'static str> {
    let parsed = Uuid::parse_str(value).map_err(|_| "desktop request path invalid")?;
    if parsed.hyphenated().to_string() != value.to_ascii_lowercase() {
        return Err("desktop request path invalid");
    }
    Ok(())
}

fn validate_problem(status: u16, problem: &Problem) -> Result<(), &'static str> {
    let contract = match status {
        400 => ("validation", "REQUEST_INVALID"),
        401 => ("authentication", "SESSION_REQUIRED"),
        403 => ("authorization", "PERMISSION_DENIED"),
        404 => ("not_found", "RESOURCE_NOT_FOUND"),
        409 => ("conflict", "RESOURCE_CONFLICT"),
        500 => ("internal", "INTERNAL_ERROR"),
        _ => return Err("desktop response body invalid"),
    };
    if problem.status != status
        || problem.category != contract.0
        || problem.code != contract.1
        || !problem.r#type.starts_with("urn:go-admin-plus:problem:")
        || problem.r#type.len() > 128
        || problem.title.is_empty()
        || problem.title.len() > 128
        || problem.r#type
            != format!(
                "urn:go-admin-plus:problem:{}",
                contract.1.to_ascii_lowercase().replace('_', "-")
            )
        || problem.trace_id.len() != 32
        || !problem
            .trace_id
            .bytes()
            .all(|byte| byte.is_ascii_hexdigit() && !byte.is_ascii_uppercase())
        || problem
            .detail
            .as_ref()
            .is_some_and(|value| value.len() > 512)
        || problem
            .instance
            .as_ref()
            .is_some_and(|value| value.len() > 256)
        || problem.violations.as_ref().is_some_and(|values| {
            values.len() > 32
                || values.iter().any(|item| {
                    item.field.len() > 64
                        || item.rule.len() > 64
                        || item.message.as_ref().is_some_and(|value| value.len() > 256)
                })
        })
    {
        return Err("desktop response body invalid");
    }
    Ok(())
}

fn allowed_error(operation: Operation, status: u16) -> bool {
    match operation {
        Operation::List => matches!(status, 400 | 401 | 403 | 500),
        Operation::Create => matches!(status, 400 | 401 | 403 | 409 | 500),
        Operation::Get => matches!(status, 401 | 403 | 404 | 500),
        Operation::Update | Operation::Delete => {
            matches!(status, 400 | 401 | 403 | 404 | 409 | 500)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn request(path: &str, method: &str, body: Option<Value>) -> DesktopRequest {
        DesktopRequest {
            path: path.to_owned(),
            method: method.to_owned(),
            body,
        }
    }

    #[test]
    fn accepts_only_declared_demo_operations_and_shapes() {
        assert!(
            validate_request(request(
                "/demo/products?page=1&pageSize=20&sort=sku&direction=ascending&search=%25_",
                "GET",
                None
            ))
            .is_ok()
        );
        assert!(
            validate_request(request(
                "/demo/products/550e8400-e29b-41d4-a716-446655440000",
                "GET",
                None
            ))
            .is_ok()
        );
        for value in [
            request("/demo/products/anything", "GET", None),
            request("/demo/products?unknown=1", "GET", None),
            request("/demo/products?page=1&page=2", "GET", None),
            request("/demo/products/batch-delete/x", "POST", Some(Value::Null)),
            request("/demo/products", "DELETE", None),
            request("/iam/session/current", "GET", None),
        ] {
            assert!(validate_request(value).is_err());
        }
    }

    #[test]
    fn rejects_unknown_or_secret_request_fields() {
        let valid = serde_json::json!({"sku":"ABC","name":"Visible product","description":"","priceCents":1,"status":"active"});
        assert!(validate_request(request("/demo/products", "POST", Some(valid))).is_ok());
        let secret = serde_json::json!({"sku":"ABC","name":"Visible product","description":"","priceCents":1,"status":"active","csrfToken":"hidden"});
        assert!(validate_request(request("/demo/products", "POST", Some(secret))).is_err());
    }

    #[test]
    fn rejects_operation_specific_error_status_and_malformed_product_time() {
        let list = validate_request(request("/demo/products", "GET", None)).unwrap();
        let conflict = DesktopResponse {
            status: 409,
            body: serde_json::json!({
                "type":"urn:go-admin-plus:problem:resource-conflict","title":"Resource conflict",
                "status":409,"category":"conflict","code":"RESOURCE_CONFLICT",
                "traceId":"0123456789abcdef0123456789abcdef"
            }),
        };
        assert!(validate_response(&list, conflict).is_err());

        let detail = validate_request(request(
            "/demo/products/550e8400-e29b-41d4-a716-446655440000",
            "GET",
            None,
        ))
        .unwrap();
        let malformed = DesktopResponse {
            status: 200,
            body: serde_json::json!({
                "id":"550e8400-e29b-41d4-a716-446655440000","sku":"ABC","name":"Visible product",
                "description":"","priceCents":1,"status":"active","revision":1,
                "createdAt":"not-a-time","updatedAt":"2026-08-27T00:00:00Z"
            }),
        };
        assert!(validate_response(&detail, malformed).is_err());
    }
}
