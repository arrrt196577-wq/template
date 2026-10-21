#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/.."
formatter="$(go env GOROOT)/bin/gofmt"
unformatted="$("$formatter" -l cmd internal migrations tests)"
if [ -n "$unformatted" ]; then
  echo 'Go source formatting failed' >&2
  exit 1
fi
go mod verify
go mod tidy -diff
go vet ./...
go test -count=1 ./...
for flag in "$@"; do
  case "$flag" in
    --integration)
      : "${TEST_DATABASE_DSN:?TEST_DATABASE_DSN is required}"
      go test -tags=integration -count=1 -timeout=2m ./tests/integration
      ;;
    --security) go tool govulncheck ./... ;;
    *) echo "Unknown check option: $flag" >&2; exit 2 ;;
  esac
done
mkdir -p bin
go build -trimpath -o bin/ ./cmd/server ./cmd/manage
