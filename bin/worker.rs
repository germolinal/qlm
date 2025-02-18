use axum::{
    extract::Request,
    http::StatusCode,
    middleware::{self, Next},
    response::IntoResponse,
    Json, Router,
};
use prem_lm::{
    getport,
    ollama::{
        chat::{ChatReponse, ChatRequest},
        common::Ollamable,
        generate::{GenerateRequest, GenerateResponse},
    },
    orchestrator::response::LLMHandlerResponse,
};
use reqwest;
use serde::{de::DeserializeOwned, Serialize};

const OLLAMA_URL: &str = "http://localhost:11434";

async fn call_ollama<I: Ollamable, O: Serialize + DeserializeOwned>(
    data: I,
) -> LLMHandlerResponse<O> {
    let url = format!("{}/api/{}", OLLAMA_URL, data.path());
    let client = reqwest::Client::new();

    let b = serde_json::to_vec(&data).expect("could not serialise within worker/chat");
    let res = match client.post(url).body(b).send().await {
        Ok(r) => r,
        Err(e) => {
            let status = e.status().unwrap_or(StatusCode::INTERNAL_SERVER_ERROR);
            let e = format!("{}", e);
            return LLMHandlerResponse::Err((status, e));
        }
    };
    // let s: String = res.text().await.unwrap();
    // dbg!(s);
    let j: O = res
        .json()
        .await
        .expect("could not serialise Ollama response");
    LLMHandlerResponse::Ok((StatusCode::OK, j))
    // todo!()
}

async fn chat(Json(data): Json<ChatRequest>) -> impl IntoResponse {
    call_ollama::<ChatRequest, ChatReponse>(data).await
}

async fn generate(Json(data): Json<GenerateRequest>) -> impl IntoResponse {
    call_ollama::<GenerateRequest, GenerateResponse>(data).await
}

async fn healthcheck() -> impl IntoResponse {
    let body = GenerateRequest {
        model: Some("gemma2:2b".to_string()),
        prompt:
            "This is a health check of our systems. If you are OK, say it in a funny way (and making references to popupar culture) in no more than 5 words."
                .to_string(),
        ..Default::default()
    };
    call_ollama::<GenerateRequest, GenerateResponse>(body).await
}

async fn auth(request: Request, next: Next) -> impl IntoResponse {
    // eprintln!("checking API key");
    next.run(request).await
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    // build our application with a single route
    let app = Router::new()
        // Health check
        .route("/", axum::routing::get(healthcheck))
        // Ask for chat completion
        .route("/chat", axum::routing::post(chat))
        // Ask for completion
        .route("/generate", axum::routing::post(generate))
        // Implement authorisation/authentication
        .layer(middleware::from_fn(auth));

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], getport(4321)));
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    tracing::debug!("listening on {}", addr);
    eprintln!("listening on {}", addr);
    axum::serve(listener, app).await.unwrap();
}
