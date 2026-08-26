package file_store

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

func TestOBSUploadUsesConfiguredEndpoint(t *testing.T) {
	const wantBody = "speculo-characterization-fixture\n"
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/test-bucket/baseline/fixture.txt" {
			t.Errorf("path = %q, want /test-bucket/baseline/fixture.txt", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		if string(body) != wantBody {
			t.Errorf("body = %q, want %q", body, wantBody)
		}
		w.Header().Set("ETag", "fixture-etag")
		w.Header().Set("x-obs-request-id", "fixture-request-id")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client, err := obs.New(
		"test-access-key",
		"test-secret-key",
		server.URL,
		obs.WithPathStyle(true),
		obs.WithSignature(obs.SignatureV2),
	)
	if err != nil {
		t.Fatalf("create OBS client: %v", err)
	}
	defer client.Close()

	store := &HuaWeiOBS{Client: client, BucketName: "test-bucket"}
	if err := store.UpLoad("baseline/fixture.txt", "testdata/upload.txt"); err != nil {
		t.Fatalf("upload: %v", err)
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("requests = %d, want 1", got)
	}
}

func TestOBSSetupRejectsMissingEndpoint(t *testing.T) {
	store := new(HuaWeiOBS)
	if err := store.Setup("", "test-access-key", "test-secret-key", "test-bucket"); err == nil {
		t.Fatal("setup accepted an empty endpoint")
	}
	if store.Client != nil {
		t.Fatal("setup retained a client after rejecting the endpoint")
	}
}
