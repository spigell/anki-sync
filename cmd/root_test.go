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

// runVersion runs `anki-sync version` against a fake AnkiConnect and returns the sent key.
func runVersion(t *testing.T, configKey string) (string, bool) {
	t.Helper()

	viper.Reset()
	*Config = AppConfig{}

	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Write([]byte(`{"result": 6, "error": null}`))
	}))
	defer srv.Close()

	cfg := filepath.Join(t.TempDir(), "anki-sync.yaml")
	content := "anki_url: " + srv.URL + "\n"
	if configKey != "" {
		content += "anki_api_key: " + configKey + "\n"
	}
	if err := os.WriteFile(cfg, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	root := NewRootCmd(context.Background(), &Logger{Instance: logging.NewLogger(level), Level: level})
	root.SetArgs([]string{"--config", cfg, "version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	key, ok := got["key"].(string)
	return key, ok
}

func TestAPIKeyFromConfig(t *testing.T) {
	t.Setenv("ANKI_API_KEY", "")
	os.Unsetenv("ANKI_API_KEY")

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
	t.Setenv("ANKI_API_KEY", "")
	os.Unsetenv("ANKI_API_KEY")

	if key, ok := runVersion(t, ""); ok {
		t.Fatalf("request has key %q, want none", key)
	}
}
