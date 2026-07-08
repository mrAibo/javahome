package winenv

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type Scope string

const (
	User    Scope = "User"
	Machine Scope = "Machine"
)

func ParseScope(value string) (Scope, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "user", "u":
		return User, nil
	case "machine", "system", "systemwide", "system-wide", "m":
		return Machine, nil
	default:
		return "", fmt.Errorf("unsupported Windows environment scope %q; use user or machine", value)
	}
}

func Apply(scope Scope, javaHome string, dryRun bool) (string, error) {
	if runtime.GOOS != "windows" {
		return "", errors.New("windows-env is only available on Windows")
	}
	javaHome = strings.TrimSpace(javaHome)
	if javaHome == "" {
		return "", errors.New("java home is empty")
	}
	if scope != User && scope != Machine {
		return "", fmt.Errorf("unsupported Windows environment scope %q", scope)
	}
	if dryRun {
		return fmt.Sprintf("Would set JAVA_HOME to %q and prepend %q to %s Path", javaHome, javaHome+`\bin`, scope), nil
	}

	exe, err := powerShellExecutable()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(exe,
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy", "Bypass",
		"-Command", windowsEnvScript(),
		"-JavaHome", javaHome,
		"-Scope", string(scope),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("failed to update Windows environment: %w", err)
	}
	return string(out), nil
}

func powerShellExecutable() (string, error) {
	for _, name := range []string{"powershell.exe", "pwsh.exe"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", errors.New("PowerShell executable not found")
}

func windowsEnvScript() string {
	return `param(
  [Parameter(Mandatory=$true)][string]$JavaHome,
  [Parameter(Mandatory=$true)][ValidateSet('User','Machine')][string]$Scope
)
$ErrorActionPreference = 'Stop'

$javaBin = Join-Path $JavaHome 'bin'
$currentPath = [Environment]::GetEnvironmentVariable('Path', $Scope)
if ($null -eq $currentPath) { $currentPath = '' }

$seen = @{}
$parts = New-Object 'System.Collections.Generic.List[string]'

function Add-PathPart([string]$Value) {
  if ([string]::IsNullOrWhiteSpace($Value)) { return }
  $trimmed = $Value.Trim()
  if ($trimmed -eq $javaBin) { return }
  if ($trimmed -match '(?i)\\(jre|jdk|java)[^\\;]*\\bin$') { return }
  $key = $trimmed.ToLowerInvariant()
  if (-not $seen.ContainsKey($key)) {
    $seen[$key] = $true
    [void]$parts.Add($trimmed)
  }
}

$seen[$javaBin.ToLowerInvariant()] = $true
[void]$parts.Add($javaBin)
foreach ($part in ($currentPath -split ';')) { Add-PathPart $part }
$newPath = ($parts -join ';')

[Environment]::SetEnvironmentVariable('JAVA_HOME', $JavaHome, $Scope)
[Environment]::SetEnvironmentVariable('Path', $newPath, $Scope)

try {
  $signature = '[DllImport("user32.dll", SetLastError=true, CharSet=CharSet.Auto)] public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);'
  $type = Add-Type -MemberDefinition $signature -Name NativeMethods -Namespace Win32 -PassThru
  $result = [UIntPtr]::Zero
  [void]$type::SendMessageTimeout([IntPtr]0xffff, 0x1A, [UIntPtr]::Zero, 'Environment', 0x0002, 5000, [ref]$result)
} catch {
  Write-Warning "Environment update was saved, but broadcast notification failed: $($_.Exception.Message)"
}

Write-Output "Updated $Scope environment: JAVA_HOME=$JavaHome"
Write-Output "New processes will see the updated environment. Restart terminals, IDEs, services, or daemons that were already running."
`
}
