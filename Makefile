.PHONY: build
build:
	@VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev") && \
	go build -ldflags "-X main.version=$$VERSION" -o gmail-uptimekuma-alert-cleaner

.PHONY: build-all
build-all:
	@VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev") && \
	echo "Building for all platforms..." && \
	GOOS=linux   GOARCH=amd64 go build -ldflags "-X main.version=$$VERSION" -o gmail-uptimekuma-alert-cleaner-linux-amd64 && \
	GOOS=linux   GOARCH=arm64 go build -ldflags "-X main.version=$$VERSION" -o gmail-uptimekuma-alert-cleaner-linux-arm64 && \
	GOOS=darwin  GOARCH=amd64 go build -ldflags "-X main.version=$$VERSION" -o gmail-uptimekuma-alert-cleaner-darwin-amd64 && \
	GOOS=darwin  GOARCH=arm64 go build -ldflags "-X main.version=$$VERSION" -o gmail-uptimekuma-alert-cleaner-darwin-arm64 && \
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$$VERSION" -o gmail-uptimekuma-alert-cleaner-windows-amd64.exe && \
	echo "Done. Built binaries:" && \
	ls -lh gmail-uptimekuma-alert-cleaner-*

.PHONY: test
test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: clean
clean:
	rm -f gmail-uptimekuma-alert-cleaner gmail-uptimekuma-alert-cleaner-* coverage.txt

.PHONY: version
version:
	@echo $$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
