# Message queue
Currently work in progress.

A simple, reliable, and modern message queue built on top of PostgreSQL.

Also serves as a **real-world example of applying Domain-Driven Design (DDD) principles in Go**.

## TODO
- [x] Archival of finalized messages
- [x] Message prioritization
- [x] Some kind of long-polling for consumers
- [x] Atomic Ack + Publish
- [x] Permanent nack
- [x] Batch operations
- [x] Configurable queues (config file)
- [x] Message redirection
- [x] Dead-letter queues
- [x] Single-process mode
- [ ] Processing extension mechanism
- [ ] Add authentication (config file)
- [ ] Attempt IDs (like delivery tags)
- [ ] Removal of old messages
- [ ] Metrics
- [ ] Implement webhooks
- [ ] Rate-limited queues
- [ ] gRPC
- [ ] ValueObjects for ~~queue name~~, priority
- [ ] Decrease priority after some time
- [ ] Distributed tracing support
- [ ] Document
- [ ] Compare with alternatives
- [ ] More benchmarks
- [ ] Create admin API and UI (most likely won't do)

## ✨ Features

- ✅ **Message Prioritization** – High-priority messages are picked first.
- ✅ **Retries & Backoff** – Lost/NACKed messages are retried automatically.
- ✅ **Dead Letter Queues** – Failed messages are automatically routed for later inspection.
- ✅ **Efficient Long-Polling** – Consumers wait for messages without busy-looping.
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
