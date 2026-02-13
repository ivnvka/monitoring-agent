package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Heartbeat struct {
	AgentID        string  `json:"agentId"`
	Hostname       string  `json:"hostname"`
	Version         string  `json:"version"`
	Load1           float64 `json:"load1"`
	MemTotalBytes   int64   `json:"memTotalBytes"`
	MemUsedBytes    int64   `json:"memUsedBytes"`
	MemUsedPct      float64 `json:"memUsedPct"`
	DiskTotalBytes  int64   `json:"diskTotalBytes"`
	DiskUsedBytes   int64   `json:"diskUsedBytes"`
	DiskUsedPct     float64 `json:"diskUsedPct"`
	DockerRunning   int     `json:"dockerRunning"`
	DockerUnhealthy int     `json:"dockerUnhealthy"`
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

func readLoad1() float64 {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0
	}
	parts := strings.Fields(string(b))
	if len(parts) < 1 {
		return 0
	}
	v, _ := strconv.ParseFloat(parts[0], 64)
	return v
}

func readMem() (total, used int64, usedPct float64) {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, 0
	}
	var memTotalKB, memAvailKB int64
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			f := strings.Fields(line)
			if len(f) >= 2 {
				memTotalKB, _ = strconv.ParseInt(f[1], 10, 64)
			}
		}
		if strings.HasPrefix(line, "MemAvailable:") {
			f := strings.Fields(line)
			if len(f) >= 2 {
				memAvailKB, _ = strconv.ParseInt(f[1], 10, 64)
			}
		}
	}
	total = memTotalKB * 1024
	used = (memTotalKB - memAvailKB) * 1024
	if total > 0 {
		usedPct = float64(used) * 100 / float64(total)
	}
	return
}

func statDisk(path string) (total, used int64, usedPct float64) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, 0, 0
	}
	total = int64(st.Blocks) * int64(st.Bsize)
	free := int64(st.Bavail) * int64(st.Bsize)
	used = total - free
	if total > 0 {
		usedPct = float64(used) * 100 / float64(total)
	}
	return
}

func main() {
	serverURL := strings.TrimRight(getenv("SERVER_URL", "https://four20raw.ru/auth"), "/")
	projectKey := getenv("PROJECT_KEY", "")
	agentID := getenv("AGENT_ID", "")
	interval := getenv("INTERVAL_SECONDS", "30")
	version := getenv("AGENT_VERSION", "0.2.0")
	hostRoot := getenv("HOST_ROOT", "/")

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

	// if user mounted host root, prefer using it for disk stats.
	if hostRoot != "/" {
		if _, err := os.Stat(hostRoot); err != nil {
			hostRoot = "/"
		}
	}
	// statfs needs a real path
	hostRoot = filepath.Clean(hostRoot)

	hbURL := serverURL + "/ingest/heartbeat"
	client := &http.Client{Timeout: 8 * time.Second}

	log.Printf("monitoring-agent starting: server=%s agentId=%s interval=%s hostRoot=%s", serverURL, agentID, i, hostRoot)
	for {
		load1 := readLoad1()
		memTotal, memUsed, memUsedPct := readMem()
		diskTotal, diskUsed, diskUsedPct := statDisk(hostRoot)
		dockerRunning, dockerUnhealthy := readDockerStats()

		payload := Heartbeat{
			AgentID:        agentID,
			Hostname:       hostname,
			Version:        version,
			Load1:          load1,
			MemTotalBytes:  memTotal,
			MemUsedBytes:   memUsed,
			MemUsedPct:     memUsedPct,
			DiskTotalBytes: diskTotal,
			DiskUsedBytes:  diskUsed,
			DiskUsedPct:    diskUsedPct,
			DockerRunning:  dockerRunning,
			DockerUnhealthy: dockerUnhealthy,
		}
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
		log.Printf("heartbeat ok=%v mem=%.1f%% disk=%.1f%% load1=%.2f", r.OK, memUsedPct, diskUsedPct, load1)
		time.Sleep(i)
	}
}
