package desktop_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	desktophost "go-admin/internal/host/desktop"
	"go-admin/internal/platform"
	desktopplatform "go-admin/internal/platform/desktop"
)

const (
	lockHelperEnvironment = "GO_ADMIN_TEST_INSTANCE_LOCK_HELPER"
	lockRootEnvironment   = "GO_ADMIN_TEST_INSTANCE_LOCK_ROOT"
)

func TestNativeSingleInstanceLockAcrossProcesses(t *testing.T) {
	if os.Getenv(lockHelperEnvironment) == "1" {
		runInstanceLockHelper()
		return
	}

	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(base, "app-data")
	command := exec.Command(os.Args[0], "-test.run=^TestNativeSingleInstanceLockAcrossProcesses$")
	command.Env = append(os.Environ(), lockHelperEnvironment+"=1", lockRootEnvironment+"="+root)
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatalf("helper stdout: %v", err)
	}
	stdin, err := command.StdinPipe()
	if err != nil {
		t.Fatalf("helper stdin: %v", err)
	}
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		t.Fatalf("start helper: %v", err)
	}
	ready, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil || strings.TrimSpace(ready) != "READY" {
		_ = command.Process.Kill()
		t.Fatalf("helper readiness = %q, %v", ready, err)
	}
	if lock, err := desktopplatform.AcquireSecureInstanceLock(root); !errors.Is(err, desktopplatform.ErrInstanceLocked) || lock != nil {
		_ = command.Process.Kill()
		t.Fatalf("second process lock = %#v, %v; want ErrInstanceLocked", lock, err)
	}
	if err := stdin.Close(); err != nil {
		t.Fatalf("close helper stdin: %v", err)
	}
	if err := command.Wait(); err != nil {
		t.Fatalf("wait helper: %v", err)
	}

	lock, err := desktopplatform.AcquireSecureInstanceLock(root)
	if err != nil {
		t.Fatalf("reacquire after helper exit: %v", err)
	}
	if err := lock.Close(); err != nil {
		t.Fatalf("close reacquired lock: %v", err)
	}
}

func runInstanceLockHelper() {
	lock, err := desktopplatform.AcquireSecureInstanceLock(os.Getenv(lockRootEnvironment))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "acquire helper lock: %v\n", err)
		os.Exit(2)
	}
	_, _ = fmt.Fprintln(os.Stdout, "READY")
	_, _ = io.Copy(io.Discard, os.Stdin)
	if err := lock.Close(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "close helper lock: %v\n", err)
		os.Exit(3)
	}
}

func TestDesktopFileStorePersistsAndRejectsTraversalAfterRestart(t *testing.T) {
	root := filepath.Join(t.TempDir(), "files")
	store, err := platform.NewLocalFileStore(root)
	if err != nil {
		t.Fatalf("NewLocalFileStore: %v", err)
	}
	if err := store.Put(context.Background(), "uploads/offline.txt", strings.NewReader("persisted")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close first store: %v", err)
	}

	reopened, err := platform.NewLocalFileStore(root)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	defer reopened.Close()
	file, err := reopened.Open(context.Background(), "uploads/offline.txt")
	if err != nil {
		t.Fatalf("Open persisted file: %v", err)
	}
	contents, readErr := io.ReadAll(file)
	closeErr := file.Close()
	if readErr != nil || closeErr != nil || string(contents) != "persisted" {
		t.Fatalf("persisted file = %q, read=%v close=%v", contents, readErr, closeErr)
	}
	if err := reopened.Put(context.Background(), "../escape.txt", strings.NewReader("blocked")); !errors.Is(err, platform.ErrInvalidFileKey) {
		t.Fatalf("traversal error = %v, want ErrInvalidFileKey", err)
	}
}

type probeApplication struct {
	handler http.Handler
}

func (app *probeApplication) Handler() http.Handler       { return app.handler }
func (app *probeApplication) Start(context.Context) error { return nil }
func (app *probeApplication) Stop(context.Context) error  { return nil }

func TestNativeDesktopListenerIsUnreachableFromLANInterfaces(t *testing.T) {
	app := &probeApplication{handler: http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})}
	probedAddresses := 0
	host, err := desktophost.New(desktophost.Config{
		Assets:         fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("desktop")}},
		RequestTimeout: 2 * time.Second,
	}, func(context.Context) (desktophost.Runtime, error) {
		return desktophost.Runtime{Application: app}, nil
	}, func(_ context.Context, bridge *desktophost.Bridge, _ fs.FS) error {
		bootstrap := bridge.Bootstrap()
		endpoint, parseErr := url.Parse(bootstrap.APIBaseURL)
		if parseErr != nil {
			return parseErr
		}
		_, port, splitErr := net.SplitHostPort(endpoint.Host)
		if splitErr != nil {
			return splitErr
		}
		addresses, addressErr := activeNonLoopbackIPv4Addresses()
		if addressErr != nil {
			return addressErr
		}
		for _, address := range addresses {
			probedAddresses++
			connection, dialErr := net.DialTimeout("tcp4", net.JoinHostPort(address.String(), port), 250*time.Millisecond)
			if dialErr == nil {
				_ = connection.Close()
				return fmt.Errorf("desktop API accepted a LAN connection on %s", address)
			}
		}

		request, requestErr := http.NewRequest(http.MethodGet, bootstrap.APIBaseURL+"/api/v1/native-probe", nil)
		if requestErr != nil {
			return requestErr
		}
		request.Header.Set("Origin", "wails://wails")
		request.Header.Set(bootstrap.LaunchToken.Header, bootstrap.LaunchToken.Value)
		response, requestErr := http.DefaultClient.Do(request)
		if requestErr != nil {
			return requestErr
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusNoContent {
			return fmt.Errorf("authorized loopback status = %d, want %d", response.StatusCode, http.StatusNoContent)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := host.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if probedAddresses == 0 {
		t.Skip("no active non-loopback IPv4 interface was available for a native LAN probe")
	}
}

func activeNonLoopbackIPv4Addresses() ([]net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var result []net.IP
	for _, networkInterface := range interfaces {
		if networkInterface.Flags&net.FlagUp == 0 ||
			networkInterface.Flags&net.FlagLoopback != 0 ||
			networkInterface.Flags&net.FlagPointToPoint != 0 {
			continue
		}
		addresses, addressErr := networkInterface.Addrs()
		if addressErr != nil {
			return nil, addressErr
		}
		for _, address := range addresses {
			var ip net.IP
			switch value := address.(type) {
			case *net.IPNet:
				ip = value.IP
			case *net.IPAddr:
				ip = value.IP
			}
			if ip = ip.To4(); ip != nil && !ip.IsLoopback() && !ip.IsUnspecified() {
				result = append(result, ip)
			}
		}
	}
	return result, nil
}
