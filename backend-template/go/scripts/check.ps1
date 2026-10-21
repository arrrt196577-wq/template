param([switch]$Integration, [switch]$Security)
$ErrorActionPreference = 'Stop'
Push-Location (Join-Path $PSScriptRoot '..')
try {
    $sources = @(Get-ChildItem -Path cmd,internal,migrations,tests -Filter '*.go' -Recurse -File | ForEach-Object FullName)
    $formatter = Join-Path (go env GOROOT) 'bin/gofmt'
    if ($LASTEXITCODE -ne 0) { throw 'Go toolchain is unavailable' }
    $unformatted = @(& $formatter -l $sources)
    if ($LASTEXITCODE -ne 0 -or $unformatted.Count -gt 0) { throw "Go source formatting failed: $unformatted" }
    go mod verify
    if ($LASTEXITCODE -ne 0) { throw 'Module verification failed' }
    go mod tidy -diff
    if ($LASTEXITCODE -ne 0) { throw 'go.mod or go.sum is not tidy' }
    go vet ./...
    if ($LASTEXITCODE -ne 0) { throw 'go vet failed' }
    go test -count=1 ./...
    if ($LASTEXITCODE -ne 0) { throw 'Unit tests failed' }
    if ($Integration) {
        if (-not $env:TEST_DATABASE_DSN) { throw 'TEST_DATABASE_DSN is required' }
        go test -tags=integration -count=1 -timeout=2m ./tests/integration
        if ($LASTEXITCODE -ne 0) { throw 'PostgreSQL integration tests failed' }
    }
    if ($Security) {
        go tool govulncheck ./...
        if ($LASTEXITCODE -ne 0) { throw 'Vulnerability check failed' }
    }
    New-Item -ItemType Directory -Force -Path bin | Out-Null
    go build -trimpath -o bin/ ./cmd/server ./cmd/manage
    if ($LASTEXITCODE -ne 0) { throw 'Build failed' }
} finally {
    Pop-Location
}
