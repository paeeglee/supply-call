# Builds bin\SupplyCall.exe (single file, no console, no CGO).
# Usage: ./build.ps1 1.0.0
param([string]$Version = "dev")
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

function Step($name, [scriptblock]$cmd) {
    Write-Host "==> $name"
    & $cmd
    if ($LASTEXITCODE -ne 0) { throw "$name failed (exit $LASTEXITCODE)" }
}

Step "go generate (icon, exe resource)" { go generate ./assets }
Step "go vet" { go vet ./... }

if (Get-Command gcc -ErrorAction SilentlyContinue) {
    $env:CGO_ENABLED = "1"
    Step "go test -race" { go test -race -count=1 ./... }
} else {
    Write-Host "(gcc not found: running tests without -race)"
    Step "go test" { go test -count=1 ./... }
}

$env:CGO_ENABLED = "0"; $env:GOOS = "windows"; $env:GOARCH = "amd64"
try {
    Step "go build $Version" { go build -trimpath -ldflags "-H windowsgui -s -w -X main.version=$Version" -o bin\SupplyCall.exe . }
} finally {
    Remove-Item Env:CGO_ENABLED, Env:GOOS, Env:GOARCH -ErrorAction SilentlyContinue
}

$exe = Get-Item bin\SupplyCall.exe
Write-Host ("OK: {0} ({1:N1} MB)" -f $exe.FullName, ($exe.Length / 1MB))
