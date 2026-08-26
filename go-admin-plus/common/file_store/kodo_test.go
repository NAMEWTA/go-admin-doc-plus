package file_store

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

func TestKODOUploadUsesConfiguredEndpoint(t *testing.T) {
	const wantBody = "speculo-characterization-fixture\n"
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("parse multipart form: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if got := r.FormValue("key"); got != "baseline/fixture.txt" {
			t.Errorf("key = %q, want baseline/fixture.txt", got)
		}
		if got := r.FormValue("token"); !strings.HasPrefix(got, "test-access-key:") {
			t.Errorf("token = %q, want test access-key prefix", got)
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("multipart file: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		body, err := io.ReadAll(file)
		if err != nil {
			t.Errorf("read multipart file: %v", err)
		}
		if string(body) != wantBody {
			t.Errorf("body = %q, want %q", body, wantBody)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-reqid", "fixture-request-id")
		_ = json.NewEncoder(w).Encode(storage.PutRet{Key: "baseline/fixture.txt", Hash: "fixture-hash"})
	}))
	t.Cleanup(server.Close)

	store := &QiNiuKODO{
		Client:     qbox.NewMac("test-access-key", "test-secret-key"),
		BucketName: "test-bucket",
		cfg: storage.Config{
			UpHost:   server.URL,
			UseHTTPS: false,
		},
	}

	if err := store.UpLoad("baseline/fixture.txt", "testdata/upload.txt"); err != nil {
		t.Fatalf("upload: %v", err)
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("requests = %d, want 1", got)
	}
}

func TestKODOGetTempTokenIsLocalAndScoped(t *testing.T) {
	store := (&OXS{
		AccessKeyID:     "test-access-key",
		AccessKeySecret: "test-secret-key",
		BucketName:      "test-bucket",
	}).Setup(QiNiuKodo, ClientOption{"Expires": uint64(60)})

	token, err := store.GetTempToken()
	if err != nil {
		t.Fatalf("get temp token: %v", err)
	}
	if !strings.HasPrefix(token, "test-access-key:") {
		t.Fatalf("token = %q, want test access-key prefix", token)
	}
}
