$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)
if (-not (Test-Path cmd/sniper/Icon.png)) {
    go run ./tools/genicon
    if ($LASTEXITCODE) { exit 1 }
}
if (-not (Get-Command fyne -ErrorAction SilentlyContinue)) { go install fyne.io/tools/cmd/fyne@latest }
Push-Location cmd/sniper
try {
    fyne package -os windows -release
    if ($LASTEXITCODE) { exit 1 }
    New-Item -ItemType Directory -Force ../../dist | Out-Null
    Move-Item -Force Sniper.exe ../../dist/Sniper.exe
} finally {
    Pop-Location
}
$f = Get-Item dist/Sniper.exe
'dist/Sniper.exe: {0:N1} MiB ({1} bytes)' -f ($f.Length / 1MB), $f.Length
