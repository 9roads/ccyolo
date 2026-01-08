VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/9roads/ccyolo/cmd.Version=$(VERSION)"

.PHONY: all build install clean test release

all: build

build:
	go build $(LDFLAGS) -o ccyolo .

install: build
	go install $(LDFLAGS) .

clean:
	rm -f ccyolo
	rm -rf dist/

test:
	go test ./...

# Cross-platform builds
release: clean
	mkdir -p dist

	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/ccyolo-darwin-x86_64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/ccyolo-darwin-arm64 .

	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/ccyolo-linux-x86_64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/ccyolo-linux-arm64 .

	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/ccyolo-windows-x86_64.exe .

	# Create checksums
	cd dist && shasum -a 256 * > checksums.txt

	@echo "Release binaries in dist/"
