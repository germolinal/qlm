use futures::lock::Mutex;

use axum::http::StatusCode;
use serde::{Deserialize, Serialize};
use std::{collections::VecDeque, fmt::Debug, sync::Arc, usize};

use tokio::sync::mpsc::{Receiver, Sender};

use crate::ollama::{
    chat::{ChatReponse, ChatRequest},
    generate::{GenerateRequest, GenerateResponse},
};

use super::{request::LLMRequest, response::LLMHandlerResponse, worker::Worker};

#[derive(Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct OrchestratorState {
    pub model: String,
    pub workers: Vec<Worker>,
    #[serde(skip)]
    #[serde(default)]
    pub queue: VecDeque<LLMRequest>,
    #[serde(skip)]
    #[serde(default)]
    pub sender: Option<Sender<bool>>,
}

pub type SharedState = Arc<Mutex<OrchestratorState>>;

impl Default for OrchestratorState {
    fn default() -> Self {
        Self {
            model: String::new(),
            workers: vec![],
            queue: VecDeque::with_capacity(100),
            sender: None,
        }
    }
}

impl OrchestratorState {
    async fn start_consuming_queue(&self) -> LLMHandlerResponse<&'static str> {
        if let Some(sender) = &self.sender {
            if let Err(e) = sender.send(true).await {
                LLMHandlerResponse::Err((StatusCode::INTERNAL_SERVER_ERROR, format!("{}", e)))
            } else {
                LLMHandlerResponse::Ok((StatusCode::OK, "queue will be consumed"))
            }
        } else {
            LLMHandlerResponse::Err((
                StatusCode::INTERNAL_SERVER_ERROR,
                "app state has no channels".to_string(),
            ))
        }
    }

    pub fn get_available_worker(&mut self) -> Option<usize> {
        for (i, w) in self.workers.iter().enumerate() {
            let current = w.current_load.load(std::sync::atomic::Ordering::Relaxed);
            if w.concurrency > current {
                return Some(i);
            }
        }
        None
    }

    pub async fn process_queue(mut rx: Receiver<bool>, state: SharedState) {
        while let Some(_) = rx.recv().await {
            loop {
                // Get data.
                if let Some((url, index, req)) = OrchestratorState::pop_task(state.clone()).await {
                    match req {
                        LLMRequest::Generate(v) => {
                            Worker::process_queued::<GenerateRequest, GenerateResponse>(
                                state.clone(),
                                index,
                                url,
                                v,
                            )
                            .await
                        }
                        LLMRequest::Chat(v) => {
                            Worker::process_queued::<ChatRequest, ChatReponse>(
                                state.clone(),
                                index,
                                url,
                                v,
                            )
                            .await
                        }
                    };
                } else {
                    break;
                }
            }
        }
    }

    pub async fn enqueue(&mut self, req: LLMRequest) -> LLMHandlerResponse<&'static str> {
        self.queue.push_back(req);
        if let LLMHandlerResponse::Err(e) = self.start_consuming_queue().await {
            return LLMHandlerResponse::Err(e);
        }
        LLMHandlerResponse::Ok((StatusCode::ACCEPTED, "your request has been queued"))
    }

    pub async fn workload_done(
        state: SharedState,
        index: usize,
    ) -> Result<(), (StatusCode, String)> {
        let mut guard = state.lock().await;
        if let Some(w) = guard.workers.get_mut(index) {
            w.current_load
                .fetch_sub(1, std::sync::atomic::Ordering::Relaxed);
            guard.start_consuming_queue().await;
            Ok(())
        } else {
            Err((
                StatusCode::INTERNAL_SERVER_ERROR,
                format!("there are no {} workers", index),
            ))
        }
    }
    async fn pop_task(state: SharedState) -> Option<(String, usize, LLMRequest)> {
        let mut guard = state.lock().await;
        // No queued msgs? no tasks
        let w_index = guard.get_available_worker();
        if guard.queue.is_empty() || w_index.is_none() {
            // no queues or no workers available... return
            return None;
        }

        let w_index = w_index.expect("this was checked earlier");
        let w = &guard.workers[w_index];
        // Annotate it
        w.current_load
            .fetch_add(1, std::sync::atomic::Ordering::Relaxed);
        // The the URL
        let url = w.address.clone();
        // Get the msg
        let msg = guard
            .queue
            .pop_front()
            .expect("this is a bug... we checked that queue was not empty");
        // Return data
        return Some((url, w_index, msg));
    }
}
