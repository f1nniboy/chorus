default: build

schemas:
    glib-compile-schemas data/

build: schemas
    go build -o chorus ./cmd/chorus

run: schemas
    GSETTINGS_SCHEMA_DIR=data go run ./cmd/chorus

gen:
    go generate ./...

new-lang lang: gen
    mkdir -p data/po/{{lang}}
    msginit --no-translator --input=data/po/default.pot --locale={{lang}} --output=data/po/{{lang}}/default.po

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
    appstreamcli validate data/space.f1nn.chorus.metainfo.xml

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
