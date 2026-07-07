package shellenv

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Shell string

const (
	Bash       Shell = "bash"
	Zsh        Shell = "zsh"
	Fish       Shell = "fish"
	PowerShell Shell = "powershell"
	Cmd        Shell = "cmd"
)

func DetectShell() Shell {
	if runtime.GOOS == "windows" {
		return PowerShell
	}
	sh := filepath.Base(os.Getenv("SHELL"))
	switch sh {
	case "zsh":
		return Zsh
	case "fish":
		return Fish
	default:
		return Bash
	}
}

func ParseShell(value string) (Shell, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "auto":
		return DetectShell(), nil
	case "bash":
		return Bash, nil
	case "zsh":
		return Zsh, nil
	case "fish":
		return Fish, nil
	case "powershell", "pwsh", "ps":
		return PowerShell, nil
	case "cmd":
		return Cmd, nil
	default:
		return "", fmt.Errorf("unsupported shell %q", value)
	}
}

func ActivationScript(shell Shell, javaHome string, pathValue string) string {
	switch shell {
	case Fish:
		return fmt.Sprintf("set -gx JAVA_HOME %s\nset -gx PATH %s\n", fishQuote(javaHome), strings.Join(fishPathList(pathValue), " "))
	case PowerShell:
		return fmt.Sprintf("$env:JAVA_HOME = %s\n$env:Path = %s\n", psQuote(javaHome), psQuote(pathValue))
	case Cmd:
		return fmt.Sprintf("set JAVA_HOME=%s\nset PATH=%s\n", javaHome, pathValue)
	default:
		return fmt.Sprintf("export JAVA_HOME=%s\nexport PATH=%s\n", shQuote(javaHome), shQuote(pathValue))
	}
}

func InitScript(shell Shell) string {
	switch shell {
	case Fish:
		return "# >>> javahome >>>\nfunction jhome\n    javahome use $argv --shell fish | source\nend\n# <<< javahome <<<\n"
	case PowerShell:
		return "# >>> javahome >>>\nfunction jhome { javahome use @args --shell powershell | Invoke-Expression }\n# <<< javahome <<<\n"
	case Cmd:
		return ":: cmd.exe cannot safely eval command output. Use: javahome use 17 --global --shell cmd\n"
	case Zsh:
		return "# >>> javahome >>>\njhome() { eval \"$(javahome use \"$@\" --shell zsh)\"; }\n# <<< javahome <<<\n"
	default:
		return "# >>> javahome >>>\njhome() { eval \"$(javahome use \"$@\" --shell bash)\"; }\n# <<< javahome <<<\n"
	}
}

func ProfilePath(shell Shell) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch shell {
	case Bash:
		return filepath.Join(home, ".bashrc"), nil
	case Zsh:
		return filepath.Join(home, ".zshrc"), nil
	case Fish:
		return filepath.Join(home, ".config", "fish", "config.fish"), nil
	case PowerShell:
		if runtime.GOOS == "windows" {
			docs := os.Getenv("USERPROFILE")
			if docs == "" {
				docs = home
			}
			return filepath.Join(docs, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"), nil
		}
		return filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1"), nil
	default:
		return "", fmt.Errorf("global profile editing is not supported for %s", shell)
	}
}

func UpsertMarkedBlock(path string, block string) error {
	const start = "# >>> javahome >>>"
	const end = "# <<< javahome <<<"

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	existingBytes, _ := os.ReadFile(path)
	existing := string(existingBytes)
	block = strings.TrimRight(block, "\n") + "\n"

	startIdx := strings.Index(existing, start)
	endIdx := strings.Index(existing, end)
	if startIdx >= 0 && endIdx > startIdx {
		endIdx += len(end)
		replacement := strings.TrimRight(block, "\n")
		updated := existing[:startIdx] + replacement + existing[endIdx:]
		return os.WriteFile(path, []byte(updated), 0o644)
	}

	if strings.TrimSpace(existing) != "" && !strings.HasSuffix(existing, "\n") {
		existing += "\n"
	}
	if strings.TrimSpace(existing) != "" {
		existing += "\n"
	}
	existing += block
	return os.WriteFile(path, []byte(existing), 0o644)
}

func shQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func psQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func fishQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "\\'") + "'"
}

func fishPathList(pathValue string) []string {
	sep := PathSeparator()
	raw := strings.Split(pathValue, sep)
	out := []string{}
	for _, part := range raw {
		if strings.TrimSpace(part) == "" {
			continue
		}
		out = append(out, fishQuote(part))
	}
	return out
}
