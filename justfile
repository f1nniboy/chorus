default: build

schemas:
    glib-compile-schemas data/

build: schemas
    go build -o chorus ./cmd/chorus

run: schemas
    GSETTINGS_SCHEMA_DIR=data go run ./cmd/chorus

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
