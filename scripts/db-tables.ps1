Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$container = $env:POSTGRES_CONTAINER
if ([string]::IsNullOrWhiteSpace($container)) {
    $container = "llm-gateway-postgres"
}
$db = $env:POSTGRES_DB
if ([string]::IsNullOrWhiteSpace($db)) {
    $db = "llm_gateway"
}
$user = $env:POSTGRES_USER
if ([string]::IsNullOrWhiteSpace($user)) {
    $user = "llm_gateway"
}

docker exec -i $container psql -U $user -d $db -c "\dt"
