param([switch]$InspectOnly)
$ErrorActionPreference = 'Stop'

# Recover only this checkout's running app. Wails can keep its fullscreen
# flag while Windows has minimised or hidden the actual native window.
$checkout = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$allowedPaths = @('Gemu.exe', 'Gemu-dev.exe') | ForEach-Object {
    [IO.Path]::GetFullPath((Join-Path $checkout ('build/bin/' + $_)))
}
$apps = @(Get-Process -Name Gemu,Gemu-dev -ErrorAction SilentlyContinue | Where-Object {
    $_.Path -and $allowedPaths -contains [IO.Path]::GetFullPath($_.Path)
})
if (-not $apps.Count) { throw 'No running Gemu instance was found for this checkout.' }

Add-Type -TypeDefinition @'
using System;
using System.Collections.Generic;
using System.Runtime.InteropServices;
using System.Text;
public static class GemuWindowRecovery {
    public delegate bool EnumProc(IntPtr hwnd, IntPtr arg);
    [DllImport("user32.dll")] static extern bool EnumWindows(EnumProc callback, IntPtr arg);
    [DllImport("user32.dll")] static extern uint GetWindowThreadProcessId(IntPtr hwnd, out uint pid);
    [DllImport("user32.dll", CharSet=CharSet.Unicode)] static extern int GetClassName(IntPtr hwnd, StringBuilder text, int length);
    [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr hwnd);
    [DllImport("user32.dll")] public static extern bool IsIconic(IntPtr hwnd);
    [DllImport("user32.dll")] static extern bool ShowWindowAsync(IntPtr hwnd, int command);
    [DllImport("user32.dll")] static extern bool SetWindowPos(IntPtr hwnd, IntPtr after, int x, int y, int width, int height, uint flags);
    [DllImport("user32.dll")] static extern bool SetForegroundWindow(IntPtr hwnd);
    public static IntPtr[] Find(uint target) {
        var found = new List<IntPtr>();
        EnumWindows((hwnd, arg) => {
            uint pid;
            GetWindowThreadProcessId(hwnd, out pid);
            if (pid == target) {
                var name = new StringBuilder(256);
                GetClassName(hwnd, name, 256);
                if (name.ToString() == "wailsWindow") found.Add(hwnd);
            }
            return true;
        }, IntPtr.Zero);
        return found.ToArray();
    }
    public static bool Restore(IntPtr hwnd) {
        ShowWindowAsync(hwnd, IsIconic(hwnd) ? 9 : 5);
        // HWND_TOP raises the existing window without making it always-on-top.
        SetWindowPos(hwnd, IntPtr.Zero, 0, 0, 0, 0, 0x0001 | 0x0002 | 0x0040 | 0x4000);
        return SetForegroundWindow(hwnd);
    }
}
'@

$found = $false
foreach ($app in $apps) {
    foreach ($handle in [GemuWindowRecovery]::Find($app.Id)) {
        $found = $true
        $result = [ordered]@{
            ProcessId = $app.Id
            Handle = $handle.ToInt64()
            Visible = [GemuWindowRecovery]::IsWindowVisible($handle)
            Minimized = [GemuWindowRecovery]::IsIconic($handle)
        }
        if (-not $InspectOnly) { $result['ForegroundRequested'] = [GemuWindowRecovery]::Restore($handle) }
        [pscustomobject]$result
    }
}
if (-not $found) { throw 'The Gemu window is not accessible from this desktop. Run this script from your desktop PowerShell.' }
