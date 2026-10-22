<#
.SYNOPSIS
    Build ScarScript compiler for Windows, Linux, and macOS
.DESCRIPTION
    Cross-compiles the ScarScript CLI for all supported platforms.
    Outputs binaries to ./dist/ and project root
#>

param(
    [string]$Version = "1.0.1",
    [switch]$Clean
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$distDir = Join-Path $repoRoot "dist"
$cmdPath = Join-Path $repoRoot "cmd\scar"

if ($Clean -and (Test-Path $distDir)) {
    Remove-Item $distDir -Recurse -Force
}
New-Item -ItemType Directory -Path $distDir -Force | Out-Null

# Also output to project root for convenience
$rootDistDir = $repoRoot

$ldflags = "-s -w -X main.version=$Version"

$targets = @(
    @{ GOOS="windows"; GOARCH="amd64"; Ext=".exe";   Name="scar-windows-amd64" },
    @{ GOOS="linux";   GOARCH="amd64"; Ext="";      Name="scar-linux-amd64" },
    @{ GOOS="darwin";  GOARCH="amd64"; Ext="";      Name="scar-darwin-amd64" },
    @{ GOOS="darwin";  GOARCH="arm64"; Ext="";      Name="scar-darwin-arm64" }
)

Write-Host "Building ScarScript v$Version..." -ForegroundColor Cyan

foreach ($t in $targets) {
    $outName = "$($t.Name)$($t.Ext)"
    $outPath = Join-Path $distDir $outName
    $rootOutPath = Join-Path $rootDistDir $outName
    
    Write-Host "  Building $outName..." -NoNewline
    
    $env:GOOS = $t.GOOS
    $env:GOARCH = $t.GOARCH
    $env:CGO_ENABLED = "0"
    
    $result = go build -ldflags $ldflags -o $outPath $cmdPath 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host " FAILED" -ForegroundColor Red
        Write-Error $result
        exit 1
    }
    
    # Copy to project root
    Copy-Item $outPath $rootOutPath -Force
    
    Write-Host " OK" -ForegroundColor Green
}

# Generate checksums
Write-Host "`nGenerating checksums..." -ForegroundColor Cyan
$checksumPath = Join-Path $distDir "checksums.txt"
$files = Get-ChildItem $distDir -File | Where-Object { $_.Name -ne "checksums.txt" }
$hashes = @()
foreach ($f in $files) {
    $hash = (Get-FileHash -Algorithm SHA256 -Path $f.FullName).Hash
    $hashes += "$hash  $($f.Name)"
}
[IO.File]::WriteAllLines($checksumPath, $hashes)
Write-Host "  checksums.txt created" -ForegroundColor Green

Write-Host "`nDone! Binaries in $distDir and project root" -ForegroundColor Cyan
Get-ChildItem $distDir | Format-Table Name, Length