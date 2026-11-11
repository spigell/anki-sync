package anki

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func captureRequest(t *testing.T, opts ...Option) map[string]any {
	t.Helper()

	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Write([]byte(`{"result": 6, "error": null}`))
	}))
	defer srv.Close()

	if _, err := NewClient(srv.URL, opts...).GetVersion(context.Background()); err != nil {
		t.Fatalf("GetVersion: %v", err)
	}

	return got
}

func TestRequestCarriesAPIKey(t *testing.T) {
	got := captureRequest(t, WithAPIKey("secret"))

	if got["key"] != "secret" {
		t.Fatalf("key = %v, want %q", got["key"], "secret")
	}
}

func TestRequestOmitsAPIKeyWhenUnset(t *testing.T) {
	for name, opts := range map[string][]Option{
		"no option": nil,
		"empty key": {WithAPIKey("")},
	} {
		t.Run(name, func(t *testing.T) {
			got := captureRequest(t, opts...)

			if _, ok := got["key"]; ok {
				t.Fatalf("request has key field: %v", got)
			}
		})
	}
}
