default: build

schemas:
    glib-compile-schemas data/

ldflags := "-X 'github.com/f1nniboy/chorus/internal/meta.Version=" + `git describe --tags --abbrev=0 2>/dev/null || echo dev` + "'"

build: schemas
    go build -ldflags="{{ldflags}}" -o chorus ./cmd/chorus

run: schemas
    GSETTINGS_SCHEMA_DIR=data go run -ldflags="{{ldflags}}" ./cmd/chorus

gen:
    go generate ./...

potgen: gen
    #!/usr/bin/env bash
    set -euo pipefail
    for po in data/po/*/*.po; do
        msgmerge --update --backup=none "$po" data/po/default.pot
    done

new-lang lang: gen
    mkdir -p data/po/{{lang}}
    msginit --no-translator --input=data/po/default.pot --locale={{lang}} --output=data/po/{{lang}}/default.po

lint:
    golangci-lint run ./...

flatpak-build:
    flatpak run --command=flathub-build org.flatpak.Builder --install space.f1nn.chorus.dev.yml

flatpak-lint:
    flatpak run --command=flatpak-builder-lint org.flatpak.Builder manifest space.f1nn.chorus.yml
    flatpak run --command=flatpak-builder-lint org.flatpak.Builder repo repo

fix: fmt
    go fix ./...
    golangci-lint run --fix ./...

fmt:
    treefmt

test:
    go test ./...

coverage:
    go test ./... -coverprofile=/tmp/chorus.out && go tool cover -func=/tmp/chorus.out

update:
    go get -u ./...
    go mod tidy

clean:
    rm -f chorus
