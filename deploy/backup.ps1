param(
    [string]$EnvFile = ".env.prod",
    [string]$ProjectName = "contest-game",
    [string]$OutputDir = ".data/backups"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Push-Location $root

function Assert-NativeSuccess([string]$Operation) {
    if ($LASTEXITCODE -ne 0) {
        throw "$Operation failed with exit code $LASTEXITCODE"
    }
}

try {
    if (-not (Test-Path -LiteralPath $EnvFile)) {
        throw "Environment file not found: $EnvFile"
    }

    $values = @{} <# 仅读取连接变量，密码由容器环境传给 pg_dump。 #>
    foreach ($line in Get-Content -LiteralPath $EnvFile -Encoding utf8) {
        $entry = [regex]::Match($line, '^\s*([^#=\s]+)\s*=\s*(.*)\s*$')
        if ($entry.Success) {
            $values[$entry.Groups[1].Value] = $entry.Groups[2].Value.Trim().Trim("'").Trim('"')
        }
    }
    foreach ($name in @("POSTGRES_USER", "POSTGRES_DB")) {
        if (-not $values.ContainsKey($name) -or [string]::IsNullOrWhiteSpace($values[$name])) {
            throw "$name is required in $EnvFile"
        }
    }

    $outputPath = [IO.Path]::GetFullPath((Join-Path $root $OutputDir))
    New-Item -ItemType Directory -Force -Path $outputPath | Out-Null
    $stamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $composeArgs = @("--env-file", $EnvFile, "-p", $ProjectName, "-f", "docker-compose.prod.yml")

    $postgres = (@(& docker compose @composeArgs ps -q postgres) -join "").Trim()
    Assert-NativeSuccess "locate postgres container"
    $redis = (@(& docker compose @composeArgs ps -q redis) -join "").Trim()
    Assert-NativeSuccess "locate redis container"
    $minio = (@(& docker compose @composeArgs ps -q minio) -join "").Trim()
    Assert-NativeSuccess "locate minio container"
    if ([string]::IsNullOrWhiteSpace($postgres) -or [string]::IsNullOrWhiteSpace($redis) -or [string]::IsNullOrWhiteSpace($minio)) {
        throw "postgres, redis and minio containers must be running before backup"
    }

    $pgFile = "/tmp/contest-game-$stamp.dump"
    & docker exec $postgres pg_dump -U $values["POSTGRES_USER"] -d $values["POSTGRES_DB"] -Fc -f $pgFile
    Assert-NativeSuccess "postgres dump"
    & docker cp "${postgres}:$pgFile" (Join-Path $outputPath "postgres-$stamp.dump")
    Assert-NativeSuccess "copy postgres dump"
    & docker exec $postgres rm -f $pgFile
    Assert-NativeSuccess "remove temporary postgres dump"

    $redisFile = "/tmp/contest-game-$stamp.rdb"
    & docker exec $redis redis-cli --rdb $redisFile
    Assert-NativeSuccess "redis dump"
    & docker cp "${redis}:$redisFile" (Join-Path $outputPath "redis-$stamp.rdb")
    Assert-NativeSuccess "copy redis dump"
    & docker exec $redis rm -f $redisFile
    Assert-NativeSuccess "remove temporary redis dump"

    $minioArchive = "minio-$stamp.tar.gz" <# 用只读卷挂载的 Alpine sidecar 打包 MinIO 数据。 #>
    & docker run --rm --volumes-from $minio -v "${outputPath}:/backup" alpine:3.18 sh -c "tar -czf /backup/$minioArchive -C /data ."
    Assert-NativeSuccess "minio volume archive"

    Write-Output "Backup completed: $outputPath"
}
finally {
    Pop-Location
}
