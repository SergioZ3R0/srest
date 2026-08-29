# srest

`srest` is a **TUI** (terminal user interface) client for interacting remotely
with the **Slurm REST API** (the HPC workload manager).

It lets Slurm administrators and users query cluster and job state without
SSH-ing into the master node, using only HTTP requests to `slurmrestd`.

## Status

`srest` is under active development. Current features:

- [x] HTTP client with JWT authentication (`X-SLURM-USER-TOKEN`, `X-SLURM-USER-NAME`).
- [x] Auto-detection of the `data_parser` version (v0.0.40 – v0.0.45) or a version pinned via configuration.
- [x] Version-gating and reporting of `warnings`/`errors` returned by slurmrestd.
- [x] Dashboard with real cluster stats (nodes up/down, jobs by state, partitions, accounts).
- [x] Real-time views: Jobs, Nodes, Partitions (with per-item detail panels).
- [x] Table search/filter (`/`).
- [x] Query builder/inspector: a Postman-style composer with endpoint-aware parameters, options gathered from the cluster (partitions, accounts) and a request history.
- [ ] Job submission and job actions (cancel/requeue).

## Features

**Tabs** (navigate with `tab`/`shift+tab` or `[`/`]`, `esc` returns to Dashboard):

- **Dashboard** — real cluster overview: nodes up/down, jobs running/pending/completed/failed, partitions and accounts.
- **Jobs** — your jobs (slurmrestd filters by the authenticated user), with a detail panel (account, partition, time limit, run time, assigned nodes, log paths).
- **Nodes** — cluster nodes with state, CPUs, memory and partitions; select a node to see its detail.
- **Partitions** — partition list with configured/total nodes.
- **Query** — a request composer: pick an endpoint (`1-5`), fill or cycle parameters (`enter` to edit, `←`/`→` to cycle options, `del` to clear), and run (`r`). The response and request history are shown alongside. Partition and account options are gathered live from the cluster.

**Key bindings**

| Key | Action |
| --- | ------ |
| `q` / `Ctrl+C` | quit |
| `tab` / `]`, `shift+tab` / `[` | next / previous tab |
| `esc` | go to Dashboard |
| `/` | filter the current table |
| `r` | refresh (Jobs tab) |
| `?` | toggle help |

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
