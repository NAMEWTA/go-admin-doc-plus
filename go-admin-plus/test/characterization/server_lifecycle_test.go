package characterization_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestServerProcessBaseline(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	tempDir := t.TempDir()
	binaryName := "go-admin-characterization"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(tempDir, binaryName)
	configPath := filepath.Join(tempDir, "settings.yml")
	databasePath := filepath.Join(tempDir, "baseline.db")
	port := availablePort(t)

	config := fmt.Sprintf(`settings:
  application:
    mode: dev
    host: 127.0.0.1
    name: characterization
    port: %d
    readtimeout: 3
    writertimeout: 3
    enabledp: false
  logger:
    path: %s
    stdout: default
    level: error
    enableddb: false
  jwt:
    secret: characterization-only-secret
    timeout: 3600
  database:
    driver: sqlite3
    source: %s
  queue:
    memory:
      poolSize: 32
`, port, filepath.ToSlash(filepath.Join(tempDir, "logs")), filepath.ToSlash(databasePath))
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatalf("write temporary config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-tags", "sqlite3", "-o", binaryPath, ".")
	build.Dir = repoRoot
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build sqlite binary: %v\n%s", err, output)
	}

	migrate := exec.CommandContext(ctx, binaryPath, "migrate", "-c", configPath)
	migrate.Dir = repoRoot
	if output, err := migrate.CombinedOutput(); err != nil {
		t.Fatalf("migrate temporary database: %v\n%s", err, output)
	}

	var serverOutput bytes.Buffer
	server := exec.Command(binaryPath, "server", "-c", configPath)
	server.Dir = repoRoot
	server.Stdout = &serverOutput
	server.Stderr = &serverOutput
	if err := server.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	waitCh := make(chan error, 1)
	doneCh := make(chan struct{})
	go func() {
		waitCh <- server.Wait()
		close(doneCh)
	}()
	stopped := false
	t.Cleanup(func() {
		if !stopped && server.Process != nil {
			_ = server.Process.Kill()
			select {
			case <-doneCh:
			case <-time.After(5 * time.Second):
			}
		}
	})

	baseURL := "http://127.0.0.1:" + strconv.Itoa(port)
	client := &http.Client{Timeout: 2 * time.Second}
	waitForServer(t, client, baseURL+"/info", waitCh, &serverOutput)

	health := requestEnvelope(t, client, http.MethodGet, baseURL+"/info", "", nil)
	if health["message"] != "ok" {
		t.Fatalf("health response = %#v, want message=ok", health)
	}

	badLogin := requestEnvelope(t, client, http.MethodPost, baseURL+"/api/v1/login", "", map[string]any{
		"username": "admin",
		"password": "definitely-wrong",
		"code":     "0",
		"uuid":     "0",
	})
	assertCode(t, badLogin, http.StatusBadRequest)

	login := requestEnvelope(t, client, http.MethodPost, baseURL+"/api/v1/login", "", map[string]any{
		"username": "admin",
		"password": "123456",
		"code":     "0",
		"uuid":     "0",
	})
	assertCode(t, login, http.StatusOK)
	token, ok := login["token"].(string)
	if !ok || token == "" {
		t.Fatalf("login envelope omitted token: %#v", login)
	}

	for _, path := range []string{"/api/v1/getinfo", "/api/v1/menurole", "/api/v1/sysjob?pageIndex=1&pageSize=10"} {
		envelope := requestEnvelope(t, client, http.MethodGet, baseURL+path, token, nil)
		assertCode(t, envelope, http.StatusOK)
		if _, ok := envelope["data"]; !ok {
			t.Fatalf("%s response omitted data: %#v", path, envelope)
		}
	}

	created := requestEnvelope(t, client, http.MethodPost, baseURL+"/api/v1/demo-product", token, map[string]any{
		"name":   "Characterization Product",
		"code":   "BASELINE-001",
		"price":  12.5,
		"status": "2",
		"remark": "temporary fixture",
	})
	assertCode(t, created, http.StatusOK)
	id, ok := created["data"].(float64)
	if !ok || id <= 0 {
		t.Fatalf("create response data = %#v, want positive id", created["data"])
	}
	idPath := strconv.Itoa(int(id))

	updated := requestEnvelope(t, client, http.MethodPut, baseURL+"/api/v1/demo-product/"+idPath, token, map[string]any{
		"id":     int(id),
		"name":   "Characterization Product Updated",
		"code":   "BASELINE-001",
		"price":  15.75,
		"status": "2",
		"remark": "temporary fixture",
	})
	assertCode(t, updated, http.StatusOK)

	detail := requestEnvelope(t, client, http.MethodGet, baseURL+"/api/v1/demo-product/"+idPath, token, nil)
	assertCode(t, detail, http.StatusOK)
	data, ok := detail["data"].(map[string]any)
	if !ok || data["name"] != "Characterization Product Updated" {
		t.Fatalf("detail data = %#v, want updated product", detail["data"])
	}

	deleted := requestEnvelope(t, client, http.MethodDelete, baseURL+"/api/v1/demo-product", token, map[string]any{
		"ids": []int{int(id)},
	})
	assertCode(t, deleted, http.StatusOK)

	if runtime.GOOS == "windows" {
		if err := server.Process.Kill(); err != nil {
			t.Fatalf("stop server process: %v", err)
		}
	} else if err := server.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("interrupt server process: %v", err)
	}

	select {
	case err := <-waitCh:
		stopped = true
		if runtime.GOOS != "windows" && err != nil {
			t.Fatalf("server exited after interrupt: %v\n%s", err, serverOutput.String())
		}
	case <-time.After(8 * time.Second):
		t.Fatalf("server did not stop within 8s\n%s", serverOutput.String())
	}
}

func availablePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve loopback port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func waitForServer(t *testing.T, client *http.Client, url string, waitCh <-chan error, output *bytes.Buffer) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-waitCh:
			t.Fatalf("server exited before readiness: %v\n%s", err, output.String())
		default:
		}
		response, err := client.Get(url)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("server was not ready within 15s\n%s", output.String())
}

func requestEnvelope(
	t *testing.T,
	client *http.Client,
	method string,
	url string,
	token string,
	body any,
) map[string]any {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode request body: %v", err)
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer response.Body.Close()
	encoded, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read %s %s response: %v", method, url, err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("%s %s status = %d, body = %s", method, url, response.StatusCode, encoded)
	}
	var envelope map[string]any
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		t.Fatalf("decode %s %s response %q: %v", method, url, encoded, err)
	}
	return envelope
}

func assertCode(t *testing.T, envelope map[string]any, want int) {
	t.Helper()
	got, ok := envelope["code"].(float64)
	if !ok || int(got) != want {
		t.Fatalf("code = %#v, want %d; envelope = %#v", envelope["code"], want, envelope)
	}
	if message, ok := envelope["msg"].(string); ok && strings.TrimSpace(message) == "" {
		t.Fatal("response included an empty msg")
	}
}
