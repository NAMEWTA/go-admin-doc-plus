package desktop

import "time"

const LaunchTokenHeader = "X-Go-Admin-Launch-Token"

type LaunchToken struct {
	Header string `json:"header"`
	Value  string `json:"value"`
}

type Capabilities struct {
	HostProfile   string `json:"hostProfile"`
	Desktop       bool   `json:"desktop"`
	Offline       bool   `json:"offline"`
	NativeDialogs bool   `json:"nativeDialogs"`
}

type SecurityPolicy struct {
	LoopbackOnly      bool     `json:"loopbackOnly"`
	LaunchTokenHeader string   `json:"launchTokenHeader"`
	AllowedOrigins    []string `json:"allowedOrigins"`
}

type BootstrapPayload struct {
	APIBaseURL       string         `json:"apiBaseUrl"`
	RequestTimeoutMS int64          `json:"requestTimeoutMs"`
	LaunchToken      LaunchToken    `json:"launchToken"`
	Version          string         `json:"version"`
	Capabilities     Capabilities   `json:"capabilities"`
	Security         SecurityPolicy `json:"security"`
}

// Bridge is the only Go binding exposed to the Admin application.
type Bridge struct {
	payload BootstrapPayload
}

func newBridge(baseURL, token, version string, requestTimeout time.Duration) *Bridge {
	return &Bridge{payload: BootstrapPayload{
		APIBaseURL:       baseURL,
		RequestTimeoutMS: requestTimeout.Milliseconds(),
		LaunchToken:      LaunchToken{Header: LaunchTokenHeader, Value: token},
		Version:          version,
		Capabilities: Capabilities{
			HostProfile: "desktop",
			Desktop:     true,
			Offline:     true,
		},
		Security: SecurityPolicy{
			LoopbackOnly:      true,
			LaunchTokenHeader: LaunchTokenHeader,
			AllowedOrigins:    []string{"http://wails.localhost", "wails://wails"},
		},
	}}
}

func (bridge *Bridge) Bootstrap() BootstrapPayload { return bridge.payload }
