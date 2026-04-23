use axum::{
    body::Bytes,
    extract::{Multipart, Query, State},
    http::{header, StatusCode},
    response::{IntoResponse, Response},
    Json,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::Semaphore;
use tracing::{error, info};
use opencv::{core, imgproc, prelude::*};

use crate::imgutil;
use super::cache::ImageCache;

// ── State ────────────────────────────────────────────────────────────────────

pub struct AppState {
    pub cache:     ImageCache,
    pub semaphore: Arc<Semaphore>,
    pub http:      reqwest::Client,
}

impl AppState {
    pub fn new() -> Self {
        let workers = num_cpus::get().clamp(1, 8);
        info!("Worker semaphore size: {}", workers);
        Self {
            cache:     ImageCache::new(200),
            semaphore: Arc::new(Semaphore::new(workers)),
            http:      reqwest::Client::builder()
                .timeout(std::time::Duration::from_secs(30))
                .pool_max_idle_per_host(20)
                .build()
                .expect("http client"),
        }
    }
}

// ── Request / Response types ─────────────────────────────────────────────────

#[derive(Deserialize)]
pub struct ImageUrlBody {
    pub image_url: String,
}

#[derive(Deserialize)]
pub struct FormatQuery {
    pub format: Option<String>,
}

#[derive(Serialize)]
struct ErrorBody {
    error: String,
}

fn err(status: StatusCode, msg: impl Into<String>) -> Response {
    (status, Json(ErrorBody { error: msg.into() })).into_response()
}

// ── Helpers ──────────────────────────────────────────────────────────────────

async fn download(client: &reqwest::Client, url: &str) -> Result<Vec<u8>, Response> {
    tracing::info!("Downloading from URL: {}", url);
    let resp = client.get(url).send().await.map_err(|e| {
        error!("Fetch error for {}: {}", url, e);
        err(StatusCode::UNPROCESSABLE_ENTITY, format!("fetch error: {e}"))
    })?;
    
    let status = resp.status();
    let content_type = resp.headers().get(header::CONTENT_TYPE)
        .and_then(|h| h.to_str().ok())
        .unwrap_or("unknown");
    let content_len = resp.content_length().unwrap_or(0);
    
    tracing::info!("Source response: status={}, type={}, len={}", status, content_type, content_len);

    if !status.is_success() {
        return Err(err(
            StatusCode::UNPROCESSABLE_ENTITY,
            format!("source returned {}", status),
        ));
    }
    resp.bytes().await
        .map(|b| b.to_vec())
        .map_err(|e| err(StatusCode::UNPROCESSABLE_ENTITY, format!("read error: {e}")))
}

fn image_response(data: Arc<Vec<u8>>, content_type: &'static str) -> Response {
    (
        [(header::CONTENT_TYPE, content_type)],
        Bytes::copy_from_slice(&data),
    )
        .into_response()
}

/// Run the CPU-intensive processing pipeline in a blocking thread.
async fn process(
    state: Arc<AppState>,
    raw: Vec<u8>,
    mode: &'static str,
    fmt: Option<String>,
    cache_key: Option<String>,
) -> Response {
    // Short-circuit already-processed images.
    if imgutil::has_processing_tag(&raw) {
        let ct = if fmt.as_deref() == Some("webp") {
            "image/webp"
        } else {
            "image/jpeg"
        };
        return image_response(Arc::new(raw), ct);
    }

    // Check LRU cache (URL requests only).
    if let Some(key) = &cache_key {
        if let Some(cached) = state.cache.get(key) {
            return image_response(cached.data, cached.content_type);
        }
    }

    // Acquire owned semaphore permit to bound concurrency (providing 'static lifetime).
    let permit = match state.semaphore.clone().acquire_owned().await {
        Ok(p) => p,
        Err(_) => return err(StatusCode::SERVICE_UNAVAILABLE, "server shutting down"),
    };

    let result = tokio::task::spawn_blocking(move || -> anyhow::Result<(Vec<u8>, &'static str)> {
        let _permit = permit; 

        let img = imgutil::decode_image(&raw)?;

        let out_mat = if mode == "remove" {
            let bboxes = imgutil::detect_watermark(&img)?;
            imgutil::remove_watermark(&img, &bboxes, "logo.png")?
        } else {
            let mut swapped = core::Mat::default();
            imgproc::cvt_color(&img, &mut swapped, imgproc::COLOR_BGR2RGB, 0)?;
            swapped
        };

        let (bytes, ct) = if fmt.as_deref() == Some("webp") {
            (imgutil::encode_webp(&out_mat, 80)?, "image/webp")
        } else {
            (imgutil::encode_jpeg(&out_mat, 90)?, "image/jpeg")
        };

        Ok((bytes, ct))
    })
    .await;

    match result {
        Ok(Ok((data, ct))) => {
            let arc = Arc::new(data);
            if let Some(key) = cache_key {
                state.cache.insert(key, (*arc).clone(), ct);
            }
            image_response(arc, ct)
        }
        Ok(Err(e)) => {
            error!("processing error: {e:#}");
            err(StatusCode::INTERNAL_SERVER_ERROR, "processing failed")
        }
        Err(e) => {
            error!("spawn_blocking panic: {e}");
            err(StatusCode::INTERNAL_SERVER_ERROR, "internal error")
        }
    }
}

// ── Route handlers ────────────────────────────────────────────────────────────

pub async fn health() -> impl IntoResponse {
    Json(serde_json::json!({"status": "ok"}))
}

pub async fn remove_url(
    State(state): State<Arc<AppState>>,
    Query(q): Query<FormatQuery>,
    Json(body): Json<ImageUrlBody>,
) -> Response {
    let raw = match download(&state.http, &body.image_url).await {
        Ok(d) => d,
        Err(r) => return r,
    };
    let cache_key = format!("{}remove{}", body.image_url, q.format.as_deref().unwrap_or(""));
    process(state, raw, "remove", q.format, Some(cache_key)).await
}

pub async fn remove_upload(
    State(state): State<Arc<AppState>>,
    Query(q): Query<FormatQuery>,
    mut multipart: Multipart,
) -> Response {
    let raw = match read_multipart(&mut multipart).await {
        Ok(d) => d,
        Err(r) => return r,
    };
    process(state, raw, "remove", q.format, None).await
}

pub async fn swap_url(
    State(state): State<Arc<AppState>>,
    Query(q): Query<FormatQuery>,
    Json(body): Json<ImageUrlBody>,
) -> Response {
    let raw = match download(&state.http, &body.image_url).await {
        Ok(d) => d,
        Err(r) => return r,
    };
    let cache_key = format!("{}swap{}", body.image_url, q.format.as_deref().unwrap_or(""));
    process(state, raw, "swap", q.format, Some(cache_key)).await
}

pub async fn swap_upload(
    State(state): State<Arc<AppState>>,
    Query(q): Query<FormatQuery>,
    mut multipart: Multipart,
) -> Response {
    let raw = match read_multipart(&mut multipart).await {
        Ok(d) => d,
        Err(r) => return r,
    };
    process(state, raw, "swap", q.format, None).await
}

async fn read_multipart(mp: &mut Multipart) -> Result<Vec<u8>, Response> {
    while let Ok(Some(field)) = mp.next_field().await {
        if field.name() == Some("file") {
            let data = field.bytes().await.map_err(|e| {
                err(StatusCode::BAD_REQUEST, format!("read error: {e}"))
            })?;
            if data.len() > 20 * 1024 * 1024 {
                return Err(err(StatusCode::BAD_REQUEST, "file too large (max 20 MB)"));
            }
            return Ok(data.to_vec());
        }
    }
    Err(err(StatusCode::BAD_REQUEST, "no 'file' field in multipart body"))
}
