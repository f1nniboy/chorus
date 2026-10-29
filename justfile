default: build

schemas:
    glib-compile-schemas data/

build: schemas
    go build -o chorus ./cmd/chorus

run: schemas
    GSETTINGS_SCHEMA_DIR=data go run ./cmd/chorus

gen:
    go generate ./...

lint:
    golangci-lint run ./...

flatpak-mod:
    flatpak-go-mod

flatpak-build:
    flatpak run --command=flathub-build org.flatpak.Builder --install space.f1nn.chorus.dev.yml

flatpak-lint:
    flatpak run --command=flatpak-builder-lint org.flatpak.Builder manifest space.f1nn.chorus.yml
    flatpak run --command=flatpak-builder-lint org.flatpak.Builder repo repo

metainfo-lint:
    appstreamcli validate --no-net data/space.f1nn.chorus.metainfo.xml

desktop-lint:
    desktop-file-validate data/space.f1nn.chorus.desktop

fix: fmt
    go fix ./...
    golangci-lint run --fix ./...

fmt:
    treefmt

update:
    go get -u ./...
    go mod tidy

clean:
    rm -f chorus
