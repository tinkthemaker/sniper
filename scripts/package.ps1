param([switch]$Force)
$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)

$exe = 'dist/Sniper.exe'

# Rebuild only when something that goes into the exe is newer than the exe.
function Test-Stale {
    if ($Force -or -not (Test-Path $exe)) { return $true }
    $built = (Get-Item $exe).LastWriteTime
    $inputs = @(Get-ChildItem cmd, internal, tools -Recurse -File |
        Where-Object { $_.Name -notlike '*_test.go' -and $_.FullName -notmatch 'testdata' -and $_.Extension -in '.go', '.png', '.toml' }) +
        @(Get-Item go.mod, go.sum)
    $newer = $inputs | Where-Object { $_.LastWriteTime -gt $built } | Select-Object -First 1
    if ($newer) { Write-Host "stale: $($newer.FullName) is newer than $exe" }
    return [bool]$newer
}

if (-not (Test-Stale)) {
    $f = Get-Item $exe
    'dist/Sniper.exe is up to date ({0:N1} MiB, built {1:HH:mm:ss})' -f ($f.Length / 1MB), $f.LastWriteTime
    exit 0
}

if (-not (Test-Path cmd/sniper/Icon.png)) {
    go run ./tools/genicon
    if ($LASTEXITCODE) { exit 1 }
}
if (-not (Get-Command fyne -ErrorAction SilentlyContinue)) { go install fyne.io/tools/cmd/fyne@latest }

# fyne package raises Build in FyneApp.toml on every run; put the file back so
# a rebuild never dirties the working tree.
$toml = 'cmd/sniper/FyneApp.toml'
$tomlBytes = [IO.File]::ReadAllBytes((Resolve-Path $toml))
Push-Location cmd/sniper
try {
    fyne package -os windows -release
    if ($LASTEXITCODE) { exit 1 }
    New-Item -ItemType Directory -Force ../../dist | Out-Null
    Move-Item -Force Sniper.exe ../../dist/Sniper.exe
} finally {
    Pop-Location
    [IO.File]::WriteAllBytes((Resolve-Path $toml), $tomlBytes)
}
# The restore above made the toml newer than the exe; stamp the exe after it.
(Get-Item $exe).LastWriteTime = Get-Date
$f = Get-Item $exe
'dist/Sniper.exe: {0:N1} MiB ({1} bytes), built {2:HH:mm:ss}' -f ($f.Length / 1MB), $f.Length, $f.LastWriteTime
