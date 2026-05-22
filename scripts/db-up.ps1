Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$root = Resolve-Path "$PSScriptRoot\.."
Push-Location $root
try {
    docker compose up -d postgres
    docker compose ps postgres
} finally {
    Pop-Location
}
