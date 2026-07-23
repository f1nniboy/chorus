#!/usr/bin/env bash
set -euo pipefail

for f in "$@"; do
    xmllint --format "$f" -o "$f"
done
