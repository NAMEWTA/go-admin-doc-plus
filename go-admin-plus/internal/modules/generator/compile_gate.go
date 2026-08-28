package generator

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const compileGateTimeout = 10 * time.Minute

var compileGateSemaphore = make(chan struct{}, 1)

// WorkspaceCompileGate overlays a tracked repository skeleton, regenerates canonical
// transports, and compiles/tests both generated packages before output can be published.
type WorkspaceCompileGate struct{ skeleton string }

func NewWorkspaceCompileGate(skeleton string) (*WorkspaceCompileGate, error) {
	root, err := canonicalDirectory(skeleton)
	if err != nil {
		return nil, ErrInvalid
	}
	for _, path := range []string{".git", "scripts/contracts/cli.mjs", "go-admin-plus/go.mod", "go-admin-plus-ui/pnpm-workspace.yaml"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			return nil, ErrInvalid
		}
	}
	return &WorkspaceCompileGate{skeleton: root}, nil
}
func (*WorkspaceCompileGate) CompleteOutputGate() {}

func (gate *WorkspaceCompileGate) Check(ctx context.Context, _ string, preview Preview) error {
	select {
	case compileGateSemaphore <- struct{}{}:
		defer func() { <-compileGateSemaphore }()
	case <-ctx.Done():
		return context.Cause(ctx)
	}
	ctx, cancel := context.WithTimeout(ctx, compileGateTimeout)
	defer cancel()

	fixture, err := os.MkdirTemp("", "go-admin-generated-fixture-")
	if err != nil {
		return ErrGateFailed
	}
	defer os.RemoveAll(fixture)
	if err := copyTrackedSkeleton(ctx, gate.skeleton, fixture); err != nil {
		return gateContextError(ctx)
	}
	for _, file := range preview.Files {
		destination, err := safeJoin(fixture, file.Path)
		if err != nil {
			return ErrGateFailed
		}
		if _, err := os.Lstat(destination); err == nil {
			return ErrConflict
		} else if !os.IsNotExist(err) {
			return ErrGateFailed
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o750); err != nil {
			return ErrGateFailed
		}
		if err := os.WriteFile(destination, []byte(file.Content), 0o640); err != nil {
			return ErrGateFailed
		}
	}

	goRoot := filepath.Join(fixture, "go-admin-plus")
	uiRoot := filepath.Join(fixture, "go-admin-plus-ui")
	safeHome := filepath.Join(fixture, ".home")
	if err := os.Mkdir(safeHome, 0o700); err != nil {
		return ErrGateFailed
	}
	environment := minimalCommandEnvironment(safeHome)
	commands := compileGateCommands(goRoot, uiRoot, preview.Module)
	for _, command := range commands {
		if err := command.run(ctx, environment); err != nil {
			return gateContextError(ctx)
		}
	}
	for _, file := range preview.Files {
		content, err := os.ReadFile(filepath.Join(fixture, filepath.FromSlash(file.Path)))
		if err != nil || string(content) != file.Content {
			return ErrGateFailed
		}
	}
	return nil
}

type gateCommand struct {
	directory string
	name      string
	arguments []string
}

func compileGateCommands(goRoot, uiRoot, module string) []gateCommand {
	fixture := filepath.Dir(goRoot)
	return []gateCommand{
		{uiRoot, pnpmExecutable(), []string{"install", "--lockfile-only", "--no-frozen-lockfile", "--ignore-scripts"}},
		{fixture, "node", []string{"scripts/contracts/cli.mjs", "lint", "--contract", "contracts/openapi/modules/" + module + ".yaml"}},
		{fixture, "node", []string{"scripts/contracts/cli.mjs", "generate"}},
		{fixture, "node", []string{"scripts/contracts/cli.mjs", "generate", "--check"}},
		{goRoot, "go", []string{"test", "./internal/application"}},
		{goRoot, "go", []string{"test", "./internal/modules/" + module + "/..."}},
		{goRoot, "go", []string{"build", "./internal/modules/" + module + "/..."}},
		{uiRoot, pnpmExecutable(), []string{"install", "--frozen-lockfile", "--ignore-scripts"}},
		{uiRoot, pnpmExecutable(), []string{"--filter", "@go-admin-plus/domain-" + module, "typecheck"}},
		{uiRoot, pnpmExecutable(), []string{"--filter", "@go-admin-plus/domain-" + module, "test"}},
		{uiRoot, pnpmExecutable(), []string{"--filter", "@go-admin-plus/web-domain-" + module, "typecheck"}},
		{uiRoot, pnpmExecutable(), []string{"--filter", "@go-admin-plus/web-domain-" + module, "test"}},
		{uiRoot, pnpmExecutable(), []string{"check:workspace"}},
	}
}

func (command gateCommand) run(ctx context.Context, environment []string) error {
	executable, err := resolveToolExecutable(command.name)
	if err != nil {
		return err
	}
	process := exec.CommandContext(ctx, executable, command.arguments...)
	process.Dir = command.directory
	process.Env = environment
	process.Stdout = io.Discard
	process.Stderr = io.Discard
	return process.Run()
}

func gateContextError(ctx context.Context) error {
	if cause := context.Cause(ctx); cause != nil {
		return cause
	}
	return ErrGateFailed
}

func pnpmExecutable() string {
	if runtime.GOOS == "windows" {
		return "pnpm.cmd"
	}
	return "pnpm"
}

var compileSkeletonPrefixes = []string{
	"contracts/openapi/",
	"scripts/contracts/",
	"go-admin-plus/",
	"go-admin-plus-ui/apps/",
	"go-admin-plus-ui/packages/",
	"go-admin-plus-ui/tests/shell/",
}

var compileSkeletonFiles = map[string]struct{}{
	"go-admin-plus-ui/.npmrc":              {},
	"go-admin-plus-ui/package.json":        {},
	"go-admin-plus-ui/pnpm-lock.yaml":      {},
	"go-admin-plus-ui/pnpm-workspace.yaml": {},
}

func copyTrackedSkeleton(ctx context.Context, source, destination string) error {
	gitExecutable, err := resolveToolExecutable("git")
	if err != nil {
		return err
	}
	command := exec.CommandContext(ctx, gitExecutable, "ls-files", "-z")
	command.Dir = source
	command.Env = minimalCommandEnvironment("")
	output, err := command.Output()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Split(splitZero)
	for scanner.Scan() {
		relative := filepath.ToSlash(scanner.Text())
		if !allowedSkeletonPath(relative) {
			continue
		}
		from := filepath.Join(source, filepath.FromSlash(relative))
		info, err := os.Lstat(from)
		if err != nil || !info.Mode().IsRegular() {
			return ErrInvalid
		}
		to := filepath.Join(destination, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(to), 0o750); err != nil {
			return err
		}
		content, err := os.ReadFile(from)
		if err != nil {
			return err
		}
		if err := os.WriteFile(to, content, info.Mode().Perm()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func splitZero(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if index := bytes.IndexByte(data, 0); index >= 0 {
		return index + 1, data[:index], nil
	}
	if atEOF && len(data) != 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func allowedSkeletonPath(path string) bool {
	for _, segment := range strings.Split(path, "/") {
		if segment == ".env" || strings.HasPrefix(segment, ".env.") {
			return false
		}
	}
	if _, ok := compileSkeletonFiles[path]; ok {
		return true
	}
	for _, prefix := range compileSkeletonPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

var environmentOnce sync.Once
var baseEnvironment []string

func minimalCommandEnvironment(safeHome string) []string {
	environmentOnce.Do(func() {
		allowed := map[string]struct{}{
			"PATH": {}, "SystemRoot": {}, "WINDIR": {}, "PATHEXT": {},
			"TMPDIR": {}, "TMP": {}, "TEMP": {},
			"GOCACHE": {}, "GOMODCACHE": {}, "GOPATH": {}, "GOENV_ROOT": {},
		}
		for _, item := range os.Environ() {
			key, _, found := strings.Cut(item, "=")
			if _, ok := allowed[key]; found && ok {
				baseEnvironment = append(baseEnvironment, item)
			}
		}
		baseEnvironment = replaceEnvironmentValue(baseEnvironment, "PATH", toolchainPath(os.Getenv("PATH")))
		nullConfig := "/dev/null"
		if runtime.GOOS == "windows" {
			nullConfig = "NUL"
		}
		baseEnvironment = append(baseEnvironment,
			"CI=1",
			"GIT_TERMINAL_PROMPT=0",
			"NPM_CONFIG_USERCONFIG="+nullConfig,
			"GOENV=off",
			"GOTOOLCHAIN=local",
			"GOPROXY=off",
			"GOSUMDB=off",
		)
	})
	result := append([]string(nil), baseEnvironment...)
	if safeHome != "" {
		result = append(result, "HOME="+safeHome, "USERPROFILE="+safeHome, "XDG_CONFIG_HOME="+filepath.Join(safeHome, ".config"), "XDG_CACHE_HOME="+filepath.Join(safeHome, ".cache"))
	}
	return result
}

func toolchainPath(path string) string {
	prefixes := make([]string, 0, 2)
	if node, err := resolveToolExecutable("node"); err == nil {
		prefixes = append(prefixes, filepath.Dir(node))
	}
	pnpm, err := exec.LookPath(pnpmExecutable())
	if err == nil {
		clean := filepath.ToSlash(filepath.Clean(pnpm))
		marker := "/.volta/bin/"
		if index := strings.Index(clean, marker); index >= 0 {
			candidate := filepath.FromSlash(clean[:index] + "/.volta/tools/image/packages/pnpm/bin")
			if info, statErr := os.Stat(filepath.Join(candidate, pnpmExecutable())); statErr == nil && !info.IsDir() {
				prefixes = append(prefixes, candidate)
			}
		}
	}
	if len(prefixes) == 0 {
		return path
	}
	return strings.Join(prefixes, string(os.PathListSeparator)) + string(os.PathListSeparator) + path
}

func resolveToolExecutable(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	if name == "pnpm" || name == "pnpm.cmd" {
		clean := filepath.ToSlash(filepath.Clean(path))
		if index := strings.Index(clean, "/.volta/bin/"); index >= 0 {
			candidate := filepath.FromSlash(clean[:index] + "/.volta/tools/image/packages/pnpm/bin/" + pnpmExecutable())
			if info, statErr := os.Stat(candidate); statErr == nil && info.Mode().IsRegular() && (runtime.GOOS == "windows" || info.Mode().Perm()&0o111 != 0) {
				path = candidate
			}
		}
	} else if name == "node" {
		path = resolveVoltaNodeExecutable(path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absolute)
	if err != nil || !info.Mode().IsRegular() || (runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0) {
		return "", ErrInvalid
	}
	return absolute, nil
}

func resolveVoltaNodeExecutable(path string) string {
	if runtime.GOOS == "windows" {
		return path
	}
	voltaHome := filepath.Dir(filepath.Dir(filepath.Clean(path)))
	metadata, err := os.ReadFile(filepath.Join(voltaHome, "tools", "user", "platform.json"))
	if err != nil {
		return path
	}
	var platform struct {
		Node struct {
			Runtime string `json:"runtime"`
		} `json:"node"`
	}
	if json.Unmarshal(metadata, &platform) != nil || !validVoltaRuntime(platform.Node.Runtime) {
		return path
	}
	candidate := filepath.Join(voltaHome, "tools", "image", "node", platform.Node.Runtime, "bin", "node")
	info, err := os.Stat(candidate)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
		return path
	}
	return candidate
}

func validVoltaRuntime(version string) bool {
	return version != "" && version != "." && version != ".." && filepath.Base(version) == version
}

func replaceEnvironmentValue(environment []string, key, value string) []string {
	prefix := key + "="
	for index := range environment {
		if strings.HasPrefix(environment[index], prefix) {
			environment[index] = prefix + value
			return environment
		}
	}
	return append(environment, prefix+value)
}
