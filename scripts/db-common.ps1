function Get-PostgresContainerName {
    $container = $env:POSTGRES_CONTAINER
    if ([string]::IsNullOrWhiteSpace($container)) {
        $container = "llm-gateway-postgres"
    }
    return $container
}

function Get-PostgresDatabaseName {
    $db = $env:POSTGRES_DB
    if ([string]::IsNullOrWhiteSpace($db)) {
        $db = "llm_gateway"
    }
    return $db
}

function Get-PostgresUserName {
    $user = $env:POSTGRES_USER
    if ([string]::IsNullOrWhiteSpace($user)) {
        $user = "llm_gateway"
    }
    return $user
}

function Assert-PostgresContainerRunning {
    param(
        [Parameter(Mandatory = $true)]
        [string] $Container
    )

    $running = docker inspect -f "{{.State.Running}}" $Container 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Postgres container '$Container' was not found. Set POSTGRES_CONTAINER or start the project's database container."
    }
    if ($running.Trim() -ne "true") {
        throw "Postgres container '$Container' is not running. Start it with 'docker start $Container' or set POSTGRES_CONTAINER to a running Postgres container."
    }
}

function Invoke-PostgresScalar {
    param(
        [Parameter(Mandatory = $true)]
        [string] $Container,
        [Parameter(Mandatory = $true)]
        [string] $User,
        [Parameter(Mandatory = $true)]
        [string] $Database,
        [Parameter(Mandatory = $true)]
        [string] $Sql
    )

    $value = docker exec -i $Container psql -v ON_ERROR_STOP=1 -U $User -d $Database -t -A -c $Sql
    if ($LASTEXITCODE -ne 0) {
        throw "Postgres command failed in container '$Container'."
    }
    if ($null -eq $value) {
        return ""
    }
    return $value.Trim()
}
