INTEGRATION_TESTS := ./tests/integration
VERSION := $(shell git describe --tags --always --dirty)

build:
	go build -ldflags="-w -s -X github.com/spigell/anki-sync/cmd.CliVersion=$(VERSION)"

integration-tests:
	./anki-sync version && \
		./anki-sync sync --config $(INTEGRATION_TESTS)/anki-sync-ci.yaml --models $(INTEGRATION_TESTS)/testdata/models.yaml --log-level debug --dry-run && \
		./anki-sync sync --config $(INTEGRATION_TESTS)/anki-sync-ci.yaml --models $(INTEGRATION_TESTS)/testdata/models.yaml --log-level debug
