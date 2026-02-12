# monitoring-agent

Docker agent for **Monitoring Mini App (MVP)**.

It sends periodic heartbeats to the backend so the Mini App can show **online/offline** status.

## Run (recommended)

1) Get your `PROJECT_KEY` from the Mini App → Project page.

2) Run on your server:

```bash
docker run -d --name monitoring-agent --restart unless-stopped \
  -e SERVER_URL=https://four20raw.ru/auth \
  -e PROJECT_KEY=<PASTE_PROJECT_KEY_HERE> \
  ghcr.io/ivnvka/monitoring-agent:latest
```

### Optional env

- `AGENT_ID` (default: hostname)
- `INTERVAL_SECONDS` (default: 30)
- `AGENT_VERSION` (default: 0.1.0)

## Development

```bash
go run .
```

## GitHub Container Registry

This repo includes a GitHub Actions workflow that builds & publishes the image to:

- `ghcr.io/ivnvka/monitoring-agent:latest`

You may need to enable Actions permissions:

- Repo → Settings → Actions → General → Workflow permissions → **Read and write permissions**

