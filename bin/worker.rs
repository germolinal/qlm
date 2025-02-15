use axum::{
    extract::Request,
    middleware::{self, Next},
    response::IntoResponse,
    Json, Router,
};
use prem_lm::{
    getport,
    ollama::{chat::ChatRequest, common::Ollamable, generate::GenerateRequest},
};
use reqwest;
use serde::{Deserialize, Serialize};

const OLLAMA_URL: &str = "http://localhost:11434";

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
struct Chat {
    msg: String,
}

async fn chat(Json(data): Json<ChatRequest>) -> impl IntoResponse {
    let url = format!("{}/api/{}", OLLAMA_URL, data.path());
    let client = reqwest::Client::new();

    let b = serde_json::to_vec(&data).expect("could not serialise within worker/chat");
    let res = match client.post(url).body(b).send().await {
        Ok(r) => r,
        Err(e) => {
            eprintln!("{}", e);
            return e.to_string();
        }
    };
    res.text().await.expect("worker text")
}

async fn generate(Json(data): Json<GenerateRequest>) -> impl IntoResponse {
    let url = format!("{}/api/{}", OLLAMA_URL, data.path());
    let client = reqwest::Client::new();

    let b = serde_json::to_vec(&data).expect("could not serialise within worker/generate");
    let res = match client.post(url).body(b).send().await {
        Ok(r) => r,
        Err(e) => {
            eprintln!("xxx {}", e);
            return e.to_string();
        }
    };
    res.text().await.unwrap()
}

async fn healthcheck() -> impl IntoResponse {
    let body = GenerateRequest {
        model: "gemma2".to_string(),
        prompt:
            "This is a health check of our systems. If you are OK, say it in a funny way (and making references to popupar culture) in no more than 5 words."
                .to_string(),
        ..Default::default()
    };
    generate(Json(body)).await
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
