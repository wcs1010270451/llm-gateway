Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. "$PSScriptRoot\db-common.ps1"

$root = Resolve-Path "$PSScriptRoot\.."
$container = Get-PostgresContainerName
$db = Get-PostgresDatabaseName
$user = Get-PostgresUserName
Assert-PostgresContainerRunning -Container $container

$migrations = Get-ChildItem -Path (Join-Path $root "migrations") -Filter "*.up.sql" | Sort-Object Name
docker exec -i $container psql -v ON_ERROR_STOP=1 -U $user -d $db -c "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());"

$existingUsers = Invoke-PostgresScalar -Container $container -User $user -Database $db -Sql "SELECT to_regclass('public.users') IS NOT NULL;"
$has0001 = Invoke-PostgresScalar -Container $container -User $user -Database $db -Sql "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '0001_init');"
if ($existingUsers -eq "t" -and $has0001 -ne "t") {
    docker exec -i $container psql -v ON_ERROR_STOP=1 -U $user -d $db -c "INSERT INTO schema_migrations (version) VALUES ('0001_init') ON CONFLICT DO NOTHING;"
}

foreach ($migration in $migrations) {
    $version = $migration.Name -replace '\.up\.sql$', ''
    $applied = Invoke-PostgresScalar -Container $container -User $user -Database $db -Sql "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '$version');"
    if ($applied -eq "t") {
        Write-Host "Skipping applied migration $version"
        continue
    }

    $containerMigration = "/tmp/llm_gateway_$($migration.Name)"
    docker cp $migration.FullName "${container}:$containerMigration"
    docker exec -i $container psql -v ON_ERROR_STOP=1 -U $user -d $db -f $containerMigration
    docker exec -i $container psql -v ON_ERROR_STOP=1 -U $user -d $db -c "INSERT INTO schema_migrations (version) VALUES ('$version');"
}
