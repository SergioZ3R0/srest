# Test Lab

Guide to spin up a **real** Slurm cluster in containers and test `srest`
against it, without HPC infrastructure or access to a production cluster.

There are two options:

1. **Full Slurm cluster** (recommended): `slurmctld` + `slurmdbd` +
   `slurmrestd` + compute nodes, with swappable Slurm versions.
2. **Lightweight mock**: a minimal HTTP server that imitates `/ping` (to
   iterate quickly on the UI without Docker).

---

## 1. Full Slurm cluster (Docker)

We use [`giovtorres/slurm-docker-cluster`](https://github.com/giovtorres/slurm-docker-cluster),
which brings up the whole stack and exposes `slurmrestd` on port `6820`.

### Requirements

- Docker + Docker Compose.

### Step by step

```bash
# 1. Clone the cluster repo
git clone https://github.com/giovtorres/slurm-docker-cluster.git
cd slurm-docker-cluster
cp .env.example .env            # defaults to Slurm 26.05.2

# 2. Pull the prebuilt image (faster than building)
docker pull giovtorres/slurm-docker-cluster:26.05.2
docker tag giovtorres/slurm-docker-cluster:26.05.2 slurm-docker-cluster:26.05.2

# 3. Start the cluster
make up

# 4. Verify everything is healthy (slurmrestd on :6820)
make status
```

### Generate the JWT token

```bash
docker exec slurmctld scontrol token
# prints something like: SLURM_JWT=eyJhbGciOi...
```

### Connect `srest`

```bash
cd /path/to/srest
SLURM_URL=http://localhost:6820 \
SLURM_JWT=<token> \
SLURM_USER_NAME=slurm \
go run .
```

You should see something like:

```
srest

Connection successful!
API:   v0.0.45
Slurm: 26.05.2

Press 'q' to quit
```

### Run the integration test

```bash
SLURM_URL=http://localhost:6820 SLURM_JWT=<token> \
  go test ./internal/api -run TestPingIntegration -v
```

### Try multiple API versions

The cluster supports several Slurm versions, which determine the `data_parser`
plugin version:

| Slurm   | data_parser |
| ------- | ----------- |
| 26.05.2 | v0.0.45     |
| 25.11.4 | v0.0.44     |

```bash
make set-version VER=25.11.4   # change the version in .env
make rebuild                    # clean, rebuild and start
```

`srest` will auto-detect the available version (or you can pin it with
`SLURM_API_VERSION`).

### Stop / clean up

```bash
make down        # stop (keeps data)
make clean       # stop and remove volumes
```

---

## 2. Lightweight mock (no Docker)

To iterate on the UI without spinning up a cluster, use a `slurmrestd` mock
that only answers `/ping`. You can use the following script (requires Python 3):

```python
#!/usr/bin/env python3
import json
from http.server import BaseHTTPRequestHandler, HTTPServer

PING = {
    "meta": {
        "plugin": {"type": "openapi/v0.0.44", "name": "Slurm OpenAPI v0.0.44"},
        "slurm": {"version": {}, "release": "23.11.9"},
    },
    "errors": [],
    "warnings": [],
}

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/slurm/v0.0.44/ping":
            body = json.dumps(PING).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, *args):
        pass

HTTPServer(("127.0.0.1", 6820), Handler).serve_forever()
```

```bash
python3 mock_slurm.py &
SLURM_URL=http://localhost:6820 go run .
```

The mock answers `200` on `v0.0.44` and `404` on the rest, so `srest`'s
auto-detection picks `v0.0.44`.
