<div align="center">
  <img src="https://raw.githubusercontent.com/SergioZ3R0/srest/main/docs/srest.svg" alt="srest" width="180">
</div>

<h1 align="center">srest</h1>
<p align="center">
  <b> TUI client for the Slurm REST API</b><br>
  <i> View jobs, nodes, partitions &bull; Build requests from your terminal </i><br>
  <a href="https://srest.scszero.com/">srest.scszero.com</a> &nbsp;|&nbsp;
  <a href="https://srest.scszero.com/docs.html">Documentation</a>
</p>

`srest` is a **TUI** (terminal user interface) client for interacting remotely
with the **Slurm REST API** (the HPC workload manager).

It lets Slurm administrators and users query cluster and job state without
SSH-ing into the master node, using only HTTP requests to `slurmrestd`.

## About

**srest** is a terminal user interface (TUI) built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) that turns the Slurm
REST API into a live, interactive dashboard. It lets you:

- Monitor your **jobs**, cluster **nodes** and **partitions** at a glance.
- Inspect job details — account, partition, time limits, run time, assigned
  nodes and log paths — without SSH.
- **Compose API requests** visually (endpoint-aware parameters with options
  gathered live from the cluster: state, account, partition, qos) and
  inspect the request history.
- **Custom query** panel to write or paste any request path directly.
- Detect and adapt to the `slurmrestd` **API version** automatically
  (v0.0.40 – v0.0.45).

It is written in **Go**, styled with
[Lip Gloss](https://github.com/charmbracelet/lipgloss), and speaks directly to
`slurmrestd` using JSON Web Token authentication — no SSH, no node login.

## Status

`srest` is under active development. Current features:

- [x] HTTP client with JWT authentication (`X-SLURM-USER-TOKEN`, `X-SLURM-USER-NAME`).
- [x] Auto-detection of the `data_parser` version (v0.0.40 – v0.0.45) or a version pinned via configuration.
- [x] Version-gating and reporting of `warnings`/`errors` returned by slurmrestd.
- [x] Dashboard with real cluster stats (nodes up/down, jobs by state, partitions, accounts).
- [x] Real-time views: Jobs, Nodes, Partitions (with per-item detail panels).
- [x] Table search/filter (`/`).
- [x] Query builder with user-focused endpoints (ping, get jobs, submit jobs) and cluster-gathered options (state, account, partition, qos, gres).
- [x] Custom query panel for writing/pasting any request path.
- [x] Job submission with script path, $EDITOR support, and gathered partition/account/qos/gres options.
- [x] Persistent request history (saved to `~/.local/share/srest/history.json`, max 100 entries).
- [ ] Job cancellation and requeue.

## Features

**Tabs** (navigate with `tab`/`shift+tab` or `[`/`]`, `esc` returns to Dashboard):

- **Dashboard** — real cluster overview: nodes up/down, jobs running/pending/completed/failed, partitions and accounts.
- **Jobs** — your jobs (slurmrestd filters by the authenticated user), with a detail panel (account, partition, time limit, run time, assigned nodes, log paths, exit code). Press `j` in Partitions to view jobs by partition.
- **Nodes** — cluster nodes with state, CPUs, memory and partitions; select a node to see its detail.
- **Partitions** — partition list with configured/total nodes and max wall time. Press `j` to filter jobs by partition.
- **Query** — a request composer with three user-focused endpoints:
  - **ping** — connectivity check.
  - **get jobs** — query with filters: state, account, partition, qos, node, users. Account, partition and qos options are gathered live from the cluster.
  - **submit jobs** — submit a job with name, partition, qos, account, gres, wall time, nodes, cpus/task, memory, script. Partition, account, qos, and gres options are gathered from the cluster. Press `e` to open `$EDITOR`.
  - **Custom query** panel: type or paste any request path to run it directly.
  - **Request history** — every request logged with status, duration and warnings. Persisted across sessions (max 100 entries).

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
