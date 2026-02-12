package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Heartbeat struct {
	AgentID  string `json:"agentId"`
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
}

type Resp struct {
	OK bool `json:"ok"`
}

func getenv(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func main() {
	serverURL := strings.TrimRight(getenv("SERVER_URL", "https://four20raw.ru/auth"), "/")
	projectKey := getenv("PROJECT_KEY", "")
	agentID := getenv("AGENT_ID", "")
	interval := getenv("INTERVAL_SECONDS", "30")
	version := getenv("AGENT_VERSION", "0.1.0")

	if projectKey == "" {
		log.Fatal("PROJECT_KEY is required")
	}

	i, err := time.ParseDuration(interval + "s")
	if err != nil || i < 5*time.Second {
		i = 30 * time.Second
	}

	hostname, _ := os.Hostname()
	if agentID == "" {
		agentID = hostname
	}

	hbURL := serverURL + "/ingest/heartbeat"

	client := &http.Client{Timeout: 8 * time.Second}

	log.Printf("monitoring-agent starting: server=%s agentId=%s interval=%s", serverURL, agentID, i)
	for {
		payload := Heartbeat{AgentID: agentID, Hostname: hostname, Version: version}
		b, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", hbURL, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Project-Key", projectKey)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("heartbeat error: %v", err)
			time.Sleep(i)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Printf("heartbeat bad status: %s body=%q", resp.Status, string(body))
			time.Sleep(i)
			continue
		}

		var r Resp
		_ = json.Unmarshal(body, &r)
		log.Printf("heartbeat ok=%v", r.OK)
		time.Sleep(i)
	}
}
