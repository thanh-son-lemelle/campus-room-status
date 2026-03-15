param(
    [string]$BaseUrl = "http://localhost:8080/api/v1",
    [string]$Duration = "60s",
    [string]$SleepSeconds = "0.2",
    [string]$OutDir = "loadtests/results",
    [string]$RoomCode = ""
)

$ErrorActionPreference = "Stop"

$tiers = @(10, 50, 100, 1000)
New-Item -ItemType Directory -Path $OutDir -Force | Out-Null

foreach ($vus in $tiers) {
    $stamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $baseName = "k6-vus-$vus-$stamp"
    $summaryPath = Join-Path $OutDir "$baseName-summary.json"
    $logPath = Join-Path $OutDir "$baseName.log"

    Write-Host "=== k6 run: VUS=$vus, duration=$Duration ==="

    $args = @(
        "run",
        "loadtests/k6_rooms.js",
        "--env", "BASE_URL=$BaseUrl",
        "--env", "VUS=$vus",
        "--env", "DURATION=$Duration",
        "--env", "SLEEP_SECONDS=$SleepSeconds",
        "--summary-export", $summaryPath
    )

    if (-not [string]::IsNullOrWhiteSpace($RoomCode)) {
        $args += @("--env", "ROOM_CODE=$RoomCode")
    }

    & k6 @args 2>&1 | Tee-Object -FilePath $logPath

    if ($LASTEXITCODE -ne 0) {
        throw "k6 failed for VUS=$vus. Check $logPath"
    }
}

Write-Host "All runs completed. Results are in $OutDir"
