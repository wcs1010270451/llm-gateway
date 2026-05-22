Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Push-Location "$PSScriptRoot\..\backend"
try {
    if (-not $env:DATABASE_URL) {
        $env:DATABASE_URL = "postgres://llm_gateway:llm_gateway@127.0.0.1:55432/llm_gateway?sslmode=disable"
    }
    if (-not $env:HTTP_ADDR) {
        $env:HTTP_ADDR = "127.0.0.1:3212"
    }
    if (-not $env:PROVIDER_KEY_SECRET) {
        $env:PROVIDER_KEY_SECRET = "change-me-provider-key-secret"
    }
    go run ./cmd/server
} finally {
    Pop-Location
}
