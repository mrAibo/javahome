package shellenv

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
		return "# >>> javahome helper >>>\nfunction jhome\n    javahome use $argv --shell fish | source\nend\n# <<< javahome helper <<<\n"
	case PowerShell:
		return "# >>> javahome helper >>>\nfunction jhome { javahome use @args --shell powershell | Invoke-Expression }\n# <<< javahome helper <<<\n"
	case Cmd:
		return ":: cmd.exe cannot safely eval command output. Use: javahome use 17 --global --shell cmd\n"
	case Zsh:
		return "# >>> javahome helper >>>\njhome() { eval \"$(javahome use \"$@\" --shell zsh)\"; }\n# <<< javahome helper <<<\n"
	default:
		return "# >>> javahome helper >>>\njhome() { eval \"$(javahome use \"$@\" --shell bash)\"; }\n# <<< javahome helper <<<\n"
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
	start, end := blockMarkers(block)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	existingBytes, _ := os.ReadFile(path)
	existing := string(existingBytes)
	block = strings.TrimRight(block, "\n") + "\n"

	startIdx := strings.Index(existing, start)
	endIdx := strings.Index(existing, end)
	updated := existing
	if startIdx >= 0 && endIdx > startIdx {
		endIdx += len(end)
		replacement := strings.TrimRight(block, "\n")
		updated = existing[:startIdx] + replacement + existing[endIdx:]
	} else {
		if strings.TrimSpace(updated) != "" && !strings.HasSuffix(updated, "\n") {
			updated += "\n"
		}
		if strings.TrimSpace(updated) != "" {
			updated += "\n"
		}
		updated += block
	}

	if updated == existing {
		return nil
	}
	if _, err := BackupFile(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(updated), 0o644)
}

func RemoveJavahomeBlocks(path string) (int, string, error) {
	existingBytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, "", nil
		}
		return 0, "", err
	}
	existing := string(existingBytes)
	lines := strings.SplitAfter(existing, "\n")
	out := make([]string, 0, len(lines))
	inBlock := false
	removed := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isJavahomeStartMarker(trimmed) {
			inBlock = true
			removed++
			continue
		}
		if inBlock {
			if isJavahomeEndMarker(trimmed) {
				inBlock = false
			}
			continue
		}
		out = append(out, line)
	}
	if removed == 0 {
		return 0, "", nil
	}
	updated := strings.Join(out, "")
	backup, err := BackupFile(path)
	if err != nil {
		return 0, "", err
	}
	return removed, backup, os.WriteFile(path, []byte(updated), 0o644)
}

func WriteFileWithBackup(path string, data []byte, perm os.FileMode) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	backup, err := BackupFile(path)
	if err != nil {
		return "", err
	}
	return backup, os.WriteFile(path, data, perm)
}

func BackupFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("cannot back up directory %s", path)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	stamp := time.Now().Format("20060102-150405")
	backup := path + ".javahome-backup-" + stamp
	for i := 0; i < 1000; i++ {
		candidate := backup
		if i > 0 {
			candidate = fmt.Sprintf("%s-%03d", backup, i)
		}
		_, err := os.Stat(candidate)
		if os.IsNotExist(err) {
			return candidate, os.WriteFile(candidate, content, info.Mode().Perm())
		}
		if err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("could not create unique backup for %s", path)
}

func blockMarkers(block string) (string, string) {
	start := "# >>> javahome >>>"
	end := "# <<< javahome <<<"
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, ">>> javahome") {
			start = trimmed
		}
		if strings.Contains(trimmed, "<<< javahome") {
			end = trimmed
		}
	}
	return start, end
}

func isJavahomeStartMarker(line string) bool {
	return strings.Contains(line, ">>> javahome")
}

func isJavahomeEndMarker(line string) bool {
	return strings.Contains(line, "<<< javahome")
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
