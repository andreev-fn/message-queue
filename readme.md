# Message queue
Currently work in progress.

A simple, reliable, and modern message queue built on top of PostgreSQL.

Also serves as a **real-world example of applying Domain-Driven Design (DDD) principles in Go**.

## TODO
- ☑ Archival of finalized messages
- ☑ Message prioritization
- ☑ Some kind of long-polling for consumers
- ☑ Atomic Ack + Publish
- ☑ Permanent nack
- ☑ Batch operations
- ☐ Configurable queues (config file)
- ☐ Dead-letter queues
- ☐ Add authentication (config file)
- ☐ Message forwarding
- ☐ Tries history (and attempt ID)
- ☐ Configurable retries strategies
- ☐ Removal of old messages
- ☐ Metrics
- ☐ Implement webhooks
- ☐ Rate-limited queues
- ☐ Single-process mode
- ☐ gRPC
- ☐ ValueObjects for queue name, priority
- ☐ Decrease priority after some time
- ☐ Distributed tracing support
- ☐ Document
- ☐ Compare with alternatives
- ☐ Improve benchmark (many completed and many in READY status)
- ☐ Create admin API and UI (most likely won't do)

## ✨ Features

- ✅ **Message Prioritization** – High-priority messages are picked first.
- ✅ **Retries & Backoff** – Failed messages are retried automatically.
- ✅ **Long-polling** – Consumers can efficiently wait until a message comes.
- ✅ **Atomic Ack + Publish** – Consumers can publish messages atomically with Ack.

## 📦 When to Use

This queue is perfect when you:

- Want infrastructure with **minimal setup**
- Need a hackable solution to implement **custom features**
- Prefer to **avoid Kafka, RabbitMQ, etc.** for basic workflows

## 🚫 When *Not* to Use

- Your system processes **millions of messages per second**

## 👷‍♂️ Contributing

This project is designed to be **easy to understand and modify**.

- Focus on simplicity over feature-bloat
- Codebase is clean, small, and well-documented
- PRs for bug fixes and clarity are welcome

⚠️ **Please note**: I may not accept all feature requests or PRs, as the goal is to keep this project minimal and focused.  
If you have an idea, please open an issue first to discuss it before submitting a pull request.

## 📄 License

This project is licensed under the [MIT License](LICENSE). It allows you to:

- Use, copy, modify, merge, publish, and distribute the code freely
- Use it in commercial and closed-source software
