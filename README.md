# anki-sync

anki-sync is a command line tool for synchronising decks, models and notes with an Anki instance through the AnkiConnect API. Data is described in YAML and the CLI uploads notes in parallel while updating models and decks as needed.

## Installation

Prebuilt binaries are available on the [GitHub releases page](https://github.com/spigell/anki-sync/releases). Linux and macOS builds for both `amd64` and `arm64` are produced with [goreleaser](https://goreleaser.com) (see `.goreleaser.yaml` for details). Download the archive for your platform, extract it, and place `anki-sync` somewhere on your `PATH`.

To build from source instead, run:

```bash
go install github.com/spigell/anki-sync@latest
```

Requires Go 1.23 or newer.

## Development

1. Run `make build` to compile the binary.
2. Start an Anki instance with AnkiConnect via `docker-compose up`.
3. Execute `make integration-tests` to run the tests against the container.

## Roadmap

1. fix linter issues
2. add support for images and audio

See `anki-sync-example.yaml` for a sample configuration.

For a cool MCP-based server that talks to Anki, see
[anki-mcp-server](https://github.com/CamdenClark/anki-mcp-server).

