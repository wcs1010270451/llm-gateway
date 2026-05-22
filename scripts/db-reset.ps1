Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

& "$PSScriptRoot\db-migrate-down.ps1"
& "$PSScriptRoot\db-migrate-up.ps1"
