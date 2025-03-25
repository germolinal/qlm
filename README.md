# Queues for Language Models (QLM)

This project aims to provide a robust interface for running Language Models
on-premises. Obviously, some LLMs are too large to run on local machines
(which is why I dropped the first L, which stands for Large) which means that
their capacity to run complex tasks is limited. Nevertheless, Small Language
Models have become quite capable, and they are now—I think—ready to perform
tasks such as:

1. Performing basic proof-reading and rephrasing of text
2. Checking emails for tone/sensitive information before going out
3. Checking emails for malicious links before comming in
4. Checking whether messages sent to cloud-based Large Language Models contain data that might be better NOT to share.

## Tech stack

We use [RabbitMQ](https://www.rabbitmq.com) as a queueing system and [Ollama](https://ollama.com)
as the Local LLM runner. What we do in the middle is to create the queue and orchestrate the
distribution of messages between them.

## API

We simply leverage the preexistent `Ollama` API, so any call you were intending to make to
it, you can make to `QLM`. The only difference is that **you should not expect an immediate
response**. They will come back to you later, when they are processe. To make this happen, `QLM` asks for two extra params in any Ollama-ish request:

1.  a `webhook` parameter, indicating where the result should be posted after it is processed.
2.  an `id`, which is not used by `QLM`, but it will be useful for you to know which responses correspond to which messages

> **Note:** You need to set up your client to handle these async operations
> properly; e.g., you send a text message that will receive no answer immediately. On the
> contrary, it will receive an answer later, sent to the `hook` url you provided.

## FAQ

## TODO

Handle concurrency properly: 

* Ollama - https://github.com/ollama/ollama/blob/main/docs/faq.md#how-does-ollama-handle-concurrent-requests