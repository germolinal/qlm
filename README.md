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
response**, and thus you need to provide a `hook` parameter with your calls, indicating
where should we send the response from the queue.

> **Note:** Of course, you need to set up your client to handle these async operations
> properly; e.g., you send a text message that will receive no answer immediately. On the 
> contrary, it will receive an answer later, sent to the `hook` url you provided.