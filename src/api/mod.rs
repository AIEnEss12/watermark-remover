pub mod cache;
pub mod handlers;

use axum::{
    routing::{get, post},
    Router,
};
use std::sync::Arc;

pub fn router() -> Router {
    let state = Arc::new(handlers::AppState::new());
    Router::new()
        .route("/health",         get(handlers::health))
        .route("/remove",         post(handlers::remove_url))
        .route("/remove/upload",  post(handlers::remove_upload))
        .route("/swap-rb",        post(handlers::swap_url))
        .route("/swap-rb/upload", post(handlers::swap_upload))
        .with_state(state)
}
