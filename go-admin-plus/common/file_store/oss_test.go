package file_store

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestOSSUploadUsesConfiguredEndpoint(t *testing.T) {
	type observedRequest struct {
		method string
		query  string
		body   string
	}

	var mu sync.Mutex
	requests := make([]observedRequest, 0, 3)
	record := func(r *http.Request, body string) {
		mu.Lock()
		defer mu.Unlock()
		requests = append(requests, observedRequest{method: r.Method, query: r.URL.RawQuery, body: body})
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		record(r, string(body))

		if r.URL.Path != "/test-bucket/baseline/fixture.txt" {
			t.Errorf("path = %q, want /test-bucket/baseline/fixture.txt", r.URL.Path)
		}

		switch {
		case r.Method == http.MethodPost && r.URL.Query().Has("uploads"):
			w.Header().Set("Content-Type", "application/xml")
			_, _ = fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><InitiateMultipartUploadResult><Bucket>test-bucket</Bucket><Key>baseline/fixture.txt</Key><UploadId>fixture-upload</UploadId></InitiateMultipartUploadResult>`)
		case r.Method == http.MethodPut && r.URL.Query().Get("uploadId") == "fixture-upload":
			w.Header().Set("ETag", `"fixture-part-etag"`)
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Query().Get("uploadId") == "fixture-upload":
			var completed struct {
				Parts []struct {
					PartNumber int    `xml:"PartNumber"`
					ETag       string `xml:"ETag"`
				} `xml:"Part"`
			}
			if err := xml.Unmarshal(body, &completed); err != nil {
				t.Errorf("parse complete multipart body: %v", err)
			}
			if len(completed.Parts) != 1 || completed.Parts[0].PartNumber != 1 {
				t.Errorf("completed parts = %#v, want one part numbered 1", completed.Parts)
			}
			w.Header().Set("Content-Type", "application/xml")
			_, _ = fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><CompleteMultipartUploadResult><Location>local</Location><Bucket>test-bucket</Bucket><Key>baseline/fixture.txt</Key><ETag>fixture-etag</ETag></CompleteMultipartUploadResult>`)
		default:
			http.Error(w, "unexpected multipart request", http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	store := new(ALiYunOSS)
	if err := store.Setup(server.URL, "test-access-key", "test-secret-key", "test-bucket"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := store.UpLoad("baseline/fixture.txt", "testdata/upload.txt"); err != nil {
		t.Fatalf("upload: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 3 {
		t.Fatalf("requests = %d, want initiate, upload part, and complete", len(requests))
	}
	if requests[0].method != http.MethodPost || requests[0].query != "uploads" {
		t.Fatalf("initiate request = %s?%s, want POST?uploads", requests[0].method, requests[0].query)
	}
	if requests[1].method != http.MethodPut || requests[1].query != "partNumber=1&uploadId=fixture-upload" {
		t.Fatalf("part request = %s?%s, want PUT?partNumber=1&uploadId=fixture-upload", requests[1].method, requests[1].query)
	}
	if requests[1].body != "speculo-characterization-fixture\n" {
		t.Fatalf("uploaded body = %q, want fixture contents", requests[1].body)
	}
	if requests[2].method != http.MethodPost || requests[2].query != "uploadId=fixture-upload" {
		t.Fatalf("complete request = %s?%s, want POST?uploadId=fixture-upload", requests[2].method, requests[2].query)
	}
}
