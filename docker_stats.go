package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type dockerContainer struct {
	ID     string `json:"Id"`
	State  string `json:"State"`
	Status string `json:"Status"`
}

func readDockerStats() (running int, unhealthy int) {
	sock := os.Getenv("DOCKER_SOCK")
	if sock == "" {
		sock = "/var/run/docker.sock"
	}
	if _, err := os.Stat(sock); err != nil {
		return 0, 0
	}

	tr := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", sock)
		},
	}
	client := &http.Client{Transport: tr, Timeout: 5 * time.Second}

	req, _ := http.NewRequest("GET", "http://docker/containers/json", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, 0
	}

	var cs []dockerContainer
	if err := json.NewDecoder(resp.Body).Decode(&cs); err != nil {
		return 0, 0
	}

	for _, c := range cs {
		if c.State == "running" {
			running++
		}
		if strings.Contains(strings.ToLower(c.Status), "unhealthy") {
			unhealthy++
		}
	}
	return running, unhealthy
}
