use axum::{
    extract::{Request, State},
    http::StatusCode,
    middleware::{self, Next},
    response::IntoResponse,
    routing::post,
    Json, Router,
};
use prem_lm::{
    ollama::common::Ollamable,
    orchestrator::{
        response::LLMHandlerResponse,
        state::{OrchestratorState, SharedState},
    },
};
use prem_lm::{
    ollama::{
        chat::{ChatReponse, ChatRequest},
        generate::{GenerateRequest, GenerateResponse},
    },
    orchestrator::{request::LLMRequest, worker::Worker},
};
use serde::{de::DeserializeOwned, Serialize};
use tokio::sync::mpsc::channel;

use futures::lock::Mutex;
use std::env;
use std::fs;
use std::sync::Arc;

#[axum_macros::debug_handler]
async fn chat(
    State(state): State<SharedState>,
    Json(req): Json<ChatRequest>,
) -> LLMHandlerResponse<ChatReponse> {
    handler::<ChatRequest, ChatReponse>(state, req).await
}

#[axum_macros::debug_handler]
async fn generate(
    State(state): State<SharedState>,
    Json(req): Json<GenerateRequest>,
) -> LLMHandlerResponse<GenerateResponse> {
    handler::<GenerateRequest, GenerateResponse>(state, req).await
}

#[axum_macros::debug_handler]
async fn async_chat(
    State(state): State<SharedState>,
    Json(req): Json<ChatRequest>,
) -> LLMHandlerResponse<&'static str> {
    async_handler::<ChatRequest>(state, req).await
}

#[axum_macros::debug_handler]
async fn async_generate(
    State(state): State<SharedState>,
    Json(req): Json<GenerateRequest>,
) -> LLMHandlerResponse<&'static str> {
    async_handler::<GenerateRequest>(state, req).await
}

async fn handler<I: Ollamable, O: Serialize + DeserializeOwned>(
    state: SharedState,
    req: I,
) -> LLMHandlerResponse<O> {
    let mut guard = state.lock().await;
    let w_index = guard.get_available_worker();
    if let Some(i) = w_index {
        let url = guard.workers[i].address.clone();
        drop(guard);
        Worker::process::<I, O>(state, i, url, req).await
    } else {
        LLMHandlerResponse::Err((StatusCode::TOO_MANY_REQUESTS, "There are too many requests at the moment. Try using the async path if your request can wait".to_string()))
    }
}

async fn async_handler<I: Ollamable>(
    state: SharedState,
    req: I,
) -> LLMHandlerResponse<&'static str> {
    if req.webhook().is_none() {
        return LLMHandlerResponse::Err((
            StatusCode::BAD_REQUEST,
            "a webhook field is required for asynchronous processing".to_string(),
        ));
    }
    let mut guard = state.lock().await;
    let req: LLMRequest = req.into();
    guard.enqueue(req).await
}

async fn auth(request: Request, next: Next) -> impl IntoResponse {
    eprintln!("checking API key");
    next.run(request).await
}

#[tokio::main]
async fn main() {
    let args: Vec<String> = env::args().collect();
    let mut state: OrchestratorState = if args.len() == 1 {
        eprintln!("... no config file passed. Using an empty one");
        OrchestratorState::default()
    } else {
        let contents =
            fs::read_to_string(&args[1]).expect("Should have been able to read the file");
        match serde_json::from_str(&contents) {
            Ok(json) => json,
            Err(err) => {
                eprintln!("{}", err);
                std::process::exit(1)
            }
        }
    };

    let (tx, rx) = channel::<bool>(400);
    state.sender = Some(tx.clone());
    let state = Arc::new(Mutex::new(state));

    // build our application with a route
    let app = Router::new()
        .route("/chat", post(chat))
        .route("/generate", post(generate))
        .route("/async_chat", post(async_chat))
        .route("/async_generate", post(async_generate))
        .with_state(state.clone())
        // Auth
        .layer(middleware::from_fn(auth));

    tokio::spawn(async move {
        OrchestratorState::process_queue(rx, state).await;
    });

    // run it
    let listener = tokio::net::TcpListener::bind("127.0.0.1:8080")
        .await
        .unwrap();
    println!("listening on {}", listener.local_addr().unwrap());
    axum::serve(listener, app).await.unwrap();
}
