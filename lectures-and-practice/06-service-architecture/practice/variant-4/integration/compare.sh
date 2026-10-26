#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

mkdir "$tmp/old"
git show "${OLD_REF:-8b12e71}:./main.go" > "$tmp/old/main.go"
cp go.mod "$tmp/old/"
(cd "$tmp/old" && go build -o "$tmp/app-old" .)
go build -o "$tmp/app-new" .

for v in old new; do
	echo "=== $v ==="
	APP_BIN="$tmp/app-$v" go test -count=1 -v ./integration/ | grep -E '^(---|ok|FAIL|\s+integration_test)'
done
