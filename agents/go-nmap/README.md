# ZEPSEC Nmap Agent

A Go-based remote scanning agent for the ZEPSEC vulnerability tracking platform. Deploys to target machines to execute nmap scans and report results back to the ZEPSEC server for centralized network monitoring.

## Architecture

```
┌──────────────────┐         POST /scans          ┌──────────────────┐
│   ZEPSEC Server   │ ──── (dispatch scan job) ──→ │   Nmap Agent     │
│   (Rails app)     │                              │   (this binary)  │
│                   │ ←── POST /api/v1/ra_api ──── │                  │
│                   │      (scan results)           │   Runs nmap      │
└──────────────────┘                               │   Parses XML     │
                                                   │   Reports back   │
                                                   └──────────────────┘
```

**Two operating modes:**

1. **Server-dispatched** — The ZEPSEC server sends scan jobs to the agent via `POST /scans`. The agent runs nmap, parses the XML output, and POSTs structured results back to `/api/v1/ra_api`.

2. **Autonomous scheduled** — The agent runs periodic scans on its own schedule (configured in `config.yml`), reporting results back to the ZEPSEC server independently.

## Quick Start

### Build from source

```bash
# Build for current platform
make build

# Build for Linux (deploy target)
make build-linux

# Build for all platforms
make build-all
```

### Docker

```bash
# Build image
make docker

# Run
docker run -d \
  --name zepsec-agent \
  -p 8000:8000 \
  -v /path/to/config.yml:/app/config.yml:ro \
  --cap-add NET_RAW \
  zepsec-nmap-agent:latest
```

### Run directly

```bash
./zepsec-nmap-agent -config /path/to/config.yml
```

## Configuration

Copy `config.example.yml` to `config.yml` and edit for your environment. All settings support environment variable overrides:

| Config key | Env var | Description |
|---|---|---|
| `agent.listen_addr` | `ZEPSEC_LISTEN_ADDR` | Agent listen address (default: `0.0.0.0:8000`) |
| `agent.secret` | `ZEPSEC_AGENT_SECRET` | Shared secret for server→agent auth |
| `agent.tls_cert` | `ZEPSEC_TLS_CERT` | TLS certificate path (optional) |
| `agent.tls_key` | `ZEPSEC_TLS_KEY` | TLS key path (optional) |
| `server.url` | `ZEPSEC_SERVER_URL` | ZEPSEC server base URL |
| `server.api_token` | `ZEPSEC_API_TOKEN` | API token for result submission |
| `server.verify_tls` | `ZEPSEC_VERIFY_TLS` | Verify server TLS cert (default: `true`) |
| `nmap.binary_path` | `ZEPSEC_NMAP_PATH` | Path to nmap binary (default: `nmap`) |
| `nmap.use_sudo` | `ZEPSEC_NMAP_SUDO` | Run nmap with sudo (default: `true`) |
| `nmap.temp_dir` | `ZEPSEC_NMAP_TEMP_DIR` | Temp dir for XML files (default: `/tmp/zepsec-nmap`) |

## ZEPSEC Server Setup

1. In the ZEPSEC web UI, go to **Agents** and create a new agent entry:
   - **Name**: Descriptive name for this agent
   - **Hostname/Address**: IP or hostname where this agent is reachable
   - **Port**: The port the agent listens on (default: 8000)
   - **Protocol**: `http` or `https`
   - **Secret**: The shared bearer token (must match `agent.secret` in agent config)

2. Generate an API token from **Users → Generate API Token** and set it as `server.api_token` in the agent config.

3. Create a **Scan Job** in ZEPSEC and assign it to the new agent. Configure scan options and target hosts as needed.

4. Run the scan job from the ZEPSEC UI — it will dispatch to the agent automatically.

## API Endpoints

| Endpoint | Method | Auth | Description |
|---|---|---|---|
| `/scans` | POST | Bearer token | Receive scan job from ZEPSEC server |
| `/health` | GET | None | Health check (returns `{"status": "ok"}`) |
| `/status` | GET | Bearer token | List running scans |

## Scheduled Scans

For continuous network monitoring, configure autonomous scans in `config.yml`:

```yaml
scheduled_scans:
  - name: "perimeter"
    targets: "10.0.0.0/24"
    options: "-sS -sV -Pn -T4 --top-ports 1000"
    interval: "6h"
```

Intervals use Go duration format: `30m`, `1h`, `6h`, `24h`, etc.

## System Requirements

- **nmap** installed and in PATH (or specify full path in config)
- **sudo** access for privileged scans (SYN, OS detection)
- Network connectivity to ZEPSEC server and scan targets
- Outbound HTTPS for external IP detection (optional, falls back to local IP)

## Security Notes

- The agent authenticates incoming requests with a shared Bearer token
- Results are submitted to ZEPSEC using a user API token over HTTPS
- TLS is supported for the agent listener (configure `tls_cert`/`tls_key`)
- The Docker image runs as a non-root user with sudo access only for nmap
- Config files and environment variables containing secrets should be protected with appropriate file permissions
