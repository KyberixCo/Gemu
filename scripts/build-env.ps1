function Get-ScreenScraperBuildConfiguration {
    param([Parameter(Mandatory)][string]$EnvPath)

    $names = @('GEMU_SCREENSCRAPER_DEVID', 'GEMU_SCREENSCRAPER_DEVPASSWORD', 'GEMU_SCREENSCRAPER_SOFTNAME')
    $values = @{}
    if (Test-Path -LiteralPath $EnvPath) {
        $lineNumber = 0
        foreach ($line in [IO.File]::ReadAllLines($EnvPath, [Text.Encoding]::UTF8)) {
            $lineNumber++
            $entry = $line.Trim()
            if (-not $entry -or $entry.StartsWith('#')) { continue }
            if ($entry -notmatch '^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$') {
                throw "Formato .env invalido en la linea $lineNumber."
            }
            $name, $value = $Matches[1], $Matches[2]
            if ($names -notcontains $name) { continue }
            if ($values.ContainsKey($name)) {
                throw "Variable duplicada en .env, linea $lineNumber."
            }
            if ($value.StartsWith('"') -or $value.StartsWith("'")) {
                $quote = $value.Substring(0, 1)
                $closing = $value.IndexOf($quote, 1)
                if ($closing -lt 0 -or $value.Substring($closing + 1).Trim() -notmatch '^(#.*)?$') {
                    throw "Valor entre comillas invalido en .env, linea $lineNumber."
                }
                $value = $value.Substring(1, $closing - 1)
            } else {
                $value = ($value -replace '\s+#.*$', '').TrimEnd()
            }
            # Literal values only: never expand variables or execute .env content.
            $values[$name] = $value
        }
    }
    foreach ($name in $names) {
        $processValue = [Environment]::GetEnvironmentVariable($name, 'Process')
        if ($null -ne $processValue) { $values[$name] = $processValue }
    }
    if (-not $values['GEMU_SCREENSCRAPER_SOFTNAME']) { $values['GEMU_SCREENSCRAPER_SOFTNAME'] = 'Gemu' }
    if ([string]::IsNullOrWhiteSpace($values['GEMU_SCREENSCRAPER_DEVID']) -or [string]::IsNullOrWhiteSpace($values['GEMU_SCREENSCRAPER_DEVPASSWORD'])) {
        throw 'Configura GEMU_SCREENSCRAPER_DEVID y GEMU_SCREENSCRAPER_DEVPASSWORD en .env o en las variables de entorno antes de compilar. Usa .env.example como plantilla.'
    }
    return $values
}
