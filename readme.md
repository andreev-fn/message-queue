# Task queue
Currently work in progress.

A simple, reliable, and modern task queue built on top of PostgreSQL.

Also serves as a **real-world example of applying Domain-Driven Design (DDD) principles in Go**.

## TODO
- ☑ Archival of finalized tasks
- ☑ Task prioritization
- ☐ Make task types configurable (config file)
  - ☐ Configurable retries and timeouts and archive retention period
- ☐ Add authentication (config files, users.yaml, and tokens.yaml)
- ☐ Tries history (and attempt ID)
- ☐ Configurable retries strategies
- ☐ Removal of old tasks
- ☐ Metrics
- ☐ Implement webhooks
- ☐ Some kind of long-polling for workers
- ☐ Rate-limited queues
- ☐ Batch operations
- ☐ Single-process mode
- ☐ gRPC
- ☐ ValueObjects for kind, priority
- ☐ Decrease priority after some time
- ☐ Distributed tracing support
- ☐ Document
- ☐ Compare with alternatives
- ☐ Improve benchmark (many completed and many in READY status)
- ☐ Create admin API and UI (most likely won't do)

## ✨ Features

- ✅ **Task Prioritization** – High-priority tasks are picked first.
- ✅ **Retries & Backoff** – Failed tasks are retried automatically.

## 📦 When to Use

This queue is perfect when you:

- Want infrastructure with **minimal setup**
- Need a hackable solution to implement **custom features**
- Prefer to **avoid Kafka, RabbitMQ, etc.** for basic workflows

## 🚫 When *Not* to Use

- Your system processes **millions of events per second**

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
