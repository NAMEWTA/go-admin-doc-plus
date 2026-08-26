// Package config loads one validated runtime profile into an immutable value.
// It never reads process globals; bootstraps must pass every source explicitly.
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Profile identifies one independently validated runtime configuration shape.
type Profile string

const (
	ProfileServerPostgres Profile = "server-postgres"
	ProfileServerSQLite   Profile = "server-sqlite"
	ProfileDesktopSQLite  Profile = "desktop-sqlite"
)

// Input contains the complete startup-time configuration source set. CLI keys
// are configuration paths and may contain only non-sensitive overrides.
type Input struct {
	Profile     Profile
	File        string
	Environment map[string]string
	CLI         map[string]string
	Desktop     DesktopMaterial
}

// String intentionally omits every source value.
func (input Input) String() string {
	return fmt.Sprintf("config input profile=%s values=redacted", input.Profile)
}

// GoString intentionally omits every source value.
func (input Input) GoString() string {
	return fmt.Sprintf("config.Input{profile:%q, values:redacted}", input.Profile)
}

// DesktopMaterial is supplied by the native host for one sidecar invocation.
// It is deliberately separate from file, environment, and CLI configuration.
type DesktopMaterial struct {
	DataDirectory string
	LogDirectory  string
	LoopbackPort  uint16
	StartupToken  string
}

// String intentionally omits paths and the one-time token.
func (material DesktopMaterial) String() string { return "desktop launch material redacted" }

// GoString intentionally omits paths and the one-time token.
func (material DesktopMaterial) GoString() string { return "config.DesktopMaterial{redacted}" }

// Snapshot owns one profile value. Its fields are intentionally unexported so
// callers cannot mutate runtime configuration after construction.
type Snapshot struct {
	profile        Profile
	serverSQLite   ServerSQLite
	serverPostgres ServerPostgres
	desktopSQLite  DesktopSQLite
}

// Profile returns the one profile owned by this snapshot.
func (snapshot Snapshot) Profile() Profile { return snapshot.profile }

func (snapshot Snapshot) String() string {
	return fmt.Sprintf("config snapshot profile=%s values=redacted", snapshot.profile)
}

func (snapshot Snapshot) GoString() string {
	return fmt.Sprintf("config.Snapshot{profile:%q, values:redacted}", snapshot.profile)
}

// ServerSQLite returns an owned value only for ProfileServerSQLite.
func (snapshot Snapshot) ServerSQLite() (ServerSQLite, bool) {
	return snapshot.serverSQLite, snapshot.profile == ProfileServerSQLite
}

// ServerPostgres returns an owned value only for ProfileServerPostgres.
func (snapshot Snapshot) ServerPostgres() (ServerPostgres, bool) {
	return snapshot.serverPostgres, snapshot.profile == ProfileServerPostgres
}

// DesktopSQLite returns an owned value only for ProfileDesktopSQLite.
func (snapshot Snapshot) DesktopSQLite() (DesktopSQLite, bool) {
	return snapshot.desktopSQLite, snapshot.profile == ProfileDesktopSQLite
}

// ServerSQLite is the immutable Server SQLite runtime profile.
type ServerSQLite struct {
	httpListen   string
	logLevel     string
	databasePath string
}

// HTTPListen returns the validated server listen address.
func (profile ServerSQLite) HTTPListen() string { return profile.httpListen }

// LogLevel returns the validated log level.
func (profile ServerSQLite) LogLevel() string { return profile.logLevel }

// DatabasePath returns the Server SQLite database path.
func (profile ServerSQLite) DatabasePath() string { return profile.databasePath }
func (profile ServerSQLite) String() string       { return "server-sqlite configuration redacted" }
func (profile ServerSQLite) GoString() string     { return "config.ServerSQLite{redacted}" }

// ServerPostgres is the immutable Server PostgreSQL runtime profile.
type ServerPostgres struct {
	httpListen  string
	logLevel    string
	databaseDSN string
}

// HTTPListen returns the validated server listen address.
func (profile ServerPostgres) HTTPListen() string { return profile.httpListen }

// LogLevel returns the validated log level.
func (profile ServerPostgres) LogLevel() string { return profile.logLevel }

// DatabaseDSN returns the resolved in-memory secret for dependency construction.
// Callers must never log, format, persist, or expose the returned value.
func (profile ServerPostgres) DatabaseDSN() string { return profile.databaseDSN }
func (profile ServerPostgres) String() string      { return "server-postgres configuration redacted" }
func (profile ServerPostgres) GoString() string    { return "config.ServerPostgres{redacted}" }

// DesktopSQLite is the immutable native-host-provided Desktop SQLite profile.
type DesktopSQLite struct {
	logLevel        string
	dataDirectory   string
	logDirectory    string
	loopbackAddress string
	startupToken    string
}

// LogLevel returns the validated log level.
func (profile DesktopSQLite) LogLevel() string { return profile.logLevel }

// DataDirectory returns the absolute data directory supplied by Tauri.
func (profile DesktopSQLite) DataDirectory() string { return profile.dataDirectory }

// LogDirectory returns the distinct absolute log directory supplied by Tauri.
func (profile DesktopSQLite) LogDirectory() string { return profile.logDirectory }

// LoopbackAddress returns the fixed-loopback address using Tauri's random port.
func (profile DesktopSQLite) LoopbackAddress() string { return profile.loopbackAddress }

// StartupToken returns the one-time Tauri launch token. Callers must not log,
// persist, place in a URL, or expose this value to WebView JavaScript.
func (profile DesktopSQLite) StartupToken() string { return profile.startupToken }
func (profile DesktopSQLite) String() string       { return "desktop-sqlite configuration redacted" }
func (profile DesktopSQLite) GoString() string     { return "config.DesktopSQLite{redacted}" }

type serverSQLiteFile struct {
	Profile  Profile        `json:"profile"`
	HTTP     httpConfig     `json:"http"`
	Log      logConfig      `json:"log"`
	Database databaseConfig `json:"database"`
}

type serverPostgresFile struct {
	Profile Profile    `json:"profile"`
	HTTP    httpConfig `json:"http"`
	Log     logConfig  `json:"log"`
}

type desktopSQLiteFile struct {
	Profile Profile   `json:"profile"`
	Log     logConfig `json:"log"`
}

type httpConfig struct {
	Listen *string `json:"listen"`
}

type logConfig struct {
	Level *string `json:"level"`
}

type databaseConfig struct {
	Path *string `json:"path"`
}

// Load applies defaults, file, environment, then CLI and returns an owned
// immutable snapshot. It performs no listener, window, or dependency actions.
func Load(input Input) (Snapshot, error) {
	if err := validateSourceKeys(input); err != nil {
		return Snapshot{}, err
	}
	switch input.Profile {
	case ProfileServerPostgres:
		return loadServerPostgres(input)
	case ProfileServerSQLite:
		return loadServerSQLite(input)
	case ProfileDesktopSQLite:
		return loadDesktopSQLite(input)
	default:
		return Snapshot{}, fmt.Errorf("configuration profile: unsupported profile")
	}
}

func validateSourceKeys(input Input) error {
	commonEnvironment := map[string]struct{}{
		"GO_ADMIN_HTTP_LISTEN": {},
		"GO_ADMIN_LOG_LEVEL":   {},
	}
	commonCLI := map[string]struct{}{
		"http.listen": {},
		"log.level":   {},
	}
	allowedEnvironment := commonEnvironment
	allowedCLI := commonCLI
	switch input.Profile {
	case ProfileServerPostgres:
		allowedEnvironment["GO_ADMIN_DATABASE_DSN"] = struct{}{}
		allowedEnvironment["GO_ADMIN_DATABASE_DSN_FILE"] = struct{}{}
	case ProfileServerSQLite:
		allowedEnvironment["GO_ADMIN_SQLITE_PATH"] = struct{}{}
		allowedCLI["database.path"] = struct{}{}
	case ProfileDesktopSQLite:
		allowedEnvironment = map[string]struct{}{"GO_ADMIN_LOG_LEVEL": {}}
		allowedCLI = map[string]struct{}{"log.level": {}}
	default:
		return fmt.Errorf("configuration profile: unsupported profile")
	}
	for key := range input.Environment {
		if !strings.HasPrefix(key, "GO_ADMIN_") {
			continue
		}
		if _, ok := allowedEnvironment[key]; !ok {
			return fmt.Errorf("configuration environment.%s: unknown field", key)
		}
	}
	for key := range input.CLI {
		if _, ok := allowedCLI[key]; !ok {
			return fmt.Errorf("configuration cli.%s: override is not permitted", key)
		}
	}
	return nil
}

func loadDesktopSQLite(input Input) (Snapshot, error) {
	if !filepath.IsAbs(input.Desktop.DataDirectory) {
		return Snapshot{}, fmt.Errorf("configuration desktop.dataDirectory: absolute Tauri host path is required")
	}
	if !filepath.IsAbs(input.Desktop.LogDirectory) {
		return Snapshot{}, fmt.Errorf("configuration desktop.logDirectory: absolute Tauri host path is required")
	}
	if filepath.Clean(input.Desktop.DataDirectory) == filepath.Clean(input.Desktop.LogDirectory) {
		return Snapshot{}, fmt.Errorf("configuration desktop.logDirectory: must not conflict with data directory")
	}
	if input.Desktop.LoopbackPort == 0 {
		return Snapshot{}, fmt.Errorf("configuration desktop.loopbackPort: Tauri host port is required")
	}
	if len(input.Desktop.StartupToken) < 32 {
		return Snapshot{}, fmt.Errorf("configuration desktop.startupToken: Tauri host token is required")
	}
	value := DesktopSQLite{
		logLevel:        "info",
		dataDirectory:   filepath.Clean(input.Desktop.DataDirectory),
		logDirectory:    filepath.Clean(input.Desktop.LogDirectory),
		loopbackAddress: net.JoinHostPort("127.0.0.1", strconv.Itoa(int(input.Desktop.LoopbackPort))),
		startupToken:    input.Desktop.StartupToken,
	}
	if input.File != "" {
		file, err := decodeFile[desktopSQLiteFile](input.File)
		if err != nil {
			return Snapshot{}, err
		}
		if file.Profile != input.Profile {
			return Snapshot{}, fmt.Errorf("configuration profile: conflicts with selected profile")
		}
		if file.Log.Level != nil {
			value.logLevel = *file.Log.Level
		}
	}
	if candidate, exists := input.Environment["GO_ADMIN_LOG_LEVEL"]; exists {
		value.logLevel = candidate
	}
	if candidate, exists := input.CLI["log.level"]; exists {
		value.logLevel = candidate
	}
	if err := validateLogLevel(value.logLevel); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{profile: input.Profile, desktopSQLite: value}, nil
}

func loadServerPostgres(input Input) (Snapshot, error) {
	direct, hasDirect := input.Environment["GO_ADMIN_DATABASE_DSN"]
	secretPath, hasFile := input.Environment["GO_ADMIN_DATABASE_DSN_FILE"]
	if hasDirect && hasFile {
		return Snapshot{}, fmt.Errorf("configuration database.dsn: conflicting secret sources")
	}
	dsn := direct
	if hasFile {
		contents, err := os.ReadFile(secretPath)
		if err != nil {
			return Snapshot{}, fmt.Errorf("configuration database.dsn: secret file is unreadable")
		}
		dsn = strings.TrimRight(string(contents), "\r\n")
	}
	if strings.TrimSpace(dsn) == "" {
		return Snapshot{}, fmt.Errorf("configuration database.dsn: secret is required")
	}
	value := ServerPostgres{
		httpListen:  "127.0.0.1:8080",
		logLevel:    "info",
		databaseDSN: dsn,
	}
	if input.File != "" {
		file, err := decodeFile[serverPostgresFile](input.File)
		if err != nil {
			return Snapshot{}, err
		}
		if file.Profile != input.Profile {
			return Snapshot{}, fmt.Errorf("configuration profile: conflicts with selected profile")
		}
		if file.HTTP.Listen != nil {
			value.httpListen = *file.HTTP.Listen
		}
		if file.Log.Level != nil {
			value.logLevel = *file.Log.Level
		}
	}
	if candidate, exists := input.Environment["GO_ADMIN_HTTP_LISTEN"]; exists {
		value.httpListen = candidate
	}
	if candidate, exists := input.Environment["GO_ADMIN_LOG_LEVEL"]; exists {
		value.logLevel = candidate
	}
	if candidate, exists := input.CLI["http.listen"]; exists {
		value.httpListen = candidate
	}
	if candidate, exists := input.CLI["log.level"]; exists {
		value.logLevel = candidate
	}
	if err := validateServer(value.httpListen, value.logLevel); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{profile: input.Profile, serverPostgres: value}, nil
}

func loadServerSQLite(input Input) (Snapshot, error) {
	value := ServerSQLite{
		httpListen:   "127.0.0.1:8080",
		logLevel:     "info",
		databasePath: "go-admin-plus.sqlite3",
	}
	if input.File != "" {
		file, err := decodeFile[serverSQLiteFile](input.File)
		if err != nil {
			return Snapshot{}, err
		}
		if file.Profile != input.Profile {
			return Snapshot{}, fmt.Errorf("configuration profile: conflicts with selected profile")
		}
		if file.HTTP.Listen != nil {
			value.httpListen = *file.HTTP.Listen
		}
		if file.Log.Level != nil {
			value.logLevel = *file.Log.Level
		}
		if file.Database.Path != nil {
			value.databasePath = *file.Database.Path
		}
	}
	if candidate, exists := input.Environment["GO_ADMIN_HTTP_LISTEN"]; exists {
		value.httpListen = candidate
	}
	if candidate, exists := input.Environment["GO_ADMIN_LOG_LEVEL"]; exists {
		value.logLevel = candidate
	}
	if candidate, exists := input.Environment["GO_ADMIN_SQLITE_PATH"]; exists {
		value.databasePath = candidate
	}
	if candidate, exists := input.CLI["http.listen"]; exists {
		value.httpListen = candidate
	}
	if candidate, exists := input.CLI["log.level"]; exists {
		value.logLevel = candidate
	}
	if candidate, exists := input.CLI["database.path"]; exists {
		value.databasePath = candidate
	}
	if err := validateServer(value.httpListen, value.logLevel); err != nil {
		return Snapshot{}, err
	}
	if strings.TrimSpace(value.databasePath) == "" || strings.ContainsRune(value.databasePath, '\x00') {
		return Snapshot{}, fmt.Errorf("configuration database.path: non-empty filesystem path is required")
	}
	return Snapshot{profile: input.Profile, serverSQLite: value}, nil
}

func validateServer(listen, logLevel string) error {
	_, port, err := net.SplitHostPort(listen)
	if err != nil {
		return fmt.Errorf("configuration http.listen: valid host and port are required")
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return fmt.Errorf("configuration http.listen: port must be between 1 and 65535")
	}
	return validateLogLevel(logLevel)
}

func validateLogLevel(level string) error {
	switch level {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("configuration log.level: must be debug, info, warn, or error")
	}
}

func decodeFile[T any](path string) (T, error) {
	var value T
	contents, err := os.ReadFile(path)
	if err != nil {
		return value, fmt.Errorf("configuration file: unreadable")
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		const unknownPrefix = `json: unknown field "`
		if strings.HasPrefix(err.Error(), unknownPrefix) {
			field := strings.TrimSuffix(strings.TrimPrefix(err.Error(), unknownPrefix), `"`)
			return value, fmt.Errorf("configuration file.%s: unknown field", field)
		}
		return value, fmt.Errorf("configuration file: invalid JSON")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return value, fmt.Errorf("configuration file: must contain one JSON object")
	}
	return value, nil
}
