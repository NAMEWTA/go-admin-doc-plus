package desktop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"testing"
	"testing/fstest"
)

func TestOfflineRuntimeLoginAndDemoCRUD(t *testing.T) {
	assets := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("desktop")}}
	host, err := New(
		Config{Assets: assets, Version: "integration"},
		NewRuntimeBuilder(RuntimeConfig{DataRoot: t.TempDir(), Mode: "dev"}),
		func(ctx context.Context, bridge *Bridge, _ fs.FS) error {
			client := &http.Client{}
			payload := bridge.Bootstrap()
			request := func(method, path, jwtToken string, body any) map[string]any {
				t.Helper()
				var reader io.Reader
				if body != nil {
					encoded, err := json.Marshal(body)
					if err != nil {
						t.Fatalf("encode %s %s: %v", method, path, err)
					}
					reader = bytes.NewReader(encoded)
				}
				httpRequest, err := http.NewRequestWithContext(ctx, method, payload.APIBaseURL+path, reader)
				if err != nil {
					t.Fatalf("build %s %s: %v", method, path, err)
				}
				httpRequest.Header.Set("Origin", "wails://wails")
				httpRequest.Header.Set(payload.LaunchToken.Header, payload.LaunchToken.Value)
				if body != nil {
					httpRequest.Header.Set("Content-Type", "application/json")
				}
				if jwtToken != "" {
					httpRequest.Header.Set("Authorization", "Bearer "+jwtToken)
				}
				response, err := client.Do(httpRequest)
				if err != nil {
					t.Fatalf("send %s %s: %v", method, path, err)
				}
				defer response.Body.Close()
				contents, err := io.ReadAll(response.Body)
				if err != nil {
					t.Fatalf("read %s %s: %v", method, path, err)
				}
				if response.StatusCode != http.StatusOK {
					t.Fatalf("%s %s status %d: %s", method, path, response.StatusCode, contents)
				}
				if got := response.Header.Get("Access-Control-Allow-Origin"); got != "wails://wails" {
					t.Fatalf("%s %s CORS origin = %q", method, path, got)
				}
				var envelope map[string]any
				if err := json.Unmarshal(contents, &envelope); err != nil {
					t.Fatalf("decode %s %s: %v; body %s", method, path, err, contents)
				}
				if envelope["code"] != float64(http.StatusOK) || envelope["msg"] == nil {
					t.Fatalf("%s %s non-canonical envelope: %#v", method, path, envelope)
				}
				if _, exists := envelope["data"]; !exists {
					t.Fatalf("%s %s success envelope omitted data: %#v", method, path, envelope)
				}
				return envelope
			}

			login := request(http.MethodPost, "/api/v1/login", "", map[string]any{
				"username": "admin",
				"password": "123456",
				"code":     "1",
				"uuid":     "desktop",
			})
			token, ok := login["token"].(string)
			if !ok || token == "" {
				t.Fatalf("login token = %#v", login["token"])
			}

			code := "DESKTOP-OFFLINE-001"
			request(http.MethodPost, "/api/v1/demo-product", token, map[string]any{
				"name": "Offline Product", "code": code, "price": 12.5, "status": "1",
			})
			list := request(http.MethodGet, "/api/v1/demo-product?"+url.Values{
				"pageIndex": {"1"}, "pageSize": {"100"}, "code": {code},
			}.Encode(), token, nil)
			id := findProductID(t, list, code)
			request(http.MethodPut, fmt.Sprintf("/api/v1/demo-product/%d", id), token, map[string]any{
				"id": id, "name": "Offline Product Updated", "code": code, "price": 13.5, "status": "1",
			})
			request(http.MethodDelete, "/api/v1/demo-product", token, map[string]any{"ids": []int{id}})
			return nil
		},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := host.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func findProductID(t *testing.T, envelope map[string]any, code string) int {
	t.Helper()
	data, ok := envelope["data"].(map[string]any)
	if !ok {
		t.Fatalf("list data = %#v", envelope["data"])
	}
	rows, ok := data["list"].([]any)
	if !ok {
		t.Fatalf("list rows = %#v", data["list"])
	}
	for _, value := range rows {
		row, _ := value.(map[string]any)
		if row["code"] == code {
			id, _ := row["id"].(float64)
			return int(id)
		}
	}
	t.Fatalf("product %q not found in %#v", code, rows)
	return 0
}
