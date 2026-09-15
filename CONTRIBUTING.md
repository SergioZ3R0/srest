# Contributing to srest

We welcome contributions to srest. This document outlines the standard process
for contributing to the project.

## Reporting Issues

Before opening a new issue, please check the existing open issues to avoid
duplicates.

When reporting a bug or requesting a feature, use the provided GitHub issue
templates. Ensure you include relevant details about your cluster and
environment, such as:

- OS distribution and version
- Slurm version (`scontrol show config | grep -i version`)
- Slurm API version served by `slurmrestd` (srest auto-detects it; you can
  check with the Query tab or `curl $SLURM_URL/slurm/v0.0.45/ping`)
- `slurmrestd` URL and authentication method (JWT / none)
- Execution environment (direct connection, Docker test lab, etc.)

## Local Development Setup

1. Fork the repository on GitHub.
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/srest.git
   cd srest
   ```
3. Build and run:
   ```bash
   go build ./...
   go run .
   ```
4. (Optional) Spin up a real Slurm cluster for testing — see
   [docs/test-lab.md](docs/test-lab.md).

## Pull Request Process

- **Branching**: Create a new branch for your changes
  (`git checkout -b feature/issue-name`), following conventional branch names
  (`feat/`, `fix/`, `docs/`, `chore/`).
- **Code Style**: Run `gofmt` and keep consistency with the existing codebase.
- **Linting**: Run the same checks as CI before submitting:
  ```bash
  gofmt -l .
  go vet ./...
  golangci-lint run ./...
  ```
- **Testing**: Verify your changes do not break existing functionality:
  ```bash
  go test ./...
  ```
  If you touched the API client, run the integration test against a live
  cluster (see `docs/test-lab.md`):
  ```bash
  SLURM_URL=http://localhost:6820 SLURM_JWT=<token> \
    go test ./internal/api -run TestPingIntegration -v
  ```
- **Commit Messages**: Write clear, concise, and descriptive conventional
  commit messages.
- **Submission**: Open a Pull Request against the `main` branch. Fill out the
  provided Pull Request template completely, linking any relevant issues
  (e.g., `Fixes #10`).

All Pull Requests will be reviewed by the maintainers before merging.