param([switch]$Quick)
$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)

function Step($name, [scriptblock]$body) {
    Write-Host "== $name" -ForegroundColor Cyan
    & $body
    if ($LASTEXITCODE -ne 0) { Write-Host "FAILED: $name" -ForegroundColor Red; exit 1 }
}

Step 'gofmt' {
    $out = gofmt -l .
    if ($out) { Write-Host $out; $global:LASTEXITCODE = 1 } else { $global:LASTEXITCODE = 0 }
}
Step 'go vet' { go vet ./... }

if (-not $Quick) {
    Step 'staticcheck' { go run honnef.co/go/tools/cmd/staticcheck@latest ./... }
    Step 'pure packages (CGO off, coverage)' {
        $env:CGO_ENABLED = '0'
        go test -cover ./internal/snippets/... ./internal/store/... ./internal/instance/...
        $code = $LASTEXITCODE
        $env:CGO_ENABLED = '1'
        $global:LASTEXITCODE = $code
    }
    Step 'race (snippets, store)' { go test -race ./internal/snippets/... ./internal/store/... }
}
Step 'go test ./...' { go test ./... }
Write-Host 'ALL CHECKS PASSED' -ForegroundColor Green
