package jobs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPJobStopsAnInFlightRequestWhenContextIsCancelled(t *testing.T) {
	requestStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		close(requestStarted)
		<-request.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	job := &HttpJob{JobCore: JobCore{InvokeTarget: server.URL, Name: "cancel-test"}}
	job.withContext(ctx)
	done := make(chan struct{})
	go func() {
		job.Run()
		close(done)
	}()
	select {
	case <-requestStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP job did not start its request")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP job did not stop after context cancellation")
	}
}
