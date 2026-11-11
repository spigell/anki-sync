package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/spigell/anki-sync/internal/logging"
	"go.uber.org/zap"
)

// fakeAnki starts a fake AnkiConnect that stores each decoded request in got
// and replies with reply. It returns the server URL.
func fakeAnki(t *testing.T, got *map[string]any, reply string) string {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(got); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Write([]byte(reply))
	}))
	t.Cleanup(srv.Close)

	return srv.URL
}

// runCLI runs anki-sync with args and a config file holding configBody.
func runCLI(t *testing.T, configBody string, args ...string) error {
	t.Helper()

	viper.Reset()
	*Config = AppConfig{}

	cfg := filepath.Join(t.TempDir(), "anki-sync.yaml")
	if err := os.WriteFile(cfg, []byte(configBody), 0o600); err != nil {
		t.Fatal(err)
	}

	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	root := NewRootCmd(context.Background(), &Logger{Instance: logging.NewLogger(level), Level: level})
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs(append([]string{"--config", cfg}, args...))
	return root.Execute()
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	t.Setenv(key, "")
	os.Unsetenv(key)
}

// runVersion runs `anki-sync version` against a fake AnkiConnect and returns the sent key.
func runVersion(t *testing.T, configKey string) (string, bool) {
	t.Helper()

	var got map[string]any
	content := "anki_url: " + fakeAnki(t, &got, `{"result": 6, "error": null}`) + "\n"
	if configKey != "" {
		content += "anki_api_key: " + configKey + "\n"
	}

	if err := runCLI(t, content, "version"); err != nil {
		t.Fatalf("execute: %v", err)
	}

	key, ok := got["key"].(string)
	return key, ok
}

func TestAPIKeyFromConfig(t *testing.T) {
	unsetEnv(t, "ANKI_API_KEY")

	key, ok := runVersion(t, "from-config")
	if !ok || key != "from-config" {
		t.Fatalf("key = %q (present %v), want %q", key, ok, "from-config")
	}
}

func TestAPIKeyEnvOverridesConfig(t *testing.T) {
	t.Setenv("ANKI_API_KEY", "from-env")

	key, ok := runVersion(t, "from-config")
	if !ok || key != "from-env" {
		t.Fatalf("key = %q (present %v), want %q", key, ok, "from-env")
	}
}

func TestAPIKeyOmittedWhenUnset(t *testing.T) {
	unsetEnv(t, "ANKI_API_KEY")

	if key, ok := runVersion(t, ""); ok {
		t.Fatalf("request has key %q, want none", key)
	}
}

func TestAnkiURLEnvOverridesConfig(t *testing.T) {
	unsetEnv(t, "ANKI_API_KEY")

	var got map[string]any
	t.Setenv("ANKI_URL", fakeAnki(t, &got, `{"result": 6, "error": null}`))

	// The config URL is unreachable, so only ANKI_URL can reach the fake AnkiConnect.
	if err := runCLI(t, "anki_url: http://127.0.0.1:1\n", "version"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got["action"] != "version" {
		t.Fatalf("request = %v, want version action", got)
	}
}

func TestAnkiURLEnvIgnoredWhenEmpty(t *testing.T) {
	unsetEnv(t, "ANKI_API_KEY")
	t.Setenv("ANKI_URL", "")

	var got map[string]any
	if err := runCLI(t, "anki_url: "+fakeAnki(t, &got, `{"result": 6, "error": null}`)+"\n", "version"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got["action"] != "version" {
		t.Fatalf("request = %v, want version action", got)
	}
}
