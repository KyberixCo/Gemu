param([switch]$ResetData)
$ErrorActionPreference = 'Stop'

foreach ($command in @('go', 'npm.cmd', 'wails')) {
    if (-not (Get-Command $command -ErrorAction SilentlyContinue)) {
        throw "No se encuentra $command en PATH. Instala los requisitos y vuelve a abrir PowerShell."
    }
}

# Development has its own store; production data is never reset here.
$devDataPath = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot 'artifacts/development-data'))
$artifactsPath = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot 'artifacts'))
if (-not $devDataPath.StartsWith($artifactsPath + [IO.Path]::DirectorySeparatorChar)) {
    throw 'La ruta de datos debe permanecer dentro de artifacts.'
}
if ($ResetData) {
    if (Get-Process -Name gemu -ErrorAction SilentlyContinue) {
        throw 'Cierra Gemu antes de reiniciar los datos.'
    }
    if (Test-Path -LiteralPath $devDataPath) {
        $backupPath = Join-Path $artifactsPath ('development-data-backup-' + (Get-Date -Format 'yyyyMMdd-HHmmss-fff'))
        Move-Item -LiteralPath $devDataPath -Destination $backupPath
        Write-Host "Datos anteriores conservados en $backupPath"
    }
}
New-Item -ItemType Directory -Path $devDataPath -Force | Out-Null
$previousDataDir = $env:GEMU_DATA_DIR
$env:GEMU_DATA_DIR = $devDataPath
Push-Location $PSScriptRoot
try {
    & (Join-Path $PSScriptRoot 'build.ps1') -Dev @args
    $exitCode = $LASTEXITCODE
}
finally {
    Pop-Location
    $env:GEMU_DATA_DIR = $previousDataDir
}
exit $exitCode
