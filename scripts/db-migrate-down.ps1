Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. "$PSScriptRoot\db-common.ps1"

$root = Resolve-Path "$PSScriptRoot\.."
$container = Get-PostgresContainerName
$db = Get-PostgresDatabaseName
$user = Get-PostgresUserName
Assert-PostgresContainerRunning -Container $container

$migrations = Get-ChildItem -Path (Join-Path $root "migrations") -Filter "*.down.sql" | Sort-Object Name -Descending
foreach ($migration in $migrations) {
    $containerMigration = "/tmp/llm_gateway_$($migration.Name)"
    docker cp $migration.FullName "${container}:$containerMigration"
    docker exec -i $container psql -v ON_ERROR_STOP=1 -U $user -d $db -f $containerMigration
}
