# anki-sync

anki-sync is a command line tool for synchronising decks, models and notes with an Anki instance through the AnkiConnect API. Data is described in YAML and the CLI uploads notes in parallel while updating models and decks as needed.

## Installation

Prebuilt binaries are available on the [GitHub releases page](https://github.com/spigell/anki-sync/releases). Linux and macOS builds for both `amd64` and `arm64` are produced with [goreleaser](https://goreleaser.com) (see `.goreleaser.yaml` for details). Download the archive for your platform, extract it, and place `anki-sync` somewhere on your `PATH`.

To build from source instead, run:

```bash
go install github.com/spigell/anki-sync@latest
```

Requires Go 1.23 or newer.

See `anki-sync-example.yaml` for a sample configuration.

### AnkiConnect API key

If AnkiConnect is configured with an `apiKey`, anki-sync sends it as the `key` field of every request. Set it with the `ANKI_API_KEY` environment variable (recommended, so the key never lands in a committed config file):

```bash
export ANKI_API_KEY=your_api_key_here
anki-sync sync --config anki-sync.yaml
```

or with `anki_api_key` in the config file. When both are set, `ANKI_API_KEY` wins. When neither is set, no key is sent. anki-sync never logs the key.

### AnkiConnect URL

The AnkiConnect URL comes from `anki_url` in the config file (default `http://127.0.0.1:8765`). Set the `ANKI_URL` environment variable to override it without editing the config; an empty `ANKI_URL` is ignored. An explicit `--anki-url` flag still wins over both. Run with `--log-level debug` to see the URL in use.

```bash
ANKI_URL=http://localhost:8765 anki-sync sync --config anki-sync.yaml
```

### Syncing with AnkiWeb

`anki-sync ankiweb-sync` asks Anki to sync the open profile with AnkiWeb through AnkiConnect's `sync` action. It uses the same URL and API key settings as `sync`, needs no deck or model files, and waits up to `--timeout` (default `120s`) for the sync to finish:

```bash
anki-sync ankiweb-sync --timeout 5m
```

It exits non-zero with AnkiConnect's error text when the sync fails, and with a timeout error when Anki does not answer in time.

The first sync of a profile may require choosing between uploading and downloading a full collection in the Anki GUI. AnkiConnect cannot answer that choice, so the command does not choose for you: it times out or fails with AnkiConnect's error. Do the first sync from the Anki GUI.

## Running in a shell MCP

anki-sync can run as the only allowed executable of a shell MCP server such as [mcp-shell](https://github.com/sonirico/mcp-shell), where there are no pipes, redirects or shell quoting. Every option is a plain flag or path, so commands look like:

```text
anki-sync sync --config /decks/anki-sync.yaml
anki-sync ankiweb-sync --timeout 5m
anki-sync version
```

Set `ANKI_URL` (and `ANKI_API_KEY` if AnkiConnect needs one) in the server's environment to point anki-sync at a local Anki, such as `http://localhost:8765` in a sidecar, while reading shared, read-only configs that keep their own `anki_url`.

## Development

1. Run `make build` to compile the binary.
2. Start an Anki instance with AnkiConnect via `docker-compose up`.
3. Execute `make integration-tests` to run the tests against the container.

## Roadmap

1. fix linter issues
2. add support for images and audio
3. add a validation layer

## Bonus
For a cool MCP-based server that talks to Anki, see
[anki-mcp-server](https://github.com/CamdenClark/anki-mcp-server).
