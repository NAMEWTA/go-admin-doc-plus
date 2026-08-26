package kernel_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"go-admin/internal/app/kernel"
)

func TestKernelStartsThenDrainsInDeterministicOrder(t *testing.T) {
	var events []string
	var runtime *kernel.Kernel
	unit := func(name string) kernel.Lifecycle {
		return kernel.Lifecycle{
			Name: name,
			Start: func(context.Context) error {
				events = append(events, "start:"+name)
				return nil
			},
			Drain: func(context.Context) error {
				if runtime.Snapshot().State != kernel.StateDraining {
					t.Fatalf("state during drain = %q", runtime.Snapshot().State)
				}
				events = append(events, "drain:"+name)
				return nil
			},
			Stop: func(context.Context) error {
				events = append(events, "stop:"+name)
				return nil
			},
		}
	}
	var err error
	runtime, err = kernel.New(unit("database"), unit("http"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if state := runtime.Snapshot().State; state != kernel.StateStarting {
		t.Fatalf("initial state = %q", state)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if state := runtime.Snapshot().State; state != kernel.StateReady {
		t.Fatalf("state after Start() = %q", state)
	}
	if err := runtime.Drain(context.Background()); err != nil {
		t.Fatalf("Drain() error = %v", err)
	}
	if state := runtime.Snapshot().State; state != kernel.StateStopped {
		t.Fatalf("state after Drain() = %q", state)
	}
	want := []string{"start:database", "start:http", "drain:http", "stop:http", "drain:database", "stop:database"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestDependencyFailureIsNotReadyAndCleansEveryLifecycle(t *testing.T) {
	var events []string
	var runtime *kernel.Kernel
	unit := func(name string, drainErr error) kernel.Lifecycle {
		return kernel.Lifecycle{
			Name:  name,
			Start: func(context.Context) error { return nil },
			Drain: func(context.Context) error {
				if snapshot := runtime.Snapshot(); snapshot.State != kernel.StateFailed || snapshot.Failure != kernel.FailureDependency {
					t.Fatalf("snapshot during dependency cleanup = %#v", snapshot)
				}
				events = append(events, "drain:"+name)
				return drainErr
			},
			Stop: func(context.Context) error {
				events = append(events, "stop:"+name)
				return nil
			},
		}
	}
	secretFailure := errors.New("database password=do-not-expose")
	var err error
	runtime, err = kernel.New(unit("database", nil), unit("worker", secretFailure))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	err = runtime.DependencyFailed(context.Background())
	if !errors.Is(err, secretFailure) {
		t.Fatalf("DependencyFailed() error = %v", err)
	}
	if strings.Contains(err.Error(), "do-not-expose") || strings.Contains(err.Error(), "password") {
		t.Fatalf("DependencyFailed() leaked cleanup detail: %v", err)
	}
	snapshot := runtime.Snapshot()
	if snapshot.State != kernel.StateFailed || snapshot.Failure != kernel.FailureDependency || len(snapshot.Started) != 0 {
		t.Fatalf("Snapshot() = %#v", snapshot)
	}
	if strings.Contains(strings.Join(snapshot.Started, ","), "do-not-expose") {
		t.Fatal("Snapshot() leaked dependency failure detail")
	}
	want := []string{"drain:worker", "stop:worker", "drain:database", "stop:database"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestStartFailureStopsPartialRuntimeAndNeverBecomesReady(t *testing.T) {
	var events []string
	secretFailure := errors.New("open /private/data: password=do-not-log")
	first := kernel.Lifecycle{
		Name:  "database",
		Start: func(context.Context) error { events = append(events, "start:database"); return nil },
		Drain: func(context.Context) error { events = append(events, "drain:database"); return nil },
		Stop:  func(context.Context) error { events = append(events, "stop:database"); return nil },
	}
	second := kernel.Lifecycle{
		Name:  "worker",
		Start: func(context.Context) error { events = append(events, "start:worker"); return secretFailure },
		Drain: func(context.Context) error { events = append(events, "drain:worker"); return nil },
		Stop:  func(context.Context) error { events = append(events, "stop:worker"); return nil },
	}
	runtime, err := kernel.New(first, second)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	err = runtime.Start(context.Background())
	if !errors.Is(err, secretFailure) {
		t.Fatalf("Start() error = %v", err)
	}
	if strings.Contains(err.Error(), "do-not-log") || strings.Contains(err.Error(), "/private/data") {
		t.Fatalf("Start() leaked dependency detail: %v", err)
	}
	snapshot := runtime.Snapshot()
	if snapshot.State != kernel.StateFailed || snapshot.Failure != kernel.FailureStartup || len(snapshot.Started) != 0 {
		t.Fatalf("Snapshot() = %#v", snapshot)
	}
	want := []string{"start:database", "start:worker", "stop:database"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
