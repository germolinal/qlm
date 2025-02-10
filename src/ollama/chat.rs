use serde::{Deserialize, Serialize};
use serde_json::Value;

use super::common::{Count, CreatedAt, Duration, ModelOptions, Ollamable, ReturnSchema};

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Role {
    /// A message sent by the user
    #[default]
    User,
    /// A messare returned by the assistant
    Assistant,
    /// A system prompt
    System,
    /// Unsure
    Tool,
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct Message {
    role: Role,
    content: String,
    /// a list of base64-encoded images (for multimodal models such as llava)
    #[serde(default)]
    #[serde(skip_serializing_if = "Vec::is_empty")]
    images: Vec<String>,
    // tool_calls: Option<string>,
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct ChatRequest {
    pub webhook: Option<String>,

    /// The model to call
    model: String,
    /// The history of messages, to keep a chat memory
    pub messages: Vec<Message>,
    /// List of tools in JSON for the model to use, if supported
    #[serde(skip_serializing_if = "Option::is_none")]
    tools: Option<Value>,
    /// the format to return a response in. Format can be `json` or a JSON schema.
    #[serde(skip_serializing_if = "Option::is_none")]
    format: Option<ReturnSchema>,
    /// additional model parameters listed in the
    /// documentation for the Modelfile such as
    /// temperature
    #[serde(skip_serializing_if = "Option::is_none")]
    options: Option<ModelOptions>,
    /// if `false` the response will be returned as a
    /// single response object, rather than a stream
    /// of objects
    #[serde(default)]
    stream: bool,
    /// controls how long the model will stay loaded into
    /// memory following the request (default: 5m)
    #[serde(skip_serializing_if = "Option::is_none")]
    keep_alive: Option<u8>,
}

impl Ollamable for ChatRequest {
    fn set_model<T: Into<String>>(&mut self, model: T) {
        self.model = model.into()
    }

    fn webhook(&self) -> &Option<String> {
        &self.webhook
    }
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct ChatReponse {
    model: String,
    created_at: CreatedAt,
    #[serde(skip_serializing_if = "Option::is_none")]
    message: Option<Message>,
    done: bool,
    total_duration: Duration,
    load_duration: Duration,
    prompt_eval_count: Count,
    prompt_eval_duration: Duration,
    eval_count: Count,
    eval_duration: Duration,
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct StreamedChatReponse {
    model: String,
    created_at: CreatedAt,
    message: Message,
    done: bool,
}

#[cfg(test)]
mod tests {
    // use super::*;

    #[test]
    fn serialize() {
        let a: bool = Default::default();
        println!("{} -- {}", a, u64::MAX)
    }
}
