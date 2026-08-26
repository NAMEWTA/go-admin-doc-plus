package desktop

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"sync"
	"testing"
	"testing/fstest"
	"time"
)

type fakeApplication struct {
	handler  http.Handler
	started  bool
	stopped  bool
	startErr error
}

func (app *fakeApplication) Handler() http.Handler { return app.handler }
func (app *fakeApplication) Start(context.Context) error {
	app.started = true
	return app.startErr
}
func (app *fakeApplication) Stop(context.Context) error {
	app.stopped = true
	return nil
}

func TestHostRunsOnlyAfterAssetsApplicationAndLoopbackAreReady(t *testing.T) {
	app := &fakeApplication{handler: http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"code":200,"data":{"name":"offline"},"msg":"ok"}`))
	})}
	closed := false
	windowRan := false
	assets := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("desktop")}}
	host, err := New(Config{
		Assets:         assets,
		Version:        "test-version",
		RequestTimeout: 2 * time.Second,
	}, func(context.Context) (Runtime, error) {
		return Runtime{Application: app, Close: func(context.Context) error { closed = true; return nil }}, nil
	}, func(_ context.Context, bridge *Bridge, received fs.FS) error {
		windowRan = true
		if _, err := fs.Stat(received, "index.html"); err != nil {
			return err
		}
		payload := bridge.Bootstrap()
		if payload.Version != "test-version" || payload.Capabilities.HostProfile != "desktop" || !payload.Capabilities.Offline {
			t.Fatalf("bootstrap = %#v", payload)
		}
		if !payload.Security.LoopbackOnly || payload.Security.LaunchTokenHeader != LaunchTokenHeader || len(payload.Security.AllowedOrigins) != 2 {
			t.Fatalf("bootstrap security = %#v", payload.Security)
		}
		request, err := http.NewRequest(http.MethodGet, payload.APIBaseURL+"/api/v1/demo", nil)
		if err != nil {
			return err
		}
		request.Header.Set("Origin", "wails://wails")
		request.Header.Set(payload.LaunchToken.Header, payload.LaunchToken.Value)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("HTTP status = %d", response.StatusCode)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := host.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !windowRan || !app.started || !app.stopped || !closed {
		t.Fatalf("lifecycle window=%v started=%v stopped=%v closed=%v", windowRan, app.started, app.stopped, closed)
	}
}

type fixedAddressListener struct {
	address net.Addr
}

func (listener *fixedAddressListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (listener *fixedAddressListener) Close() error              { return nil }
func (listener *fixedAddressListener) Addr() net.Addr            { return listener.address }

func TestHostRejectsNonLoopbackListenerBeforeStartingApplication(t *testing.T) {
	app := &fakeApplication{handler: http.NotFoundHandler()}
	windowRan := false
	host, err := New(Config{Assets: fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("ok")}}}, func(context.Context) (Runtime, error) {
		return Runtime{Application: app}, nil
	}, func(context.Context, *Bridge, fs.FS) error {
		windowRan = true
		return nil
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	host.listen = func(context.Context) (net.Listener, error) {
		return &fixedAddressListener{address: &net.TCPAddr{IP: net.IPv4zero, Port: 8080}}, nil
	}
	if err := host.Run(context.Background()); !errors.Is(err, ErrUnsafeListener) {
		t.Fatalf("Run error = %v, want ErrUnsafeListener", err)
	}
	if app.started || windowRan {
		t.Fatalf("started=%v window=%v, want both false", app.started, windowRan)
	}
}

func TestHostFailsBeforeWindowWhenAssetsAreMissing(t *testing.T) {
	windowRan := false
	_, err := New(Config{Assets: fstest.MapFS{}}, func(context.Context) (Runtime, error) {
		return Runtime{}, nil
	}, func(context.Context, *Bridge, fs.FS) error { windowRan = true; return nil })
	if !errors.Is(err, ErrMissingAssets) {
		t.Fatalf("New error = %v, want ErrMissingAssets", err)
	}
	if windowRan {
		t.Fatal("window ran despite missing assets")
	}
}

func TestHostClosesRuntimeWhenApplicationStartFails(t *testing.T) {
	app := &fakeApplication{handler: http.NotFoundHandler(), startErr: errors.New("start failed")}
	closed := false
	windowRan := false
	host, err := New(Config{Assets: fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("ok")}}}, func(context.Context) (Runtime, error) {
		return Runtime{Application: app, Close: func(context.Context) error { closed = true; return nil }}, nil
	}, func(context.Context, *Bridge, fs.FS) error { windowRan = true; return nil })
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := host.Run(context.Background()); err == nil || !errors.Is(err, app.startErr) {
		t.Fatalf("Run error = %v", err)
	}
	if windowRan || !closed {
		t.Fatalf("window=%v closed=%v", windowRan, closed)
	}
}

type lifecycleRecorder struct {
	mu     sync.Mutex
	events []string
}

func (recorder *lifecycleRecorder) add(event string) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	recorder.events = append(recorder.events, event)
}

func (recorder *lifecycleRecorder) snapshot() []string {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return append([]string(nil), recorder.events...)
}

type orderedApplication struct {
	recorder *lifecycleRecorder
	started  chan struct{}
	release  chan struct{}
}

func (app *orderedApplication) Handler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		app.recorder.add("request-start")
		close(app.started)
		<-app.release
		app.recorder.add("request-end")
		writer.WriteHeader(http.StatusNoContent)
	})
}

func (app *orderedApplication) Start(context.Context) error {
	app.recorder.add("application-start")
	return nil
}

func (app *orderedApplication) Stop(context.Context) error {
	app.recorder.add("application-stop")
	return nil
}

func TestHostDrainsRequestsBeforeStoppingApplicationAndRuntime(t *testing.T) {
	recorder := &lifecycleRecorder{}
	app := &orderedApplication{
		recorder: recorder,
		started:  make(chan struct{}),
		release:  make(chan struct{}),
	}
	requestDone := make(chan error, 1)
	host, err := New(Config{
		Assets:          fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("ok")}},
		RequestTimeout:  2 * time.Second,
		ShutdownTimeout: 2 * time.Second,
	}, func(context.Context) (Runtime, error) {
		return Runtime{
			Application: app,
			Close: func(context.Context) error {
				recorder.add("runtime-close")
				return nil
			},
		}, nil
	}, func(_ context.Context, bridge *Bridge, _ fs.FS) error {
		request, requestErr := http.NewRequest(http.MethodGet, bridge.Bootstrap().APIBaseURL+"/api/v1/drain", nil)
		if requestErr != nil {
			return requestErr
		}
		request.Header.Set("Origin", "wails://wails")
		request.Header.Set(bridge.Bootstrap().LaunchToken.Header, bridge.Bootstrap().LaunchToken.Value)
		go func() {
			response, doErr := http.DefaultClient.Do(request)
			if doErr == nil {
				doErr = response.Body.Close()
			}
			requestDone <- doErr
		}()
		select {
		case <-app.started:
		case <-time.After(time.Second):
			return errors.New("request did not reach application")
		}
		recorder.add("window-return")
		time.AfterFunc(25*time.Millisecond, func() { close(app.release) })
		return nil
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := host.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := <-requestDone; err != nil {
		t.Fatalf("request: %v", err)
	}
	want := []string{"application-start", "request-start", "window-return", "request-end", "application-stop", "runtime-close"}
	got := recorder.snapshot()
	if len(got) != len(want) {
		t.Fatalf("lifecycle = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("lifecycle = %v, want %v", got, want)
		}
	}
}
