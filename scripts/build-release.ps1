$ErrorActionPreference = "Stop"

$RootDir = Resolve-Path (Join-Path $PSScriptRoot "..")
$OutDir = Join-Path $RootDir "dist"
$Binary = "greentrace"

New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

$Targets = @(
    @{ Os = "windows"; Arch = "amd64" },
    @{ Os = "windows"; Arch = "arm64" },
    @{ Os = "linux"; Arch = "amd64" },
    @{ Os = "linux"; Arch = "arm64" },
    @{ Os = "darwin"; Arch = "amd64" },
    @{ Os = "darwin"; Arch = "arm64" }
)

foreach ($target in $Targets) {
    $os = $target.Os
    $arch = $target.Arch
    $ext = ""
    if ($os -eq "windows") { $ext = ".exe" }

    $outName = "${Binary}_${os}_${arch}${ext}"
    Write-Host "Building $outName"

    $env:GOOS = $os
    $env:GOARCH = $arch
    go build -o (Join-Path $OutDir $outName) .
}

Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host "Artifacts in $OutDir"