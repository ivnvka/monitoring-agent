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

### Local end-to-end testing without secrets

To test the full heartbeat flow without a real `PROJECT_KEY` or remote backend, point `SERVER_URL` at a local HTTP server:

```bash
# In one terminal — simple Python receiver that prints heartbeats:
python3 -c "
import http.server, json, sys
class H(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        data = json.loads(self.rfile.read(int(self.headers['Content-Length'])))
        print(json.dumps(data, indent=2)); sys.stdout.flush()
        self.send_response(200); self.end_headers(); self.wfile.write(b'{\"ok\":true}')
    def log_message(self, *a): pass
http.server.HTTPServer(('127.0.0.1', 9999), H).serve_forever()
"

# In another terminal — run the agent against it:
PROJECT_KEY=test SERVER_URL=http://127.0.0.1:9999 INTERVAL_SECONDS=5 go run .
```

### Caveats

- The agent runs an infinite heartbeat loop; use `timeout` or Ctrl-C to stop.
- Without a valid `PROJECT_KEY` the backend returns 401 but the agent keeps running and retrying — this is expected and confirms the agent works.
- `/proc/loadavg` and `/proc/meminfo` are available on this Linux VM so CPU/memory metrics work. Docker socket stats degrade to 0 if Docker is not running.
