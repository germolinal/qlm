use std::fmt::Debug;

use crate::ollama::{chat::ChatRequest, common::Ollamable, generate::GenerateRequest};

#[derive(Debug, Clone)]
pub enum LLMRequest {
    Generate(GenerateRequest),
    Chat(ChatRequest),
}

impl LLMRequest {
    pub fn path(&self) -> &'static str {
        match self {
            Self::Generate(r) => r.path(),
            Self::Chat(r) => r.path(),
        }
    }

    pub fn webhook(&self) -> Option<String> {
        match self {
            Self::Generate(r) => r.webhook.clone(),
            Self::Chat(r) => r.webhook.clone(),
        }
    }
}

impl std::convert::From<GenerateRequest> for LLMRequest {
    fn from(item: GenerateRequest) -> Self {
        LLMRequest::Generate(item)
    }
}
impl std::convert::From<ChatRequest> for LLMRequest {
    fn from(item: ChatRequest) -> Self {
        LLMRequest::Chat(item)
    }
}
