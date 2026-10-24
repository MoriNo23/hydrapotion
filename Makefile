# Makefile — hydrapotion
# Requisitos: Go 1.24+, libgl1-mesa-dev, xorg-dev, libwayland-dev, libxkbcommon-dev (Debian/Ubuntu)

BINARY  := hydrapotion
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
PREFIX  ?= $(HOME)/.local

.PHONY: build test vet fmt run clean install-user uninstall-user release-zip

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

# Instalación de usuario sin root: binario + icono + entrada de menú (.desktop)
# con rutas absolutas, para que el lanzador siempre ejecute esta versión.
install-user: build
        install -Dm755 $(BINARY) $(PREFIX)/bin/$(BINARY)
        install -Dm644 assets/icon.png $(PREFIX)/share/$(BINARY)/icon.png
        mkdir -p $(HOME)/.local/share/applications
        sed "s|^Exec=.*|Exec=$(PREFIX)/bin/$(BINARY)|; s|^Icon=.*|Icon=$(PREFIX)/share/$(BINARY)/icon.png|" \
                packaging/$(BINARY).desktop > $(HOME)/.local/share/applications/$(BINARY).desktop
        -update-desktop-database $(HOME)/.local/share/applications
        @echo "Instalado: $(PREFIX)/bin/$(BINARY) + entrada de menú (Hydrapotion $(VERSION))"

uninstall-user:
        rm -f $(PREFIX)/bin/$(BINARY) $(PREFIX)/share/$(BINARY)/icon.png \
                $(HOME)/.local/share/applications/$(BINARY).desktop
        -update-desktop-database $(HOME)/.local/share/applications

# zip de distribución manual (binario + packaging + docs)
release-zip: build
        mkdir -p build
        tar -czf build/$(BINARY)-$(VERSION)-linux-amd64.tar.gz \
                $(BINARY) packaging/ README.md LICENSE docs/screenshot.png
