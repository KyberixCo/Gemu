$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'build-env.ps1')

$fixturePath = [IO.Path]::GetTempFileName()
$buildFixture = Join-Path ([IO.Path]::GetTempPath()) ('gemu-build-env-' + [Guid]::NewGuid().ToString('N'))
$names = @('GEMU_SCREENSCRAPER_DEVID', 'GEMU_SCREENSCRAPER_DEVPASSWORD', 'GEMU_SCREENSCRAPER_SOFTNAME')
$previousValues = @{}
foreach ($name in $names) {
    $previousValues[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
    [Environment]::SetEnvironmentVariable($name, $null, 'Process')
}
function Assert-Condition([bool]$Condition, [string]$Message) {
    if (-not $Condition) { throw $Message }
}
function Write-Fixture([string]$Content) {
    [IO.File]::WriteAllText($fixturePath, $Content, [Text.UTF8Encoding]::new($false))
}
function Assert-Rejected([string]$Content) {
    Write-Fixture $Content
    $rejected = $false
    try { $null = Get-ScreenScraperBuildConfiguration -EnvPath $fixturePath }
    catch {
        $rejected = $true
        Assert-Condition (-not $_.Exception.Message.Contains('fixture-secret')) 'Error disclosed a fixture value.'
    }
    Assert-Condition $rejected 'Invalid configuration was accepted.'
}
try {
    Write-Fixture @'
# Literal values, quoted whitespace, comments and export prefixes.
export GEMU_SCREENSCRAPER_DEVID = fixture-id # comment
GEMU_SCREENSCRAPER_DEVPASSWORD = 'fixture-secret#$()=literal' # comment
GEMU_SCREENSCRAPER_SOFTNAME = "Gemu test"
UNRELATED_VARIABLE=ignored
'@
    $values = Get-ScreenScraperBuildConfiguration -EnvPath $fixturePath
    Assert-Condition ($values[$names[0]] -eq 'fixture-id') 'ID or comment parsing failed.'
    Assert-Condition ($values[$names[1]] -ceq 'fixture-secret#$()=literal') 'Password was expanded or changed.'
    Assert-Condition ($values[$names[2]] -eq 'Gemu test') 'Quoted whitespace parsing failed.'
    Assert-Condition ($values.Count -eq 3) 'Unrelated .env variables were imported.'

    $env:GEMU_SCREENSCRAPER_DEVID = 'ci-id'
    $values = Get-ScreenScraperBuildConfiguration -EnvPath $fixturePath
    Assert-Condition ($values[$names[0]] -eq 'ci-id') 'Process environment did not override .env.'
    [Environment]::SetEnvironmentVariable($names[0], $null, 'Process')

    Write-Fixture "GEMU_SCREENSCRAPER_DEVID=fixture-id`nGEMU_SCREENSCRAPER_DEVPASSWORD=fixture-secret#literal"
    $values = Get-ScreenScraperBuildConfiguration -EnvPath $fixturePath
    Assert-Condition ($values[$names[2]] -eq 'Gemu') 'Default software name failed.'
    Assert-Condition ($values[$names[1]] -eq 'fixture-secret#literal') 'Unquoted hash was changed.'

    Assert-Rejected ''
    Assert-Rejected 'GEMU_SCREENSCRAPER_DEVID=fixture-id'
    Assert-Rejected "GEMU_SCREENSCRAPER_DEVID=fixture-id`nGEMU_SCREENSCRAPER_DEVPASSWORD=   "
    Assert-Rejected "GEMU_SCREENSCRAPER_DEVID=fixture-id`nGEMU_SCREENSCRAPER_DEVID=duplicate`nGEMU_SCREENSCRAPER_DEVPASSWORD=fixture-secret"
    Assert-Rejected 'GEMU_SCREENSCRAPER_DEVPASSWORD="fixture-secret'
    Assert-Rejected 'fixture-secret invalid line'

    $env:GEMU_SCREENSCRAPER_DEVID = 'ci-id'
    $env:GEMU_SCREENSCRAPER_DEVPASSWORD = 'ci-secret'
    $values = Get-ScreenScraperBuildConfiguration -EnvPath ($fixturePath + '.missing')
    Assert-Condition ($values[$names[1]] -eq 'ci-secret') 'CI-only configuration failed.'

    foreach ($name in $names) { [Environment]::SetEnvironmentVariable($name, $null, 'Process') }
    New-Item -ItemType Directory -Path (Join-Path $buildFixture 'scripts'), (Join-Path $buildFixture 'internal/providers/content') -Force | Out-Null
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot '../build.ps1') -Destination (Join-Path $buildFixture 'build.ps1')
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'build-env.ps1') -Destination (Join-Path $buildFixture 'scripts/build-env.ps1')
    $runner = @'
param([switch]$FailBuild, [switch]$Dev)
function wails {
    $generated = [IO.File]::ReadAllText((Join-Path $PSScriptRoot 'internal/providers/content/screenscraper_build_local.go'))
    foreach ($assignment in @('screenScraperDeveloperID = "fixture-id"', 'screenScraperDeveloperPassword = "fixture-secret"', 'screenScraperSoftwareName = "Gemu"')) {
        if (-not $generated.Contains($assignment)) { throw 'Build did not inject the expected .env configuration.' }
    }
    $expectedMode = if ($Dev) { 'dev' } else { 'build' }
    if ($args[0] -ne $expectedMode) { throw 'Unexpected Wails invocation.' }
    $global:LASTEXITCODE = if ($FailBuild) { 7 } else { 0 }
}
if ($Dev) { & (Join-Path $PSScriptRoot 'build.ps1') -Dev } else { & (Join-Path $PSScriptRoot 'build.ps1') }
exit $LASTEXITCODE
'@
    [IO.File]::WriteAllText((Join-Path $buildFixture 'runner.ps1'), $runner)
    [IO.File]::WriteAllText((Join-Path $buildFixture '.env'), "GEMU_SCREENSCRAPER_DEVID=fixture-id`nGEMU_SCREENSCRAPER_DEVPASSWORD=fixture-secret")
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File (Join-Path $buildFixture 'runner.ps1')
    Assert-Condition ($LASTEXITCODE -eq 0) 'Default build did not inject .env without a flag.'
    Assert-Condition (-not (Test-Path -LiteralPath (Join-Path $buildFixture 'internal/providers/content/screenscraper_build_local.go'))) 'Successful build left temporary secrets.'
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File (Join-Path $buildFixture 'runner.ps1') -FailBuild
    Assert-Condition ($LASTEXITCODE -eq 7) 'Wails failure code was not propagated.'
    Assert-Condition (-not (Test-Path -LiteralPath (Join-Path $buildFixture 'internal/providers/content/screenscraper_build_local.go'))) 'Failed build left temporary secrets.'

    $generatedPath = Join-Path $buildFixture 'internal/providers/content/screenscraper_build_local.go'
    $staleSource = "// Code generated by build.ps1. DO NOT COMMIT.`npackage content`n// stale fixture"
    [IO.File]::WriteAllText($generatedPath, $staleSource)
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File (Join-Path $buildFixture 'runner.ps1')
    Assert-Condition ($LASTEXITCODE -eq 0) 'Interrupted build configuration was not recovered.'
    Assert-Condition (-not (Test-Path -LiteralPath $generatedPath)) 'Recovered build did not clean configuration.'
    [IO.File]::WriteAllText($generatedPath, $staleSource)
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File (Join-Path $buildFixture 'runner.ps1') -Dev
    Assert-Condition ($LASTEXITCODE -eq 0) 'Development did not recover stale configuration.'
    Assert-Condition (-not (Test-Path -LiteralPath $generatedPath)) 'Development left temporary configuration.'

    $hasher = [Security.Cryptography.SHA256]::Create()
    try { $fixtureHash = [BitConverter]::ToString($hasher.ComputeHash([Text.Encoding]::UTF8.GetBytes([IO.Path]::GetFullPath($buildFixture).TrimEnd('\').ToUpperInvariant()))).Replace('-', '') }
    finally { $hasher.Dispose() }
    $heldMutex = [Threading.Mutex]::new($false, ('Local\GemuBuild-' + $fixtureHash))
    $held = $heldMutex.WaitOne(0)
    try {
        [IO.File]::WriteAllText($generatedPath, $staleSource)
        $ErrorActionPreference = 'Continue'
        $blockedOutput = & powershell.exe -NoProfile -ExecutionPolicy Bypass -File (Join-Path $buildFixture 'runner.ps1') 2>&1
        $ErrorActionPreference = 'Stop'
        Assert-Condition ($LASTEXITCODE -ne 0) 'Concurrent build was accepted.'
        Assert-Condition ([IO.File]::ReadAllText($generatedPath) -ceq $staleSource) 'Concurrent build changed the active configuration.'
    }
    finally { if ($held) { $heldMutex.ReleaseMutex() }; $heldMutex.Dispose() }
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File (Join-Path $buildFixture 'runner.ps1')
    Assert-Condition ($LASTEXITCODE -eq 0) 'Build remained locked after its owner released the mutex.'

    [IO.File]::WriteAllText($generatedPath, 'custom source to preserve')
    $ErrorActionPreference = 'Continue'
    $customOutput = & powershell.exe -NoProfile -ExecutionPolicy Bypass -File (Join-Path $buildFixture 'runner.ps1') 2>&1
    $ErrorActionPreference = 'Stop'
    Assert-Condition ($LASTEXITCODE -ne 0) 'Unrecognized source was overwritten.'
    Assert-Condition ([IO.File]::ReadAllText($generatedPath) -ceq 'custom source to preserve') 'Unrecognized source was removed.'
    Remove-Item -LiteralPath $generatedPath
    Remove-Item -LiteralPath (Join-Path $buildFixture '.env')
    $ErrorActionPreference = 'Continue'
    $missingOutput = & powershell.exe -NoProfile -ExecutionPolicy Bypass -File (Join-Path $buildFixture 'runner.ps1') 2>&1
    $ErrorActionPreference = 'Stop'
    Assert-Condition ($LASTEXITCODE -ne 0) 'Default build accepted missing credentials.'
    Assert-Condition (-not (Test-Path -LiteralPath (Join-Path $buildFixture 'internal/providers/content/screenscraper_build_local.go'))) 'Validation failure created temporary secrets.'
    Write-Output 'Build configuration checks passed.'
}
finally {
    Remove-Item -LiteralPath $fixturePath -Force
    # Remove only explicitly named fixture files and then empty directories.
    foreach ($relative in @('internal/providers/content/screenscraper_build_local.go', 'scripts/build-env.ps1', 'build.ps1', 'runner.ps1', '.env', 'internal/providers/content', 'internal/providers', 'internal', 'scripts')) {
        $fixtureEntry = Join-Path $buildFixture $relative
        if (Test-Path -LiteralPath $fixtureEntry) { Remove-Item -LiteralPath $fixtureEntry -Force }
    }
    if (Test-Path -LiteralPath $buildFixture) { Remove-Item -LiteralPath $buildFixture -Force }
    foreach ($name in $names) {
        [Environment]::SetEnvironmentVariable($name, $previousValues[$name], 'Process')
    }
}
