## Cursor Cloud specific instructions

This is a single-binary Go monitoring agent with zero external dependencies (stdlib only). See `README.md` for full context.

### Build / Lint / Test

- **Build:** `go build -o monitoring-agent .`
- **Lint:** `go vet ./...`
- **Test:** `go test ./...` (no test files exist yet; command exits 0)

### Running the agent

The agent requires `PROJECT_KEY` env var (fatal without it). Minimal dev invocation:

```bash
PROJECT_KEY=<key> INTERVAL_SECONDS=5 go run .
```

Optional env vars: `SERVER_URL` (default `https://four20raw.ru/auth`), `AGENT_ID`, `AGENT_VERSION`, `HOST_ROOT`, `DOCKER_SOCK`.

### Caveats

- The agent runs an infinite heartbeat loop; use `timeout` or Ctrl-C to stop.
- Without a valid `PROJECT_KEY` the backend returns 401 but the agent keeps running and retrying — this is expected and confirms the agent works.
- `/proc/loadavg` and `/proc/meminfo` are available on this Linux VM so CPU/memory metrics work. Docker socket stats degrade to 0 if Docker is not running.
