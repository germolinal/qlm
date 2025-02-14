use std::fmt::Debug;

use axum::{http::status::StatusCode, response::IntoResponse, Json};
use serde::Serialize;

#[derive(Clone, Debug)]
pub enum LLMHandlerResponse<T> {
    Ok((StatusCode, T)),
    Err((StatusCode, String)),
}

impl<T: Clone + Serialize + Debug> IntoResponse for LLMHandlerResponse<T> {
    fn into_response(self) -> axum::response::Response {
        match self {
            Self::Ok((code, obj)) => (code, Json(obj)).into_response(),
            Self::Err((code, msg)) => (code, msg).into_response(),
        }
    }
}
