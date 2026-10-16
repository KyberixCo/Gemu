$ErrorActionPreference = 'Stop'
$assistRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$assistOutput = Join-Path $assistRoot 'build/bin/gemu-input.exe'
New-Item -ItemType Directory -Path (Split-Path -Parent $assistOutput) -Force | Out-Null
Push-Location $assistRoot
try {
    & go build -ldflags '-H=windowsgui' -o $assistOutput ./cmd/input-assist
    if ($LASTEXITCODE -ne 0) { throw 'No se pudo compilar el asistente de escritorio.' }
}
finally { Pop-Location }
