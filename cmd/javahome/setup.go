package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mrAibo/javahome/internal/javaenv"
	"github.com/mrAibo/javahome/internal/shellenv"
	"github.com/mrAibo/javahome/internal/termui"
)

func cmdSetup(args []string) error {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	shellName := fs.String("shell", "", "shell to configure: bash, zsh, fish, powershell")
	dryRun := fs.Bool("dry-run", false, "show planned changes without writing files")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"shell": true})); err != nil {
		return err
	}

	ui := termui.New(os.Stdout)
	reader := bufio.NewReader(os.Stdin)

	shell, err := shellenv.ParseShell(*shellName)
	if err != nil {
		return err
	}
	if *shellName == "" {
		fmt.Printf("%s %s\n", ui.Bold("Detected shell:"), ui.Cyan(string(shell)))
		if !promptYesNo(reader, "Use this shell?", true) {
			shell, err = promptShell(reader)
			if err != nil {
				return err
			}
		}
	}

	installs := javaenv.Discover()
	if len(installs) == 0 {
		return errors.New("no Java installations found; install a JDK first or run javahome doctor")
	}
	inst, err := promptInstallation(reader, ui, installs)
	if err != nil {
		return err
	}

	mode, err := promptChoice(reader, "How should javahome apply this JDK?", []string{
		"Show current-shell activation command only",
		"Make it permanent in the shell profile",
		"Write project-local .javahome.toml",
	}, 2)
	if err != nil {
		return err
	}

	versionArg := strconv.Itoa(inst.Major)
	shellArg := string(shell)
	switch mode {
	case 1:
		if err := applyInstallation(inst, versionArg, "", shellArg, false, false, *dryRun, true); err != nil {
			return err
		}
	case 2:
		if err := applyInstallation(inst, versionArg, "", shellArg, true, false, *dryRun, true); err != nil {
			return err
		}
	case 3:
		if err := applyInstallation(inst, versionArg, "", shellArg, false, true, *dryRun, true); err != nil {
			return err
		}
	}

	if shell != shellenv.Cmd && promptYesNo(reader, "Install the short jhome helper?", mode == 2) {
		if err := installHelper(shell, *dryRun); err != nil {
			return err
		}
	}

	if shell != shellenv.Cmd && promptYesNo(reader, "Install shell completion?", false) {
		if err := installCompletion(shell, *dryRun); err != nil {
			return err
		}
	}

	if promptYesNo(reader, "Run javahome doctor now?", true) {
		fmt.Println()
		return cmdDoctor([]string{})
	}
	return nil
}

func cmdUninstall(args []string) error {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	shellName := fs.String("shell", "", "shell profile to clean: bash, zsh, fish, powershell")
	all := fs.Bool("all", false, "clean all supported shell profiles")
	dryRun := fs.Bool("dry-run", false, "show planned changes without writing files")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"shell": true})); err != nil {
		return err
	}

	shells := []shellenv.Shell{}
	if *all {
		shells = []shellenv.Shell{shellenv.Bash, shellenv.Zsh, shellenv.Fish, shellenv.PowerShell}
	} else {
		shell, err := shellenv.ParseShell(*shellName)
		if err != nil {
			return err
		}
		if shell == shellenv.Cmd {
			return errors.New("cmd has no supported javahome profile blocks to uninstall")
		}
		shells = append(shells, shell)
	}

	ui := termui.New(os.Stdout)
	for _, shell := range shells {
		profile, err := shellenv.ProfilePath(shell)
		if err != nil {
			return err
		}
		if *dryRun {
			fmt.Printf("%s remove javahome blocks from %s\n", ui.Warning("Would"), ui.Path(profile))
			continue
		}
		removed, backup, err := shellenv.RemoveJavahomeBlocks(profile)
		if err != nil {
			return err
		}
		if removed == 0 {
			fmt.Printf("%s No javahome blocks found in %s\n", ui.Warning("SKIP"), ui.Path(profile))
			continue
		}
		fmt.Printf("%s Removed %d javahome block(s) from %s\n", ui.Success("OK"), removed, ui.Path(profile))
		if backup != "" {
			fmt.Printf("%s Backup: %s\n", ui.Bullet(""), ui.Path(backup))
		}
	}
	return nil
}

func promptInstallation(reader *bufio.Reader, ui termui.UI, installs []javaenv.Installation) (javaenv.Installation, error) {
	fmt.Println()
	fmt.Println(ui.Bold("Choose Java installation"))
	fmt.Println()
	for i, inst := range installs {
		current := " "
		if inst.IsCurrent {
			current = "*"
		}
		fmt.Printf("%2d) %s Java %-3d %-18s %s\n", i+1, current, inst.Major, emptyDash(inst.Vendor), ui.Path(inst.Path))
	}
	choice, err := promptChoice(reader, "Enter number", installLabels(installs), 1)
	if err != nil {
		return javaenv.Installation{}, err
	}
	return installs[choice-1], nil
}

func installLabels(installs []javaenv.Installation) []string {
	labels := make([]string, len(installs))
	for i, inst := range installs {
		labels[i] = fmt.Sprintf("Java %d %s", inst.Major, inst.Path)
	}
	return labels
}

func promptShell(reader *bufio.Reader) (shellenv.Shell, error) {
	choice, err := promptChoice(reader, "Choose shell", []string{"bash", "zsh", "fish", "powershell"}, 1)
	if err != nil {
		return "", err
	}
	return shellenv.ParseShell([]string{"bash", "zsh", "fish", "powershell"}[choice-1])
}

func promptChoice(reader *bufio.Reader, prompt string, options []string, defaultIndex int) (int, error) {
	if len(options) == 0 {
		return 0, errors.New("no choices available")
	}
	if defaultIndex < 1 || defaultIndex > len(options) {
		defaultIndex = 1
	}
	for {
		fmt.Println()
		fmt.Println(prompt + ":")
		for i, option := range options {
			defaultMark := ""
			if i+1 == defaultIndex {
				defaultMark = " [default]"
			}
			fmt.Printf("  %d) %s%s\n", i+1, option, defaultMark)
		}
		fmt.Printf("Choice [%d]: ", defaultIndex)
		line, err := reader.ReadString('\n')
		if err != nil && strings.TrimSpace(line) == "" {
			return 0, err
		}
		value := strings.TrimSpace(line)
		if value == "" {
			return defaultIndex, nil
		}
		choice, err := strconv.Atoi(value)
		if err == nil && choice >= 1 && choice <= len(options) {
			return choice, nil
		}
		fmt.Printf("Invalid choice %q. Please enter 1-%d.\n", value, len(options))
	}
}

func promptYesNo(reader *bufio.Reader, question string, defaultYes bool) bool {
	def := "y/N"
	if defaultYes {
		def = "Y/n"
	}
	for {
		fmt.Printf("%s [%s]: ", question, def)
		line, err := reader.ReadString('\n')
		if err != nil && strings.TrimSpace(line) == "" {
			return defaultYes
		}
		value := strings.ToLower(strings.TrimSpace(line))
		if value == "" {
			return defaultYes
		}
		switch value {
		case "y", "yes", "j", "ja":
			return true
		case "n", "no", "nein":
			return false
		}
		fmt.Println("Please answer yes or no.")
	}
}

func installHelper(shell shellenv.Shell, dryRun bool) error {
	profile, err := shellenv.ProfilePath(shell)
	if err != nil {
		return err
	}
	script := shellenv.InitScript(shell)
	ui := termui.New(os.Stdout)
	if dryRun {
		fmt.Printf("%s install jhome helper into %s:\n\n%s", ui.Warning("Would"), ui.Path(profile), script)
		return nil
	}
	if err := shellenv.UpsertMarkedBlock(profile, script); err != nil {
		return err
	}
	fmt.Printf("%s Installed jhome helper in %s\n", ui.Success("OK"), ui.Path(profile))
	return nil
}

func installCompletion(shell shellenv.Shell, dryRun bool) error {
	script, err := completionScript(shell)
	if err != nil {
		return err
	}
	path, err := completionFilePath(shell)
	if err != nil {
		return err
	}
	ui := termui.New(os.Stdout)
	if dryRun {
		fmt.Printf("%s write completion script to %s\n", ui.Warning("Would"), ui.Path(path))
	} else {
		backup, err := shellenv.WriteFileWithBackup(path, []byte(script), 0o644)
		if err != nil {
			return err
		}
		fmt.Printf("%s Wrote completion script to %s\n", ui.Success("OK"), ui.Path(path))
		if backup != "" {
			fmt.Printf("%s Backup: %s\n", ui.Bullet(""), ui.Path(backup))
		}
	}

	block, err := completionProfileBlock(shell, path)
	if err != nil {
		return err
	}
	if block == "" {
		return nil
	}
	profile, err := shellenv.ProfilePath(shell)
	if err != nil {
		return err
	}
	if dryRun {
		fmt.Printf("%s source completion from %s:\n\n%s", ui.Warning("Would"), ui.Path(profile), block)
		return nil
	}
	if err := shellenv.UpsertMarkedBlock(profile, block); err != nil {
		return err
	}
	fmt.Printf("%s Added completion loader to %s\n", ui.Success("OK"), ui.Path(profile))
	return nil
}

func completionScript(shell shellenv.Shell) (string, error) {
	switch shell {
	case shellenv.Bash:
		return bashCompletion(), nil
	case shellenv.Zsh:
		return zshCompletion(), nil
	case shellenv.Fish:
		return fishCompletion(), nil
	case shellenv.PowerShell:
		return powerShellCompletion(), nil
	default:
		return "", fmt.Errorf("completion install is not supported for %s", shell)
	}
}

func completionFilePath(shell shellenv.Shell) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch shell {
	case shellenv.Bash:
		return filepath.Join(home, ".javahome-completion.bash"), nil
	case shellenv.Zsh:
		return filepath.Join(home, ".javahome-completion.zsh"), nil
	case shellenv.Fish:
		return filepath.Join(home, ".config", "fish", "completions", "javahome.fish"), nil
	case shellenv.PowerShell:
		return filepath.Join(home, ".javahome-completion.ps1"), nil
	default:
		return "", fmt.Errorf("completion install is not supported for %s", shell)
	}
}

func completionProfileBlock(shell shellenv.Shell, completionPath string) (string, error) {
	switch shell {
	case shellenv.Bash:
		return "# >>> javahome completion >>>\n[ -f " + shellSingleQuote(completionPath) + " ] && source " + shellSingleQuote(completionPath) + "\n# <<< javahome completion <<<\n", nil
	case shellenv.Zsh:
		return "# >>> javahome completion >>>\n[ -f " + shellSingleQuote(completionPath) + " ] && source " + shellSingleQuote(completionPath) + "\n# <<< javahome completion <<<\n", nil
	case shellenv.PowerShell:
		return "# >>> javahome completion >>>\nif (Test-Path " + psSingleQuote(completionPath) + ") { . " + psSingleQuote(completionPath) + " }\n# <<< javahome completion <<<\n", nil
	case shellenv.Fish:
		return "", nil
	default:
		return "", fmt.Errorf("completion profile block is not supported for %s", shell)
	}
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func psSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
