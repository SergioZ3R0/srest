# srest

`srest` is a **TUI** (terminal user interface) client for interacting remotely
with the **Slurm REST API** (the HPC workload manager).

It lets Slurm administrators and users query cluster and job state without
SSH-ing into the master node, using only HTTP requests to `slurmrestd`.

## Status

Currently in early development (MVP):

- [x] HTTP client with JWT authentication (`X-SLURM-USER-TOKEN`, `X-SLURM-USER-NAME`).
- [x] Auto-detection of the `data_parser` version (v0.0.40 – v0.0.45) or a version pinned via configuration.
- [x] Asynchronous connectivity check (`/ping`) using Bubble Tea.
- [x] Version-gating and reporting of `warnings`/`errors` returned by slurmrestd.
- [ ] Jobs, nodes and partitions views.

## Stack

- **Language:** Go 1.24+
- **TUI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)

## Architecture

Follows the standard Go layout with a strict separation of responsibilities:

```
.
├── main.go                  # Entry point: configuration + Bubble Tea startup
└── internal/
    ├── config/              # Configuration loading (environment)
    ├── api/                 # Pure HTTP client (no UI)
    └── ui/                  # Bubble Tea model, view and update
```

- `internal/config` loads configuration, prioritizing environment variables.
- `internal/api` is a pure HTTP client: no UI. It is consumed asynchronously
  via `tea.Cmd`.
- `internal/ui` consumes `internal/api` without blocking the interface.

## Requirements

- Go 1.24+ (to build from source).

## Installation and usage

```bash
go build -o srest .
```

### Configuration

`srest` is configured through environment variables:

| Variable            | Required | Description                                                        |
| ------------------- | -------- | ------------------------------------------------------------------ |
| `SLURM_URL`         | No       | Base URL of `slurmrestd`. Defaults to `http://localhost:6820`.     |
| `SLURM_JWT`         | Yes*     | JWT token for the `X-SLURM-USER-TOKEN` header.                     |
| `SLURM_USER_NAME`   | No       | User for `X-SLURM-USER-NAME`. Defaults to the current OS user.     |
| `SLURM_API_VERSION` | No       | API version to use (e.g. `v0.0.45`). If omitted, it is auto-detected. |

\* For endpoints requiring authentication. Can be omitted on clusters with a
JWT-less `slurmrestd`.

```bash
SLURM_URL=http://localhost:6820 \
SLURM_JWT=<token> \
SLURM_USER_NAME=slurm \
./srest
```

Press `q` (or `Ctrl+C`) to quit.

## Test Lab

To try `srest` against a **real** Slurm cluster (with `slurmctld`,
`slurmdbd` and `slurmrestd`) without setting up infrastructure, see:

- [docs/test-lab.md](docs/test-lab.md) — spin up a Slurm test cluster with Docker.

## Tests

```bash
go test ./...                  # unit tests (the integration one skips itself)
```

Integration test against a real `slurmrestd` (requires `SLURM_URL` and `SLURM_JWT`):

```bash
SLURM_URL=http://localhost:6820 SLURM_JWT=<token> \
  go test ./internal/api -run TestPingIntegration -v
```

## License

[Apache License 2.0](LICENSE).
