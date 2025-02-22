use std::{fmt::Debug, sync::atomic::AtomicU16};

use axum::http::StatusCode;
use serde::{de::DeserializeOwned, Deserialize, Serialize};

use crate::{ollama::common::Ollamable, orchestrator::state::OrchestratorState};

use super::{response::LLMHandlerResponse, state::SharedState};

#[derive(Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct Worker {
    pub address: String,
    pub concurrency: u16,
    #[serde(skip)]
    pub current_load: AtomicU16,
}

impl Worker {
    pub async fn process_queued<I: Ollamable, O: Serialize + DeserializeOwned>(
        state: SharedState,
        index: usize,
        url: String,
        mut req: I,
    ) {
        // Do the work
        let client = reqwest::Client::new();
        let worker_url = format!("{}/{}", url, req.path());

        let guard = state.lock().await;
        req.set_model(guard.model.clone());
        drop(guard);

        if let Some(hook) = &req.webhook() {
            let res = client.post(&worker_url).json(&req).send().await;
            let r = match res {
                Err(e) => {
                    eprintln!("could not reach worker at address {}: {}", worker_url, e);
                    return;
                }
                Ok(v) => {
                    let r: O = v.json().await.unwrap();
                    r
                }
            };
            let res = client.post(hook).json(&r).send().await;

            // Unload the work.
            OrchestratorState::workload_done(state, index)
                .await
                .unwrap();

            if let Err(e) = res {
                eprintln!("could not reach webhook at address {}: {}", hook, e)
            }
        } else {
            eprintln!("a webhook url is required")
        }
    }
    pub async fn process<I: Ollamable, O: Serialize + DeserializeOwned>(
        state: SharedState,
        index: usize,
        url: String,
        mut req: I,
    ) -> LLMHandlerResponse<O> {
        // Setup
        let guard = state.lock().await;
        req.set_model(guard.model.clone());
        drop(guard);

        // Do the work
        let client = reqwest::Client::new();
        let worker_url = format!("{}/{}", url, req.path());
        let res = client.post(worker_url).json(&req).send().await;
        let r = match res {
            Err(e) => return LLMHandlerResponse::Err((StatusCode::BAD_REQUEST, format!("{}", e))),
            Ok(v) => {
                let r: O = v.json().await.unwrap();
                r
            }
        };
        // Unload the work.
        OrchestratorState::workload_done(state, index)
            .await
            .unwrap();
        LLMHandlerResponse::Ok((StatusCode::OK, r))
    }
}
