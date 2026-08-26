//go:build desktop_e2e

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	desktophost "go-admin/internal/host/desktop"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const desktopE2EAuthenticatedEvent = "desktop-e2e-authenticated"

type desktopE2EEnvelope struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func startDesktopE2E(wailsCtx context.Context, bridge *desktophost.Bridge) {
	wailsRuntime.EventsOnce(wailsCtx, desktopE2EAuthenticatedEvent, func(data ...interface{}) {
		token, ok := firstString(data)
		if !ok {
			fmt.Println("GO_ADMIN_DESKTOP_E2E_FAIL: UI login did not provide a token")
			wailsRuntime.Quit(wailsCtx)
			return
		}
		go func() {
			ctx, cancel := context.WithTimeout(wailsCtx, 30*time.Second)
			defer cancel()
			if err := runDesktopE2ECRUD(ctx, bridge, token); err != nil {
				fmt.Printf("GO_ADMIN_DESKTOP_E2E_FAIL: %v\n", err)
			} else {
				fmt.Println("GO_ADMIN_DESKTOP_E2E_PASS")
			}
			wailsRuntime.Quit(wailsCtx)
		}()
	})
}

func firstString(data []interface{}) (string, bool) {
	if len(data) != 1 {
		return "", false
	}
	value, ok := data[0].(string)
	return value, ok && value != ""
}

func runDesktopE2ECRUD(ctx context.Context, bridge *desktophost.Bridge, token string) error {
	bootstrap := bridge.Bootstrap()
	client := &http.Client{Timeout: time.Duration(bootstrap.RequestTimeoutMS) * time.Millisecond}
	request := func(method, path string, body any) (desktopE2EEnvelope, error) {
		var payload io.Reader
		if body != nil {
			contents, err := json.Marshal(body)
			if err != nil {
				return desktopE2EEnvelope{}, fmt.Errorf("encode %s %s: %w", method, path, err)
			}
			payload = bytes.NewReader(contents)
		}
		httpRequest, err := http.NewRequestWithContext(ctx, method, bootstrap.APIBaseURL+path, payload)
		if err != nil {
			return desktopE2EEnvelope{}, fmt.Errorf("create %s %s: %w", method, path, err)
		}
		httpRequest.Header.Set("Origin", "wails://wails")
		httpRequest.Header.Set("Authorization", "Bearer "+token)
		httpRequest.Header.Set(bootstrap.LaunchToken.Header, bootstrap.LaunchToken.Value)
		if body != nil {
			httpRequest.Header.Set("Content-Type", "application/json")
		}
		response, err := client.Do(httpRequest)
		if err != nil {
			return desktopE2EEnvelope{}, fmt.Errorf("execute %s %s: %w", method, path, err)
		}
		defer response.Body.Close()
		contents, err := io.ReadAll(response.Body)
		if err != nil {
			return desktopE2EEnvelope{}, fmt.Errorf("read %s %s: %w", method, path, err)
		}
		if response.StatusCode != http.StatusOK {
			return desktopE2EEnvelope{}, fmt.Errorf("%s %s returned HTTP %d", method, path, response.StatusCode)
		}
		var envelope desktopE2EEnvelope
		if err := json.Unmarshal(contents, &envelope); err != nil {
			return desktopE2EEnvelope{}, fmt.Errorf("decode %s %s: %w", method, path, err)
		}
		if envelope.Code != http.StatusOK || envelope.Data == nil {
			return desktopE2EEnvelope{}, fmt.Errorf("%s %s returned an incompatible envelope", method, path)
		}
		return envelope, nil
	}

	if _, err := request(http.MethodGet, "/api/v1/getinfo", nil); err != nil {
		return fmt.Errorf("get current user: %w", err)
	}
	code := "WAILS-NATIVE-E2E"
	if _, err := request(http.MethodPost, "/api/v1/demo-product", map[string]any{
		"name": "Wails Native Product", "code": code, "price": 21.5, "status": "1",
	}); err != nil {
		return fmt.Errorf("create demo product: %w", err)
	}
	query := url.Values{"pageIndex": {"1"}, "pageSize": {"100"}, "code": {code}}
	listed, err := request(http.MethodGet, "/api/v1/demo-product?"+query.Encode(), nil)
	if err != nil {
		return fmt.Errorf("list demo products: %w", err)
	}
	var page struct {
		List []struct {
			ID   int    `json:"id"`
			Code string `json:"code"`
		} `json:"list"`
	}
	if err := json.Unmarshal(listed.Data, &page); err != nil {
		return fmt.Errorf("decode demo product list: %w", err)
	}
	var id int
	for _, row := range page.List {
		if row.Code == code {
			id = row.ID
			break
		}
	}
	if id == 0 {
		return errors.New("created demo product was not listed")
	}
	if _, err := request(http.MethodPut, fmt.Sprintf("/api/v1/demo-product/%d", id), map[string]any{
		"id": id, "name": "Wails Native Product Updated", "code": code, "price": 22.5, "status": "1",
	}); err != nil {
		return fmt.Errorf("update demo product: %w", err)
	}
	if _, err := request(http.MethodDelete, "/api/v1/demo-product", map[string]any{"ids": []int{id}}); err != nil {
		return fmt.Errorf("delete demo product: %w", err)
	}
	return nil
}

func desktopE2EScript() string {
	return `
(async () => {
  const started = sessionStorage.getItem('go-admin-desktop-e2e-started');
  const existingJWT = localStorage.getItem('Admin-Token');
  if (started && existingJWT) {
    window.runtime.EventsEmit('desktop-e2e-authenticated', existingJWT);
    return;
  }
  if (started) return;
  localStorage.removeItem('Admin-Token');
  sessionStorage.setItem('go-admin-desktop-e2e-started', '1');
  const waitFor = async (predicate, description, timeoutMs = 30000) => {
    const deadline = Date.now() + timeoutMs;
    while (Date.now() < deadline) {
      const value = predicate();
      if (value) return value;
      await new Promise(resolve => setTimeout(resolve, 100));
    }
    throw new Error('timed out waiting for ' + description);
  };
  const codeInput = await waitFor(
    () => document.querySelector('input[name="code"]'),
    'login form'
  );
  const storageSetItem = Storage.prototype.setItem;
  Storage.prototype.setItem = function(key, value) {
    storageSetItem.call(this, key, value);
    if (this === localStorage && key === 'Admin-Token' && value) {
      window.runtime.EventsEmit('desktop-e2e-authenticated', value);
    }
  };
  const valueSetter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set;
  codeInput.focus();
  valueSetter.call(codeInput, '1');
  codeInput.dispatchEvent(new Event('input', { bubbles: true, composed: true }));
  codeInput.dispatchEvent(new Event('change', { bubbles: true, composed: true }));
  await new Promise(resolve => setTimeout(resolve, 250));
  const submitButton = document.querySelector('.submit-btn');
  if (!submitButton) throw new Error('login submit button is unavailable');
  submitButton.click();
})().catch(error => {
  console.error('GO_ADMIN_DESKTOP_E2E_FAIL', error);
  document.title = 'GO_ADMIN_DESKTOP_E2E_FAIL';
});`
}
