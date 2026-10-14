param([switch]$DownloadDependencies)
$ErrorActionPreference = 'Stop'
$overlayRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$outputPath = Join-Path $overlayRoot 'artifacts/overlay'
$dependencyPath = Join-Path $outputPath 'deps'
New-Item -ItemType Directory -Path $dependencyPath -Force | Out-Null

function Get-OverlayDependency([string]$Name, [string]$Url, [string]$Directory, [string]$Sha256 = '') {
    $destination = Join-Path $dependencyPath $Directory
    if (Test-Path -LiteralPath $destination) { return $destination }
    if (-not $DownloadDependencies) { throw "Falta $Name. Ejecuta scripts/build-overlay.ps1 -DownloadDependencies para preparar el compilador portable y las bibliotecas." }
    $archive = Join-Path $dependencyPath ($Name + '.zip')
    Write-Host "Descargando $Name..."
    Invoke-WebRequest -Uri $Url -OutFile $archive -UseBasicParsing
    if ($Sha256 -and (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash -ne $Sha256) { throw "Hash incorrecto para $Name." }
    Expand-Archive -LiteralPath $archive -DestinationPath $dependencyPath -Force
    if (-not (Test-Path -LiteralPath $destination)) { throw "No se encontró el directorio de $Name tras extraerlo." }
    return $destination
}

$zigPath = Get-OverlayDependency 'zig-0.14.1' 'https://ziglang.org/download/0.14.1/zig-x86_64-windows-0.14.1.zip' 'zig-x86_64-windows-0.14.1' '554f5378228923ffd558eac35e21af020c73789d87afeabf4bfd16f2e6feed2c'
$imguiPath = Get-OverlayDependency 'imgui-1.91.9b' 'https://github.com/ocornut/imgui/archive/refs/tags/v1.91.9b.zip' 'imgui-1.91.9b'
$minhookPath = Get-OverlayDependency 'minhook-1.3.4' 'https://github.com/TsudaKageyu/minhook/archive/refs/tags/v1.3.4.zip' 'minhook-1.3.4'
$compiler = Join-Path $zigPath 'zig.exe'
$sourcePath = Join-Path $overlayRoot 'native/overlay'
$previousCache = $env:ZIG_GLOBAL_CACHE_DIR
$env:ZIG_GLOBAL_CACHE_DIR = Join-Path $outputPath 'zig-cache'
try {
    $objects = @()
    foreach ($relative in @('buffer.c','hook.c','trampoline.c','hde/hde64.c')) {
        $source = Join-Path $minhookPath ('src/' + $relative)
        $object = Join-Path $outputPath ([IO.Path]::GetFileNameWithoutExtension($relative) + '.obj')
        & $compiler cc -target x86_64-windows-gnu -O2 -c $source -o $object
        if ($LASTEXITCODE -ne 0) { throw "Error compilando $relative." }
        $objects += $object
    }
    $sources = @((Join-Path $sourcePath 'overlay.cpp'))
    foreach ($relative in @('imgui.cpp','imgui_draw.cpp','imgui_tables.cpp','imgui_widgets.cpp','backends/imgui_impl_dx11.cpp','backends/imgui_impl_win32.cpp')) { $sources += Join-Path $imguiPath $relative }
    $includes = @('-I' + $imguiPath, '-I' + (Join-Path $imguiPath 'backends'), '-I' + (Join-Path $minhookPath 'include'))
    $libraries = @('-ld3d11','-ldxgi','-ldxguid','-ld3dcompiler','-luser32','-lgdi32','-ldwmapi','-limm32')
    Write-Host 'Compilando overlay DirectX 11 x64...'
    & $compiler c++ -target x86_64-windows-gnu -std=c++17 -fms-extensions -O2 -g -DUNICODE -D_UNICODE -shared @sources @objects @includes @libraries -o (Join-Path $outputPath 'gemu-overlay.dll')
    if ($LASTEXITCODE -ne 0) { throw 'Error compilando gemu-overlay.dll.' }
    & $compiler c++ -target x86_64-windows-gnu -std=c++17 -fms-extensions -O2 -DUNICODE -D_UNICODE (Join-Path $sourcePath 'smoke.cpp') @libraries -o (Join-Path $outputPath 'overlay-smoke.exe')
    if ($LASTEXITCODE -ne 0) { throw 'Error compilando la escena de prueba.' }
    Push-Location $overlayRoot
    try {
        & go build -o (Join-Path $outputPath 'overlay-check.exe') ./cmd/overlay-check
        if ($LASTEXITCODE -ne 0) { throw 'Error compilando overlay-check.' }
    } finally { Pop-Location }
    $licenses = Join-Path $outputPath 'licenses'
    New-Item -ItemType Directory -Path $licenses -Force | Out-Null
    Copy-Item -LiteralPath (Join-Path $imguiPath 'LICENSE.txt') -Destination (Join-Path $licenses 'imgui.txt') -Force
    Copy-Item -LiteralPath (Join-Path $minhookPath 'LICENSE.txt') -Destination (Join-Path $licenses 'minhook.txt') -Force
    Copy-Item -LiteralPath (Join-Path $zigPath 'LICENSE') -Destination (Join-Path $licenses 'zig.txt') -Force
    Write-Host "Overlay y escena de prueba listos en $outputPath"
} finally { $env:ZIG_GLOBAL_CACHE_DIR = $previousCache }
