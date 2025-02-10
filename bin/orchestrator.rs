use axum::{
    extract::{Request, State},
    http::StatusCode,
    middleware::{self, Next},
    response::IntoResponse,
    Json, Router,
};
use prem_lm::getport;
use prem_lm::ollama::common::Ollamable;
use prem_lm::ollama::completion::CompletionResponse;
use prem_lm::orchestrator::LLMHandlerResponse;
use prem_lm::orchestrator::LLMRequest;
use prem_lm::{ollama::completion::CompletionRequest, orchestrator::OrchestratorState};

use std::env;
use std::fs;
use std::sync::Arc;
use std::sync::Mutex;
use tokio::sync::mpsc::Receiver;

use serde::{Deserialize, Serialize};

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
struct Chat {
    msg: String,
}

// #[axum_macros::debug_handler]
// async fn chat(
//     State(state): State<Arc<Mutex<OrchestratorState>>>,
//     data: Json<CompletionRequest>,
// ) -> impl IntoResponse {
//     handler("chat", state, data).await
// }

#[axum_macros::debug_handler]
async fn completion(
    State(state): State<Arc<Mutex<OrchestratorState>>>,
    Json(data): Json<CompletionRequest>,
) -> impl IntoResponse {
    handler("generate", state, data).await
}

#[axum_macros::debug_handler]
async fn async_completion(
    State(state): State<Arc<Mutex<OrchestratorState>>>,
    Json(data): Json<CompletionRequest>,
) -> impl IntoResponse {
    handler("generate", state, data).await
}

#[tracing::instrument]
async fn handler(
    path: &'static str,
    state: Arc<Mutex<OrchestratorState>>,
    mut data: CompletionRequest,
) -> LLMHandlerResponse<CompletionResponse> {
    OrchestratorState::call_ollama(state, path, data).await
}

#[tracing::instrument]
async fn hook_handler(
    path: &str,
    state: Arc<Mutex<OrchestratorState>>,
    data: CompletionRequest,
) -> LLMHandlerResponse<CompletionResponse> {
    if data.webhook().is_some() {
        // Queue message for later... might be process right away, or not
        let state_guard = state.lock().unwrap();
        if let Some(q) = &state_guard.queue {
            if let Err(_) = q.send(data.into()).await {
                let err = "Failed to enqueue task.".to_string();
                return LLMHandlerResponse::Err((StatusCode::INTERNAL_SERVER_ERROR, err.into()));
            }
            let msg = "Task queued for later processing.".to_string();
            let ret = CompletionResponse {
                response: msg.into(),
                ..CompletionResponse::default()
            };
            LLMHandlerResponse::Ok((StatusCode::ACCEPTED, ret))
        } else {
            unreachable!("queue should never be None");
        }
    } else {
        LLMHandlerResponse::Err((
            StatusCode::BAD_REQUEST,
            "a webhook field is required".into(),
        ))
    }
}

async fn auth(request: Request, next: Next) -> impl IntoResponse {
    eprintln!("checking API key");
    next.run(request).await
}

async fn process_queue(state: Arc<Mutex<OrchestratorState>>, mut receiver: Receiver<LLMRequest>) {
    while let Some(req) = receiver.recv().await {
        let hook = req
            .webhook()
            .expect("messages with no webhook should be caught before queueing them");

        let path = req.path();
        let ret = match req {
            LLMRequest::Completion(body) => {
                match OrchestratorState::call_ollama(state.clone(), path, body).await {
                    LLMHandlerResponse::Ok((_, d)) => {
                        serde_json::to_string(&d).expect("could not serialize")
                    }
                    LLMHandlerResponse::Err((_, e)) => e,
                }
            }
            LLMRequest::Chat(_r) => todo!(), //, handler(path, state, r.clone()).await,
        };
        let client = reqwest::Client::new();
        let r = client
            .post(hook)
            .header("Content-Type", "application/json")
            .body(ret)
            .send()
            .await;
        if let Err(e) = r {
            tracing::error!("{}", e)
        }
    }

    // let data = {
    //     let state_guard = state.lock().unwrap();

    //     let mut queue_rx = state_guard.queue.unwrap().subscribe();
    // };
    // while let Ok(task) = queue_rx.recv().await {
    //     // ... Actual processing logic ...
    //     let result = format!("Processed later: {}", task.data);

    //     // Send result to webhook
    //     if let Some(url) = task.webhook_url {
    //         let client = reqwest::Client::new();
    //         let _ = client.post(url).json(&result).send().await; // Handle errors appropriately
    //     }
    // }
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        // .with_target(false)
        .compact()
        .init();
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
    let (tx, rx) = tokio::sync::mpsc::channel::<LLMRequest>(100);
    state.queue = Some(tx);
    let state = Arc::new(Mutex::new(state));

    // build our application with a single route
    let app = Router::new()
        // Dashboard?
        // .route("/", axum::routing::get(healthcheck))
        // Ask for chat completion
        // .route("/chat", axum::routing::post(handler))
        // Ask for completion
        // .route("/chat", axum::routing::post(chat))
        .route("/generate", axum::routing::post(completion))
        .route("/async_generate", axum::routing::post(async_completion))
        // Implement authorisation/authentication
        .with_state(state.clone())
        // .layer(TraceLayer::new_for_http()); // Add the TraceLayer
        .layer(middleware::from_fn(auth));

    // Spawn the processing of new elements in the queue
    tokio::spawn(async move { process_queue(state, rx).await });

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], getport(3000)));
    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .expect("could not build listener");
    tracing::debug!("listening on {}", addr);

    axum::serve(listener, app).await.expect("server panic!");
}
