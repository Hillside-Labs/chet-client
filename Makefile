SRC := $(shell find . -name '*.go')

chet: $(SRC)
	go mod tidy
	go build ./cmd/chet

install:
	go install ./cmd/chet

test: $(SRC)
	go mod tidy
	go test ./...

.PHONY: goreleaser
goreleaser: nfpm
	go install github.com/goreleaser/goreleaser@latest
