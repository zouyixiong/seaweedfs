//go:build elastic
// +build elastic

package elastic

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/seaweedfs/seaweedfs/weed/filer"

	elastic "github.com/olivere/elastic/v7"
)

func TestKvGet(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		body         string
		want         string
		wantNotFound bool
		wantErr      bool
	}{
		{
			name:         "http 404",
			status:       http.StatusNotFound,
			body:         `{"_index":".seaweedfs_kv_entries","_id":"key","found":false}`,
			wantNotFound: true,
		},
		{
			name:    "http 500",
			status:  http.StatusInternalServerError,
			body:    `{"error":{"type":"internal_server_error","reason":"failed"},"status":500}`,
			wantErr: true,
		},
		{
			name:   "successful value",
			status: http.StatusOK,
			body:   `{"_index":".seaweedfs_kv_entries","_id":"key","found":true,"_source":{"Value":"dmFsdWU="}}`,
			want:   "value",
		},
		{
			name:         "successful response not found",
			status:       http.StatusOK,
			body:         `{"_index":".seaweedfs_kv_entries","_id":"key","found":false}`,
			wantNotFound: true,
		},
		{
			name:    "malformed source",
			status:  http.StatusOK,
			body:    `{"_index":".seaweedfs_kv_entries","_id":"key","found":true,"_source":"invalid"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestElasticStore(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tt.status,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(tt.body)),
					Request:    r,
				}, nil
			}))
			value, err := store.KvGet(context.Background(), []byte("key"))

			if tt.wantNotFound {
				if !errors.Is(err, filer.ErrKvNotFound) {
					t.Fatalf("KvGet() error = %v, want %v", err, filer.ErrKvNotFound)
				}
				return
			}
			if tt.wantErr {
				if err == nil {
					t.Fatal("KvGet() error = nil, want non-nil")
				}
				if errors.Is(err, filer.ErrKvNotFound) {
					t.Fatalf("KvGet() error = %v, must not be %v", err, filer.ErrKvNotFound)
				}
				return
			}
			if err != nil {
				t.Fatalf("KvGet() error = %v", err)
			}
			if string(value) != tt.want {
				t.Fatalf("KvGet() value = %q, want %q", value, tt.want)
			}
		})
	}
}

func TestKvGetConnectionError(t *testing.T) {
	connectionErr := errors.New("connection refused")
	store := newTestElasticStore(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, connectionErr
	}))
	_, err := store.KvGet(context.Background(), []byte("key"))
	if err == nil {
		t.Fatal("KvGet() error = nil, want connection error")
	}
	if !errors.Is(err, connectionErr) {
		t.Fatalf("KvGet() error = %v, want wrapped connection error", err)
	}
	if errors.Is(err, filer.ErrKvNotFound) {
		t.Fatalf("KvGet() error = %v, must not be %v", err, filer.ErrKvNotFound)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newTestElasticStore(t *testing.T, transport http.RoundTripper) *ElasticStore {
	t.Helper()
	client, err := elastic.NewClient(
		elastic.SetURL("http://elasticsearch.test"),
		elastic.SetHttpClient(&http.Client{Transport: transport}),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("create elastic client: %v", err)
	}
	return &ElasticStore{client: client}
}
