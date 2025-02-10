use serde::{Deserialize, Serialize};
use std::{
    fmt::Debug,
    sync::{
        atomic::{AtomicU16, Ordering},
        Arc, Mutex,
    },
    usize,
};

use crate::ollama::{
    chat::ChatRequest,
    common::Ollamable,
    completion::{CompletionRequest, CompletionResponse},
};
use axum::{http::StatusCode, response::IntoResponse, Json};
use tokio::sync::mpsc;

pub enum LLMRequest {
    Completion(CompletionRequest),
    Chat(ChatRequest),
}

impl LLMRequest {
    pub fn path(&self) -> &'static str {
        match self {
            Self::Completion(_) => "generate",
            Self::Chat(_) => "chat",
        }
    }
    pub fn webhook(&self) -> Option<String> {
        match self {
            Self::Completion(r) => r.webhook.clone(),
            Self::Chat(r) => r.webhook.clone(),
        }
    }
}

impl std::convert::From<CompletionRequest> for LLMRequest {
    fn from(item: CompletionRequest) -> Self {
        LLMRequest::Completion(item)
    }
}
impl std::convert::From<ChatRequest> for LLMRequest {
    fn from(item: ChatRequest) -> Self {
        LLMRequest::Chat(item)
    }
}

#[derive(Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct WorkerState {
    pub address: String,
    pub concurrency: u16,
    #[serde(skip)]
    pub nrequests: AtomicU16,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct OrchestratorState {
    pub model: String,
    pub workers: Vec<WorkerState>,
    #[serde(skip)]
    #[serde(default)]
    pub queue: Option<mpsc::Sender<LLMRequest>>,
}

impl Default for OrchestratorState {
    fn default() -> Self {
        Self {
            model: String::new(),
            workers: vec![],
            queue: None,
        }
    }
}

#[derive(Clone, Debug)]
pub enum LLMHandlerResponse<T: Clone + Serialize + Debug> {
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

impl OrchestratorState {
    pub fn get_url(&mut self) -> Option<String> {
        let mut p = 1.; // start with all full busy
        let mut i = None;
        self.workers.iter().enumerate().for_each(|(index, w)| {
            let util = w.nrequests.load(Ordering::Relaxed) as f32 / w.concurrency as f32;
            if util < p - 1e-4 {
                i = Some(index);
                p = util;
            }
        });
        if let Some(i) = i {
            self.workers[i].nrequests.fetch_add(1, Ordering::SeqCst);
            Some(self.workers[i].address.clone())
        } else {
            None
        }
    }

    pub async fn call_ollama(
        state: Arc<Mutex<Self>>,
        path: &'static str,
        mut body: CompletionRequest,
    ) -> LLMHandlerResponse<CompletionResponse> {
        let (model, url) = {
            let mut state_guard = state.lock().unwrap();
            let model = state_guard.model.clone();
            let url = state_guard.get_url();
            (model, url)
        };
        if let Some(base_url) = url {
            // send request
            let url = format!("{}/{}", base_url, path);

            body.set_model(model);
            // let body: CompletionRequest = data.into();

            // Forward response
            let client = reqwest::Client::new();
            let b = serde_json::to_string(&body).unwrap();
            let ret = client
                .post(url)
                .header("Content-Type", "application/json")
                .body(b)
                .send()
                .await;

            // Shape response
            match ret {
                Ok(v) => {
                    let status = v.status();
                    let txt = v.text().await.unwrap();
                    let body: CompletionResponse = serde_json::from_str(&txt).unwrap();
                    LLMHandlerResponse::Ok((status, body))
                }
                Err(e) => {
                    LLMHandlerResponse::Err((StatusCode::INTERNAL_SERVER_ERROR, format!("{}", e)))
                }
            }
        } else {
            // No webhook and no resources? Return 429
            let e = "the servers are currently busy. If you can wait, resend this request with a 'webhook' address, and your request will be queued and processed in due time".to_string();
            LLMHandlerResponse::Err((StatusCode::TOO_MANY_REQUESTS, e))
        }
    }
}
