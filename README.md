# QLM (Queue Language Model)

A privacy-focused and async LLM processing framework

**⚠️ Warning: This project is in very early stages and is not production-ready. Security is a top priority, but has not yet been implemented. (Contributions welcome) ⚠️**

`QLM` is an open-source project designed to enable privacy-sensitive tasks using small Language Models (or just LMs) on your own infrastructure. We leverage **RabbitMQ** for asynchronous task queuing and **Ollama** for running LMs. This allows you to process data without sending it to external cloud services, ensuring maximum privacy and control.

## The Need for Privacy

Wouldn't it be great to be able to use Language Models for tasks that are deeply personal
or sensitive to you? Like, with your company's or your own data? I think it would,
but of course this can be unwise in many situations. `QLM` aims to fix this: we believe
that relatively small language models can do a lot!

## Example Use Cases

- **Email Analysis:**
  - Check outgoing emails for tone, appropriateness, and potential information leaks.
  - Scan incoming emails for suspicious links or phishing attempts.
- **Code Review:**
  - Analyze code before committing to Git for potential security vulnerabilities (e.g., exposed API keys).
  - Generate documentation from the code.
- **Data Sanitization:**
  - Replicate databases with anonymized data for development and testing.
  - Redact PII from documents.
- **Document Summarization:**
  - Summarize long legal documents or reports.

## Key Features

- **Fully private:** Process everythong on your own infrastructure.
- **Asynchronous Processing:** Your infrastructure does not scale as much as the cloud, making queues a crucial element of this solution.
- **Ollama Integration:** Leverage the power of Ollama and its growing ecosystem of on-premises LMs.
- **RabbitMQ Queues:** Reliable and scalable task queuing.
- **Simple API:** Use the same request format as Ollama, with added `webhook` and `id` fields.

## Architecture

1.  **API Endpoints servers** (i.e., `orchestrators`) receive requests in the exact same format as the very mature Ollama's API, just adding a `webhook` URL and an `id`.
2.  API endpoint serves qnqueue incoming requests into **RabbitMQ Queues**
3.  **Worker VMs** listen to the RabbitMQ queues, fetching requests and forwarding them to Ollama.
4.  Once **Ollama** is done processing the requests, the workers will `POST` the response to the provided `webhook` URL, including the original `id`.

## Get it to run on your machine

1. **[Install and run Ollama](https://github.com/ollama/ollama)** using `ollama serve` or the desktop application. The worker will make requests to it. Make sure the `model` you use in the requests to the orchestrators (e.g., `gemma3`) is downloaded and loaded into Ollama.

Then there are two options

### Using Docker and Docker-compose (recommended for quick starts)

2. **Run `docker-compose build && docker-compose up`** - This will run the three main elements of the system: the Orchestrator, Worker, and Rabbit (Ollama is assumed to be already running), and it will also run the playground. Go to:
   - `http://localhost:3000` to check the playground, where you can test some stuff
   - `http://localhost:15672` to see Rabbit dashboard (password and username are `guest`)

### Using Go

2. Install `Docker` (or `Podman` I guess)
3. Have a look at the `Makefile`, as it might give you an idea of what is going on.
4. Run the components using the following commands (in different terminals):
   - `make rabbit` will start the queue that the orchestrator and the worker need to listen to. This needs to be running for the rest to work.
   - `make orchestrator` will run the orchestrator **without Docker** (there are also commands to build and run the container in the `Makefile`)
   - `make worker` runs the worker with no container (I am still thinking about how to pack this and Ollama together better.)
   - `make playground` (available at `http://localhost:3000`).

> Ollama and the worker are tightly coupled, and they need to agree on how many requests they will handle simultaneously. This can be done by passing a a single environment variable called `CONCURRENCY` to the worker and one called `OLLAMA_NUM_PARALLEL` to Ollama. (e.g., `CONCURRENCY=2 go run ./worker.go` and `OLLAMA_NUM_PARALLEL=N ollama serve`)

## Making a request

If you sent this in the request:

```shell
scurl -X POST http://localhost:8080/api/generate -d  '{
  "model": "gemma2",
  "prompt": "Translate this into english: Mi nombre es Rodrigo, y me gusta mucho el lenguaje",
  "webhook": "https://your-webhook-url.com/receive",
  "id": "12345"
}'
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

- **Security**: Implement robust security measures, including authentication, authorization, and encryption.
- **Error Handling**: Improve error handling and retry mechanisms, especially for webhook deliveries.
- **Resource Management**: Add dynamic worker scaling and resource monitoring.
- **Customization**: Provide hooks or plugins for custom pre- and post-processing.
  Monitoring and Logging: Add comprehensive monitoring and logging capabilities.
- **Testing**: Implement thorough unit and integration tests.
- **Documentation**: Improve documentation and provide more examples.
- **Support for other LLM servers**: Explore supporting other LLM serving frameworks.

## Contributing

We welcome contributions! Please feel free to submit pull requests or open issues on GitHub.

## License

[MIT License]

Disclaimer: This project is in its early stages and is provided "as is" without any warranty. Use at your own risk.
