package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mrAibo/javahome/internal/javaenv"
	"github.com/mrAibo/javahome/internal/shellenv"
	"github.com/mrAibo/javahome/internal/termui"
)

func cmdSelect(args []string) error {
	fs := flag.NewFlagSet("select", flag.ContinueOnError)
	vendor := fs.String("vendor", "", "filter by vendor text")
	shellName := fs.String("shell", "", "shell to emit: bash, zsh, fish, powershell, cmd")
	global := fs.Bool("global", false, "write activation block to the user's shell profile")
	project := fs.Bool("project", false, "write .javahome.toml in the current directory")
	dryRun := fs.Bool("dry-run", false, "show planned changes without writing files")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"vendor": true, "shell": true})); err != nil {
		return err
	}

	installs := javaenv.Discover()
	if strings.TrimSpace(*vendor) != "" {
		filtered := []javaenv.Installation{}
		needle := strings.ToLower(strings.TrimSpace(*vendor))
		for _, inst := range installs {
			if strings.Contains(strings.ToLower(inst.Vendor), needle) || strings.Contains(strings.ToLower(inst.Path), needle) {
				filtered = append(filtered, inst)
			}
		}
		installs = filtered
	}
	if len(installs) == 0 {
		return errors.New("no Java installations found")
	}

	ui := termui.New(os.Stdout)
	fmt.Println(ui.Bold("Select Java installation"))
	fmt.Println()
	for i, inst := range installs {
		current := " "
		if inst.IsCurrent {
			current = "*"
		}
		fmt.Printf("%2d) %s Java %-3d %-18s %s\n", i+1, current, inst.Major, emptyDash(inst.Vendor), ui.Path(inst.Path))
	}
	fmt.Println()
	fmt.Print("Enter number: ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return err
	}
	choice, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || choice < 1 || choice > len(installs) {
		return fmt.Errorf("invalid selection %q", strings.TrimSpace(line))
	}
	inst := installs[choice-1]
	return applyInstallation(inst, strconv.Itoa(inst.Major), *vendor, *shellName, *global, *project, *dryRun, true)
}

func cmdActivate(args []string) error {
	fs := flag.NewFlagSet("activate", flag.ContinueOnError)
	file := fs.String("file", ".javahome.toml", "project config file")
	shellName := fs.String("shell", "", "shell to emit: bash, zsh, fish, powershell, cmd")
	global := fs.Bool("global", false, "write activation block to the user's shell profile")
	dryRun := fs.Bool("dry-run", false, "show planned changes without writing files")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"file": true, "shell": true})); err != nil {
		return err
	}

	cfg, err := readProjectConfig(*file)
	if err != nil {
		return err
	}
	inst, err := installationFromConfig(cfg)
	if err != nil {
		return err
	}
	return applyInstallation(inst, cfg.Version, cfg.Vendor, *shellName, *global, false, *dryRun, false)
}

func cmdCompletion(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: javahome completion bash|zsh|fish|powershell")
	}
	switch strings.ToLower(args[0]) {
	case "bash":
		fmt.Print(bashCompletion())
	case "zsh":
		fmt.Print(zshCompletion())
	case "fish":
		fmt.Print(fishCompletion())
	case "powershell", "pwsh", "ps":
		fmt.Print(powerShellCompletion())
	default:
		return fmt.Errorf("unsupported completion shell %q", args[0])
	}
	return nil
}

func applyInstallation(inst javaenv.Installation, versionArg string, vendor string, shellName string, global bool, project bool, dryRun bool, showHints bool) error {
	shell, err := shellenv.ParseShell(shellName)
	if err != nil {
		return err
	}
	newPath := shellenv.CleanPathForJava(os.Getenv("PATH"), inst.Path)
	script := shellenv.ActivationScript(shell, inst.Path, newPath)

	if project {
		content := projectConfig(inst, vendor)
		if dryRun {
			fmt.Print(content)
			return nil
		}
		if err := os.WriteFile(".javahome.toml", []byte(content), 0o644); err != nil {
			return err
		}
		ui := termui.New(os.Stdout)
		fmt.Printf("%s Wrote %s for Java %d\n", ui.Success("OK"), ui.Path(".javahome.toml"), inst.Major)
		return nil
	}

	if global {
		path, err := shellenv.ProfilePath(shell)
		if err != nil {
			return err
		}
		block := "# >>> javahome >>>\n" + strings.TrimRight(script, "\n") + "\n# <<< javahome <<<\n"
		if dryRun {
			ui := termui.New(os.Stdout)
			fmt.Printf("%s %s with:\n\n%s", ui.Warning("Would update"), ui.Path(path), block)
			return nil
		}
		if err := shellenv.UpsertMarkedBlock(path, block); err != nil {
			return err
		}
		ui := termui.New(os.Stdout)
		fmt.Printf("%s Updated %s\n", ui.Success("OK"), ui.Path(path))
		fmt.Println(ui.Bullet("Open a new shell or reload your profile to apply the change."))
		return nil
	}

	if shellName != "" {
		// This output is intended to be evaluated by a shell. Never add color here.
		fmt.Print(script)
		return nil
	}

	ui := termui.New(os.Stdout)
	fmt.Printf("%s Selected Java %s: %s\n", ui.Success("OK"), strconv.Itoa(inst.Major), ui.Path(inst.Path))
	fmt.Println()
	cmd := activationCommand(versionArg, shell)
	permanentCmd := fmt.Sprintf("javahome use %s --global --shell %s", versionArg, shell)
	if !showHints {
		cmd = activateCommand(shell)
		permanentCmd = fmt.Sprintf("javahome activate --global --shell %s", shell)
	}
	fmt.Printf("%s\n  %s\n", ui.Bold("For the current shell, run:"), ui.Command(cmd))
	fmt.Println()
	fmt.Printf("%s\n  %s\n", ui.Bold("To make it permanent, run:"), ui.Command(permanentCmd))
	return nil
}

func activateCommand(shell shellenv.Shell) string {
	switch shell {
	case shellenv.Fish:
		return "javahome activate --shell fish | source"
	case shellenv.PowerShell:
		return "javahome activate --shell powershell | Invoke-Expression"
	case shellenv.Cmd:
		return "javahome activate --shell cmd"
	case shellenv.Zsh:
		return "eval \"$(javahome activate --shell zsh)\""
	default:
		return "eval \"$(javahome activate --shell bash)\""
	}
}

func activationCommand(versionArg string, shell shellenv.Shell) string {
	switch shell {
	case shellenv.Fish:
		return fmt.Sprintf("javahome use %s --shell fish | source", versionArg)
	case shellenv.PowerShell:
		return fmt.Sprintf("javahome use %s --shell powershell | Invoke-Expression", versionArg)
	case shellenv.Cmd:
		return fmt.Sprintf("javahome use %s --shell cmd", versionArg)
	case shellenv.Zsh:
		return fmt.Sprintf("eval \"$(javahome use %s --shell zsh)\"", versionArg)
	default:
		return fmt.Sprintf("eval \"$(javahome use %s --shell bash)\"", versionArg)
	}
}

type ProjectConfig struct {
	Version string
	Vendor  string
	Path    string
}

func readProjectConfig(path string) (ProjectConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return ProjectConfig{}, fmt.Errorf("could not read %s: %w", path, err)
	}
	cfg := ProjectConfig{}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		switch key {
		case "version":
			cfg.Version = value
		case "vendor":
			cfg.Vendor = value
		case "path":
			cfg.Path = value
		}
	}
	if cfg.Version == "" && cfg.Path == "" {
		return ProjectConfig{}, fmt.Errorf("%s does not define version or path", path)
	}
	return cfg, nil
}

func installationFromConfig(cfg ProjectConfig) (javaenv.Installation, error) {
	if strings.TrimSpace(cfg.Path) != "" {
		if inst, ok := javaenv.InstallationFromHome(cfg.Path, "project", os.Getenv("JAVA_HOME")); ok {
			return inst, nil
		}
	}
	return selectInstallation(cfg.Version, cfg.Vendor)
}

func bashCompletion() string {
	return `_javahome_complete() {
  local cur="${COMP_WORDS[COMP_CWORD]}"
  local prev="${COMP_WORDS[COMP_CWORD-1]}"
  local commands="list ls current print use select activate doctor init completion version help"
  local shells="bash zsh fish powershell cmd"
  case "$prev" in
    --shell|init|completion) COMPREPLY=( $(compgen -W "$shells" -- "$cur") ); return ;;
    --vendor) COMPREPLY=(); return ;;
  esac
  if [[ $COMP_CWORD -eq 1 ]]; then
    COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
  else
    COMPREPLY=( $(compgen -W "--json --vendor --shell --global --project --dry-run --file" -- "$cur") )
  fi
}
complete -F _javahome_complete javahome
`
}

func zshCompletion() string {
	return `#compdef javahome
_arguments \
  '1:command:(list ls current print use select activate doctor init completion version help)' \
  '--json[print JSON]' \
  '--vendor[filter by vendor text]:vendor:' \
  '--shell[shell to emit]:shell:(bash zsh fish powershell cmd)' \
  '--global[write to shell profile]' \
  '--project[write project config]' \
  '--dry-run[preview changes]' \
  '--file[project config file]:file:_files'
`
}

func fishCompletion() string {
	return `complete -c javahome -f
complete -c javahome -n "not __fish_seen_subcommand_from list ls current print use select activate doctor init completion version help" -a "list ls current print use select activate doctor init completion version help"
complete -c javahome -l json -d "print JSON"
complete -c javahome -l vendor -r -d "filter by vendor text"
complete -c javahome -l shell -r -a "bash zsh fish powershell cmd" -d "shell to emit"
complete -c javahome -l global -d "write to shell profile"
complete -c javahome -l project -d "write .javahome.toml"
complete -c javahome -l dry-run -d "preview changes"
complete -c javahome -l file -r -d "project config file"
`
}

func powerShellCompletion() string {
	return `Register-ArgumentCompleter -Native -CommandName javahome -ScriptBlock {
  param($wordToComplete, $commandAst, $cursorPosition)
  $commands = 'list','ls','current','print','use','select','activate','doctor','init','completion','version','help'
  $flags = '--json','--vendor','--shell','--global','--project','--dry-run','--file'
  $values = 'bash','zsh','fish','powershell','cmd'
  ($commands + $flags + $values) | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
    [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
  }
}
`
}
