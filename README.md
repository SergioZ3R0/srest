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

https://github.com/user-attachments/assets/3a20799b-b075-4556-ad54-e5e0ca921562

## Philosophy

**srest is a pure REST API client.** It talks to `slurmrestd` and nothing
else — no SSH, no local SLURM commands, no filesystem access. Run it from
your laptop against any cluster with zero dependencies on the cluster's
tooling.

This makes srest:

- **Portable** — a single binary, no SLURM installation required.
- **Secure** — no shell access needed; only the REST API endpoint must be reachable.
- **Cluster-agnostic** — works against any `slurmrestd` version (v0.0.40–45) without modification.

If a feature isn't exposed by the Slurm REST API, srest doesn't attempt to
work around it. What you see is exactly what `slurmrestd` provides.

## Stack

- **Language:** Go 1.24+
- **TUI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)

## Architecture

Follows the standard Go layout with a strict separation of responsibilities:

```
.
├── main.go                  # Entry point: vault CLI + Bubble Tea startup
└── internal/
    ├── config/              # Configuration loading (env vars + encrypted vault)
    ├── api/                 # Pure HTTP client (no UI)
    └── ui/                  # Bubble Tea model, view and update
```

- `internal/config` loads configuration from environment variables or the encrypted vault (`~/.srest/config.vault`), prioritizing env vars.
- `internal/api` is a pure HTTP client: no UI. It is consumed asynchronously
  via `tea.Cmd`.
- `internal/ui` consumes `internal/api` without blocking the interface.

## Requirements

- Go 1.24+ (to build from source).

## Quick Start

### Download

```bash
curl -LO https://github.com/SergioZ3R0/srest/releases/latest/download/srest-linux-amd64.zip
unzip srest-linux-amd64.zip
chmod +x srest
```

### Configure

Point it at your `slurmrestd` and authenticate with a JWT:

```bash
SLURM_URL=http://localhost:6820 \
SLURM_JWT=$(scontrol token | cut -d= -f2) \
./srest
```

### Encrypted vault (recommended)

Store credentials encrypted with AES-256-GCM:

```bash
# Create encrypted vault
srest vault init

# Run (prompts for vault password)
./srest

# Or skip vault prompt with env var
SREST_VAULT_PASS=myscret ./srest
```

Vault commands:

| Command | Description |
|---------|-------------|
| `srest vault init` | Create new encrypted config |
| `srest vault encrypt` | Encrypt existing plain config |
| `srest vault decrypt` | Decrypt and display contents |

Vault password can be set via `SREST_VAULT_PASS` env var to skip the prompt.

### Run

```bash
./srest
```

## Features

- **Encrypted credential vault** — AES-256-GCM encrypted config file (`~/.srest/config.vault`) with PBKDF2 key derivation.
- **JWT authentication** — `X-SLURM-USER-TOKEN` and `X-SLURM-USER-NAME` headers, plus Bearer token support for proxy setups.
- **Auto-detection** — discovers the `slurmrestd` data_parser version (v0.0.40–v0.0.45) or accepts a pinned version.
- **Version-gating** — adapts request fields to the detected API version; surfaces warnings and errors from slurmrestd.
- **Dashboard** — real-time cluster overview: nodes up/down, jobs by state, partitions and accounts.
- **Jobs, Nodes, Partitions** — live tables with detail panels, search/filter (`/`), and CSV export.
- **Job actions** — cancel (`x`) and requeue (`r`) jobs directly from the TUI.
- **Query builder** — visual request composer for ping, get jobs, and submit jobs with cluster-gathered options (state, account, partition, qos, gres).
- **Custom query** — write or paste any request path and run it directly.
- **Request history** — every request logged with status, duration and warnings. Persisted across sessions (max 100 entries).

### Tabs

- **Dashboard** — real cluster overview: nodes up/down, jobs running/pending/completed/failed, partitions and accounts.
- **Jobs** — your jobs (slurmrestd filters by the authenticated user), with a detail panel (account, partition, time limit, run time, assigned nodes, log paths, exit code). Cancel (`x`) and requeue (`r`) jobs directly from the TUI. Press `enter` in Partitions to view jobs by partition.
- **Nodes** — cluster nodes with state, CPUs, memory and partitions; select a node to see its detail.
- **Partitions** — partition list with configured/total nodes and max wall time. Press `enter` to filter jobs by partition.
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
| `↑/↓` / `j/k` | navigate table rows |
| `PgUp/PgDn` / `b/f` | page up / page down |
| `Ctrl+U/Ctrl+D` | half page up / down |
| `Home/End` / `g/G` | go to start / end |
| `enter` | select / drill-down |
| `/` | filter the current table |
| `F5` | refresh (Jobs, Nodes, Partitions) |
| `x` | cancel selected job (Jobs tab) |
| `r` | requeue selected job (Jobs tab) |
| `?` | toggle help |

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
