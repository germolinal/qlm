pub mod ollama;
pub mod orchestrator;
pub mod worker;

/// Defines which port to use for the server.
///
/// If an environmental variable `PORT` is defined, that value
/// is used; otherwise, we default to `4000`.
#[tracing::instrument]
pub fn getport(default: u16) -> u16 {
    match std::env::var("PORT") {
        Ok(v) => v.parse().unwrap(),
        Err(_) => std::env::var("PORT")
            .ok()
            .and_then(|s| s.parse().ok())
            .unwrap_or(default),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn it_works() {
        assert!(true)
    }
}
