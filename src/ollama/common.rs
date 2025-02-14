use std::collections::HashMap;
use std::fmt::Debug;

use serde::{Deserialize, Serialize};
use serde_json::Value;

use crate::orchestrator::request::LLMRequest;

/// General options for Language Models. Not all
/// will be supported in all models.
#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct ModelOptions {
    temperature: f32,
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum ToolType {
    #[default]
    Function,
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct Tool {
    #[serde(rename = "type")]
    class: ToolType,

    description: String,
    parameters: Value,
}

pub type Duration = u64;
pub type CreatedAt = String;
pub type Count = u32;

pub trait Ollamable: Debug + Clone + Serialize + Default + Into<LLMRequest> {
    fn set_model<T: Into<String>>(&mut self, model: T);
    fn webhook(&self) -> &Option<String>;
    fn path(&self) -> &'static str;
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(tag = "type", rename_all = "lowercase")]
pub enum ReturnSchema {
    Object {
        properties: HashMap<String, Box<ReturnSchema>>,
        required: Option<Vec<String>>,
    },
    Integer,
    Boolean,
    String,
    Array {
        items: Box<ReturnSchema>,
    },
    // Add other types as needed...
}
