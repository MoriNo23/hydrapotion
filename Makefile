# Makefile — hydrapotion
# Requisitos: Go 1.24+, libgl1-mesa-dev, xorg-dev, libwayland-dev, libxkbcommon-dev (Debian/Ubuntu)

BINARY  := hydrapotion
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test vet fmt run clean release-zip

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w main.go assets.go headless.go theme.go hydrograph.go \
		internal/goal/*.go internal/reminder/*.go

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)

# zip de distribución manual (binario + packaging + docs)
release-zip: build
	mkdir -p build
	tar -czf build/$(BINARY)-$(VERSION)-linux-amd64.tar.gz \
		$(BINARY) packaging/ README.md LICENSE docs/screenshot.png
