# Contributing to GopherQueue 🚀

Thank you for your interest in contributing to **GopherQueue**! We welcome contributions from developers of all skill levels. Whether you are fixing a bug, improving documentation, adding unit tests, or proposing a major feature, your help is appreciated.

---

## 📌 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Submitting Pull Requests](#submitting-pull-requests)
- [Development Workflow](#development-workflow)
  - [Prerequisites](#prerequisites)
  - [Local Setup](#local-setup)
  - [Running Tests & Quality Gates](#running-tests--quality-gates)
- [Coding Standards & Conventions](#coding-standards--conventions)
- [License](#license)

---

## Code of Conduct

Please be respectful, professional, and empathetic in all interactions (GitHub Issues, Pull Requests, Discussions). We aim to foster an inclusive and welcoming environment for everyone.

---

## How Can I Contribute?

### Reporting Bugs

Before creating a new bug report, please check existing GitHub Issues to see if the issue has already been reported. When submitting a bug report:

- Use a clear and descriptive title.
- Provide step-by-step instructions to reproduce the bug.
- Include expected vs. actual behavior.
- Mention your environment details (Go version, OS, Redis version).

### Suggesting Features

Feature requests are welcome! When proposing a new feature:

- Explain the use case and why this feature would benefit GopherQueue users.
- Outline the proposed API signature or user interface.
- Discuss how it aligns with existing design principles (Zero Job Loss, Dual-Mode, Middleware Synergy).

### Submitting Pull Requests

1. **Fork the Repository**: Create a fork of `github.com/balramadan/gopherqueue` under your GitHub account.
2. **Create a Feature Branch**:
   ```bash
   git checkout -b feature/my-amazing-feature
   ```
3. **Write Clean, Idiomatic Go**:
   - Follow standard Go formatting (`gofmt`).
   - Add descriptive GoDoc comments for any exported types or functions.
   - Include comprehensive unit tests (`*_test.go`).
4. **Pass All Quality Gates**:
   ```bash
   make lint && make test-unit
   ```
5. **Open a Pull Request**: Submit your PR against the `main` branch with a detailed description of your changes.

---

## Development Workflow

### Prerequisites

- **Go**: `v1.22+` or `v1.24+`
- **Docker**: For running local Redis containers (`docker run -d -p 6379:6379 redis:alpine`)
- **Node.js & pnpm** *(Optional - only needed if modifying `web/` Vue GUI source)*: Node.js v22+, `pnpm`

### Local Setup

```bash
# 1. Clone your fork
git clone https://github.com/<your-username>/gopherqueue.git
cd gopherqueue

# 2. Download Go dependencies
make deps

# 3. Start local Redis instance
make run-redis
```

### Running Tests & Quality Gates

Run the test suite with the Go Race Detector enabled:

```bash
# Run unit tests and race detector
make test-unit

# Run static analysis
make lint

# Verify binary build
make build
```

---

## Coding Standards & Conventions

1. **Zero Job Loss Invariant**: Atomic Redis operations (`BLMOVE`, Lua scripts) must be maintained to ensure jobs in `PROCESSING` state are never lost.
2. **Error Handling**: Custom error types (like `ErrRateLimited`) must be handled properly without incrementing retry counters unexpectedly.
3. **Public API Stability**: Maintain idiomatic Go package layout (`client.go`, `server.go`, `job.go`, `middleware.go`, `options.go`).
4. **Documentation**: All exported symbols must include clear GoDoc comments in English.

---

## License

By contributing to GopherQueue, you agree that your contributions will be licensed under the project's [MIT License](LICENSE).
