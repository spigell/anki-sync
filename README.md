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
