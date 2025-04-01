# QLM (Queue Language Model) - Privacy-Focused On-Premises LLM Processing

**⚠️ Warning: This project is in very early stages and is not production-ready. Security is a top priority, but has not yet been implemented. (Contributions welcome) ⚠️**

`QLM` is an open-source project designed to enable privacy-sensitive tasks using small Language Models (or just LMs) on your own infrastructure. We leverage RabbitMQ for asynchronous task queuing and Ollama for running LMs. This allows you to process data without sending it to external cloud services, ensuring maximum privacy and control.

## The Need for Privacy

In today's data-driven world, many tasks require the power of language models but involve sensitive information. Sending this data to third-party cloud services raises significant privacy concerns. QLM addresses this by providing a secure, on-premises solution.

## Example Use Cases

* **Email Analysis:**
    * Check outgoing emails for tone, appropriateness, and potential information leaks.
    * Scan incoming emails for suspicious links or phishing attempts.
* **Code Review:**
    * Analyze code before committing to Git for potential security vulnerabilities (e.g., exposed API keys).
    * Generate documentation from the code.
* **Data Sanitization:**
    * Replicate databases with anonymized data for development and testing.
    * Redact PII from documents.
* **Document Summarization:**
    * Summarize long legal documents or reports.

## Key Features

* **Privacy-First:** Process sensitive data on your own infrastructure.
* **Asynchronous Processing:** Handle batch processing efficiently.
* **Ollama Integration:** Leverage the power of Ollama and its growing ecosystem of on-premises LMs.
* **RabbitMQ Queues:** Reliable and scalable task queuing.
* **Simple API:** Use the same request format as Ollama, with added `webhook` and `id` fields.


## Architecture

1.  **API Endpoints servers** receive requests in the exact same format as the very mature Ollama's API, just adding a `webhook` URL and an `id`.
2.  API endpoint serves qnqueue incoming requests into **RabbitMQ Queues** 
3.  **Worker VMs** listen to the RabbitMQ queues, fetching requests and forwarding them to Ollama.
4.  Once **Ollama** is done processing the requests, the workers will `POST` the response to the provided `webhook` URL, including the original `id`.


## Getting Started

For now, read the `Makefile` and you will get a clear idea. 

> You will need Docker to run Rabbit.

You need to run:
* `make rabbit`
* `make orchestrator`
* `make worker`
* OPTIONAL: `make playground` (available at `http://localhost:3000`).

## Example 

If you sent this in the request:

```json
{
  "model": "gemma3",
  "prompt": "Translate this into english: Mi nombre es Rodrigo, y me gusta mucho el lenguaje",
  "webhook": "https://your-webhook-url.com/receive",
  "id": "12345"
}
```

Your webook will receive something like this:

```json
{  
  "created_at": "2025-04-01T08:18:34.61715Z",
  "done": true,
  "done_reason": "stop",
  "eval_count": 30,
  "eval_duration": 1194330583,
  "id": "12345",
  "load_duration": 58733583,
  "model": "gemma3",
  "prompt_eval_count": 25,
  "prompt_eval_duration": 799852875,
  "response": "My name is Rodrigo, and I really like language.  \n\n\nLet me know if you have any other phrases you'd like translated! 😊",
  "total_duration": 2053661125
}
```
> **Note**: you can ask for structured output

## FAQ

**Is `QLM` production-ready?**

No, `QLM` is in very early stages and is not production-ready. Security is a major concern that has not yet been addressed.

**What are the security considerations?**

Currently, there are no security measures in place. This is a top priority for future development.

**How does `QLM` handle scalability?**

`QLM` is designed for on-premises use and relies on RabbitMQ for asynchronous processing. Scalability can be improved by increasing worker concurrency or using a Kubernetes cluster. (We would be very happy if you helped us streamline and document how to deploy `QML` in an on-premises Kubernetes cluster)

**Can I use different language models?**

Yes, `QLM` leverages Ollama, which supports a wide range of language models.

**How do I handle errors?**

RabbitMQ handles unacknowledged messages. Webhook error handling is still under development.

**Can I contribute to `QLM`?**

Absolutely! We welcome contributions. Please submit pull requests or open issues on GitHub.

## Future Development
* **Security**: Implement robust security measures, including authentication, authorization, and encryption.
* **Error Handling**: Improve error handling and retry mechanisms, especially for webhook deliveries.
* **Resource Management**: Add dynamic worker scaling and resource monitoring.
* **Customization**: Provide hooks or plugins for custom pre- and post-processing.
Monitoring and Logging: Add comprehensive monitoring and logging capabilities.
* **Testing**: Implement thorough unit and integration tests.
* **Documentation**: Improve documentation and provide more examples.
* **Support for other LLM servers**: Explore supporting other LLM serving frameworks.

## Contributing

We welcome contributions! Please feel free to submit pull requests or open issues on GitHub.

## License
[MIT License]

Disclaimer: This project is in its early stages and is provided "as is" without any warranty. Use at your own risk.