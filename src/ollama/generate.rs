use serde::{Deserialize, Serialize};

use super::common::{Count, CreatedAt, Duration, ModelOptions, Ollamable, ReturnSchema};

/// Used with `/api/generate` request to Ollama
#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct GenerateRequest {
    pub webhook: Option<String>,
    // Parameters (according to Ollama's docs)
    /// The name of the model to use
    pub model: String,
    /// the prompt to generate a response for
    pub prompt: String,
    /// the text after the model response (Seems to be useful for code completion? e.g., 'suffix=return result')
    #[serde(skip_serializing_if = "Option::is_none")]
    pub suffix: Option<String>,
    /// a list of base64-encoded images (for multimodal models such as llava)
    #[serde(default)]
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub images: Vec<String>,

    // Advanced parameters (according to Ollama's docs)
    /// the format to return a response in. Format can be json or a JSON schema
    #[serde(skip_serializing_if = "Option::is_none")]
    pub format: Option<ReturnSchema>,
    /// additional model parameters listed in the documentation for the Modelfile such as temperature
    #[serde(skip_serializing_if = "Option::is_none")]
    pub options: Option<ModelOptions>,
    /// system message to (overrides what is defined in the Modelfile)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub system: Option<String>,
    /// the prompt template to use (overrides what is defined in the Modelfile)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub template: Option<String>,
    /// if false the response will be returned as a single response object, rather than a stream of objects
    /// default in ollama is True
    #[serde(default)]
    pub stream: bool,
    /// if true no formatting will be applied to the prompt. You may choose to use the raw parameter if you are specifying a full templated prompt in your request to the API
    #[serde(skip_serializing_if = "Option::is_none")]
    pub raw: Option<bool>,
    /// controls how long the model will stay loaded into
    /// memory following the request (default: 5m)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub keep_alive: Option<u8>,
}
impl Ollamable for GenerateRequest {
    fn set_model<T: Into<String>>(&mut self, model: T) {
        self.model = model.into()
    }
    fn webhook(&self) -> &Option<String> {
        &self.webhook
    }
    fn path(&self) -> &'static str {
        "generate"
    }
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct GenerateResponse {
    pub model: String,
    pub created_at: CreatedAt,
    pub response: String,
    pub done: bool,
    pub total_duration: Duration,
    pub load_duration: Duration,
    pub prompt_eval_count: Count,
    pub prompt_eval_duration: Duration,
    pub eval_count: Count,
    pub eval_duration: Duration,
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct StreamedGenerateResponse {
    pub model: String,
    pub created_at: CreatedAt,
    pub response: String,
    pub done: bool,
}
