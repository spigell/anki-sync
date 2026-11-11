package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnkiWebSyncSendsSyncWithKey(t *testing.T) {
	unsetEnv(t, "ANKI_URL")
	t.Setenv("ANKI_API_KEY", "from-env")

	var got map[string]any
	url := fakeAnki(t, &got, `{"result": null, "error": null}`)

	// No decks or models: ankiweb-sync needs only the AnkiConnect URL.
	if err := runCLI(t, "anki_url: "+url+"\n", "ankiweb-sync"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got["action"] != "sync" || got["key"] != "from-env" {
		t.Fatalf("request = %v, want sync action with key %q", got, "from-env")
	}
}

func TestAnkiWebSyncReportsAnkiError(t *testing.T) {
	unsetEnv(t, "ANKI_URL")
	unsetEnv(t, "ANKI_API_KEY")

	var got map[string]any
	url := fakeAnki(t, &got, `{"result": null, "error": "sync: auth not configured"}`)

	err := runCLI(t, "anki_url: "+url+"\n", "ankiweb-sync")
	if err == nil || !strings.Contains(err.Error(), "sync: auth not configured") {
		t.Fatalf("err = %v, want AnkiConnect error text", err)
	}
}

func TestAnkiWebSyncTimesOut(t *testing.T) {
	unsetEnv(t, "ANKI_URL")
	unsetEnv(t, "ANKI_API_KEY")

	// Never answer, like AnkiConnect waiting on a full-sync dialog.
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		<-release
	}))
	t.Cleanup(srv.Close)
	// Cleanups run last-in first-out, so the handler returns before srv.Close waits for it.
	t.Cleanup(func() { close(release) })

	err := runCLI(t, "anki_url: "+srv.URL+"\n", "ankiweb-sync", "--timeout", "100ms")
	if err == nil || !strings.Contains(err.Error(), "timed out after 100ms") {
		t.Fatalf("err = %v, want timeout", err)
	}
}

func TestAnkiWebSyncRejectsNonPositiveTimeout(t *testing.T) {
	err := runCLI(t, "", "ankiweb-sync", "--timeout", "0s")
	if err == nil || !strings.Contains(err.Error(), "--timeout") {
		t.Fatalf("err = %v, want --timeout validation error", err)
	}
}
