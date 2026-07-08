package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/mrAibo/javahome/internal/javaenv"
	"github.com/mrAibo/javahome/internal/shellenv"
	"github.com/mrAibo/javahome/internal/termui"
)

const version = "0.5.1"

func main() {
	if err := run(os.Args[1:]); err != nil {
		ui := termui.New(os.Stderr)
		fmt.Fprintf(os.Stderr, "%s %v\n", ui.Error("error:"), err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printHelp()
		return nil
	}

	switch args[0] {
	case "list", "ls":
		return cmdList(args[1:])
	case "current":
		return cmdCurrent(args[1:])
	case "print":
		return cmdPrint(args[1:])
	case "use":
		return cmdUse(args[1:])
	case "select":
		return cmdSelect(args[1:])
	case "activate":
		return cmdActivate(args[1:])
	case "setup", "wizard":
		return cmdSetup(args[1:])
	case "uninstall":
		return cmdUninstall(args[1:])
	case "windows-env", "winenv":
		return cmdWindowsEnv(args[1:])
	case "completion", "completions":
		return cmdCompletion(args[1:])
	case "doctor":
		return cmdDoctor(args[1:])
	case "init":
		return cmdInit(args[1:])
	case "version", "--version", "-v":
		fmt.Println(version)
		return nil
	case "help", "--help", "-h":
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func cmdList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}

	installs := javaenv.Discover()
	if *jsonOut {
		return printJSON(installs)
	}

	ui := termui.New(os.Stdout)
	if len(installs) == 0 {
		fmt.Println(ui.Warning("No Java installations found."))
		fmt.Println(ui.Bullet("Install a JDK or run `javahome doctor` for diagnostics."))
		return nil
	}

	fmt.Println(ui.Bold("Discovered Java installations"))
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CURRENT\tMAJOR\tVERSION\tVENDOR\tSOURCE\tPATH")
	for _, inst := range installs {
		current := ""
		if inst.IsCurrent {
			current = "*"
		}
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\n", current, inst.Major, emptyDash(inst.Version), emptyDash(inst.Vendor), inst.Source, inst.Path)
	}
	return w.Flush()
}

func cmdCurrent(args []string) error {
	fs := flag.NewFlagSet("current", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}

	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		if *jsonOut {
			return printJSON(map[string]any{"java_home": "", "valid": false})
		}
		ui := termui.New(os.Stdout)
		fmt.Println(ui.Warning("JAVA_HOME is not set."))
		fmt.Println(ui.Bullet("Run `javahome list`, then activate a version with `javahome use <version>`."))
		return nil
	}

	inst, ok := javaenv.InstallationFromHome(javaHome, "JAVA_HOME", javaHome)
	if *jsonOut {
		if ok {
			return printJSON(inst)
		}
		return printJSON(map[string]any{"java_home": javaHome, "valid": false})
	}

	ui := termui.New(os.Stdout)
	if !ok {
		fmt.Printf("%s=%s\n", ui.Key("JAVA_HOME"), javaHome)
		fmt.Printf("%s %s\n", ui.Warning("Status:"), "invalid Java home")
		return nil
	}
	fmt.Printf("%s=%s\n", ui.Key("JAVA_HOME"), ui.Path(inst.Path))
	fmt.Printf("%s %s\n", ui.Key("Version:"), emptyDash(inst.Version))
	fmt.Printf("%s  %s\n", ui.Key("Vendor:"), emptyDash(inst.Vendor))
	fmt.Printf("%s %s\n", ui.Key("Status:"), ui.Success("valid"))
	return nil
}

func cmdPrint(args []string) error {
	fs := flag.NewFlagSet("print", flag.ContinueOnError)
	vendor := fs.String("vendor", "", "filter by vendor text")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"vendor": true})); err != nil {
		return err
	}
	versionArg := ""
	if fs.NArg() > 0 {
		versionArg = fs.Arg(0)
	}

	inst, err := selectInstallation(versionArg, *vendor)
	if err != nil {
		return err
	}
	if *jsonOut {
		return printJSON(inst)
	}
	fmt.Println(inst.Path)
	return nil
}

func cmdUse(args []string) error {
	fs := flag.NewFlagSet("use", flag.ContinueOnError)
	vendor := fs.String("vendor", "", "filter by vendor text")
	shellName := fs.String("shell", "", "shell to emit: bash, zsh, fish, powershell, cmd")
	global := fs.Bool("global", false, "write activation block to the user's shell profile")
	project := fs.Bool("project", false, "write .javahome.toml in the current directory")
	dryRun := fs.Bool("dry-run", false, "show planned changes without writing files")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"vendor": true, "shell": true})); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return errors.New("usage: javahome use <major-version> [--vendor text] [--shell bash|zsh|fish|powershell|cmd] [--global|--project|--dry-run]")
	}

	versionArg := fs.Arg(0)
	inst, err := selectInstallation(versionArg, *vendor)
	if err != nil {
		return err
	}
	return applyInstallation(inst, versionArg, *vendor, *shellName, *global, *project, *dryRun, true)
}

func cmdDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}

	type Check struct {
		Name    string `json:"name"`
		OK      bool   `json:"ok"`
		Message string `json:"message"`
	}

	checks := []Check{}
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		checks = append(checks, Check{"JAVA_HOME", false, "JAVA_HOME is not set"})
	} else if javaenv.IsJavaHome(javaHome) {
		checks = append(checks, Check{"JAVA_HOME", true, javaHome})
	} else {
		checks = append(checks, Check{"JAVA_HOME", false, javaHome + " is not a valid Java home"})
	}

	if javaHome != "" {
		javaBin := filepath.Join(javaHome, "bin", javaBinary())
		if exists(javaBin) {
			checks = append(checks, Check{"java", true, javaBin})
		} else {
			checks = append(checks, Check{"java", false, javaBin + " not found"})
		}
		javacBin := filepath.Join(javaHome, "bin", javacBinary())
		if exists(javacBin) {
			checks = append(checks, Check{"javac", true, javacBin})
		} else {
			checks = append(checks, Check{"javac", false, javacBin + " not found; this may be a JRE, not a JDK"})
		}
	}

	installs := javaenv.Discover()
	if len(installs) > 0 {
		checks = append(checks, Check{"discovery", true, fmt.Sprintf("found %d Java installation(s)", len(installs))})
	} else {
		checks = append(checks, Check{"discovery", false, "no Java installations found"})
	}

	for _, manager := range javaenv.DetectManagers(os.Getenv("PATH")) {
		if !manager.Found {
			continue
		}
		checks = append(checks, Check{"manager:" + manager.Name, !manager.Active, manager.Message})
	}

	if *jsonOut {
		return printJSON(checks)
	}
	ui := termui.New(os.Stdout)
	fmt.Println(ui.Bold("javahome doctor"))
	fmt.Println()
	for _, check := range checks {
		mark := "WARN"
		if check.OK {
			mark = "OK"
		}
		fmt.Printf("%s%s %-12s %s\n", ui.Check(check.OK), strings.Repeat(" ", 5-len(mark)), check.Name, check.Message)
	}
	return nil
}

func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	global := fs.Bool("global", false, "write init helper to shell profile")
	dryRun := fs.Bool("dry-run", false, "show planned changes without writing files")
	if err := fs.Parse(reorderArgs(args, nil)); err != nil {
		return err
	}
	shellName := ""
	if fs.NArg() > 0 {
		shellName = fs.Arg(0)
	}
	shell, err := shellenv.ParseShell(shellName)
	if err != nil {
		return err
	}
	if *global {
		return installHelper(shell, *dryRun)
	}
	fmt.Print(shellenv.InitScript(shell))
	return nil
}

func reorderArgs(args []string, valueFlags map[string]bool) []string {
	if len(args) == 0 {
		return args
	}
	if valueFlags == nil {
		valueFlags = map[string]bool{}
	}

	flags := []string{}
	positionals := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if strings.HasPrefix(arg, "-") && arg != "-" {
			flags = append(flags, arg)
			name := strings.TrimLeft(arg, "-")
			if eq := strings.Index(name, "="); eq >= 0 {
				name = name[:eq]
			}
			if !strings.Contains(arg, "=") && valueFlags[name] && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		positionals = append(positionals, arg)
	}
	return append(flags, positionals...)
}

func selectInstallation(versionArg string, vendor string) (javaenv.Installation, error) {
	installs := javaenv.Discover()
	inst, ok := javaenv.Select(installs, versionArg, vendor)
	if ok {
		return inst, nil
	}

	majors := map[int]bool{}
	for _, i := range installs {
		if i.Major > 0 {
			majors[i.Major] = true
		}
	}
	vals := make([]int, 0, len(majors))
	for major := range majors {
		vals = append(vals, major)
	}
	sort.Ints(vals)
	parts := []string{}
	for _, major := range vals {
		parts = append(parts, strconv.Itoa(major))
	}
	if len(parts) == 0 {
		return javaenv.Installation{}, fmt.Errorf("no Java installations found")
	}
	return javaenv.Installation{}, fmt.Errorf("no matching Java installation found for version %q. Available majors: %s", versionArg, strings.Join(parts, ", "))
}

func projectConfig(inst javaenv.Installation, vendorFilter string) string {
	lines := []string{
		"# Generated by javahome.",
		fmt.Sprintf("version = %q", strconv.Itoa(inst.Major)),
		fmt.Sprintf("path = %q", inst.Path),
	}
	if strings.TrimSpace(vendorFilter) != "" {
		lines = append(lines, fmt.Sprintf("vendor = %q", vendorFilter))
	} else if strings.TrimSpace(inst.Vendor) != "" {
		lines = append(lines, fmt.Sprintf("vendor = %q", inst.Vendor))
	}
	return strings.Join(lines, "\n") + "\n"
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func exists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}

func javaBinary() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

func javacBinary() string {
	if runtime.GOOS == "windows" {
		return "javac.exe"
	}
	return "javac"
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func printHelpCommand(ui termui.UI, command string, description string) {
	spaces := 42 - len(command)
	if spaces < 1 {
		spaces = 1
	}
	fmt.Printf("  %s%s %s\n", ui.Command(command), strings.Repeat(" ", spaces), description)
}

func printHelp() {
	ui := termui.New(os.Stdout)
	shell := shellenv.DetectShell()
	fmt.Println(ui.Bold("javahome "+version) + " - " + ui.Cyan("switch Java versions without fragile shell hacks"))
	fmt.Printf("%s %s / %s\n", ui.Bold("Detected:"), runtime.GOOS, shell)
	fmt.Println()
	fmt.Println(ui.Bold("Usage:"))
	fmt.Println("  javahome <command> [options]")
	fmt.Println()
	fmt.Println(ui.Bold("Common commands:"))
	printHelpCommand(ui, "javahome setup", "Guided setup wizard")
	printHelpCommand(ui, "javahome list", "List discovered JDKs")
	printHelpCommand(ui, "javahome current", "Show active JAVA_HOME")
	printHelpCommand(ui, "javahome doctor", "Diagnose Java and PATH issues")
	printHelpCommand(ui, "javahome select", "Choose a JDK from a numbered list")
	printHelpCommand(ui, "javahome print 17", "Print the path for Java 17")
	printHelpCommand(ui, "javahome uninstall", "Remove javahome blocks from profile files")
	fmt.Println()
	printPlatformHelp(ui, shell)
	fmt.Println()
	fmt.Println(ui.Bold("All commands:"))
	fmt.Println("  javahome setup [--shell name] [--dry-run]")
	fmt.Println("  javahome uninstall [--shell name|--all] [--dry-run]")
	fmt.Println("  javahome list [--json]")
	fmt.Println("  javahome current [--json]")
	fmt.Println("  javahome print [version] [--vendor text] [--json]")
	fmt.Println("  javahome use <version> [--vendor text] [--shell bash|zsh|fish|powershell|cmd]")
	fmt.Println("  javahome use <version> --global [--shell bash|zsh|fish|powershell]")
	fmt.Println("  javahome use <version> --project")
	fmt.Println("  javahome select [--vendor text] [--shell name] [--global|--project]")
	fmt.Println("  javahome activate [--file .javahome.toml] [--shell name|--global]")
	fmt.Println("  javahome completion bash|zsh|fish|powershell")
	if runtime.GOOS == "windows" {
		fmt.Println("  javahome windows-env user|machine <version> [--vendor text] [--dry-run]")
	}
	fmt.Println("  javahome doctor [--json]")
	fmt.Println("  javahome init [bash|zsh|fish|powershell] [--global]")
	fmt.Println("  javahome version")
	fmt.Println()
	fmt.Println(ui.Bold("Safety:"))
	fmt.Println("  Profile edits create timestamped .javahome-backup-* files when a profile exists.")
	fmt.Println("  Shell activation output is never colorized, so eval/source/Invoke-Expression stays safe.")
}

func printPlatformHelp(ui termui.UI, shell shellenv.Shell) {
	switch runtime.GOOS {
	case "windows":
		fmt.Println(ui.Bold("Windows / PowerShell examples:"))
		fmt.Println("  " + ui.Command("javahome use 17 --shell powershell | Invoke-Expression"))
		fmt.Println("  " + ui.Command("javahome use 17 --global --shell powershell"))
		fmt.Println("  " + ui.Command("javahome windows-env user 17"))
		fmt.Println("  " + ui.Command("javahome windows-env machine 17"))
		fmt.Println("  " + ui.Command("javahome setup --shell powershell"))
		fmt.Println()
		fmt.Println(ui.Bold("Windows note:"))
		fmt.Println("  Current PowerShell changes are immediate. Persisted Windows user/machine")
		fmt.Println("  environment changes affect new processes. Restart terminals, IDEs, daemons,")
		fmt.Println("  or services that were already running. Machine scope usually needs Admin.")
	case "darwin":
		fmt.Println(ui.Bold("macOS examples:"))
		fmt.Println("  " + ui.Command("eval \"$(javahome use 17 --shell zsh)\""))
		fmt.Println("  " + ui.Command("javahome use 17 --global --shell zsh"))
		fmt.Println("  " + ui.Command("javahome activate --shell zsh"))
		fmt.Println("  " + ui.Command("javahome setup --shell zsh"))
		fmt.Println()
		fmt.Println(ui.Bold("macOS discovery:"))
		fmt.Println("  Uses /usr/libexec/java_home -V plus common JDK, Homebrew, SDKMAN, asdf, and mise paths.")
	default:
		fmt.Println(ui.Bold("Linux examples:"))
		switch shell {
		case shellenv.Fish:
			fmt.Println("  " + ui.Command("javahome use 17 --shell fish | source"))
			fmt.Println("  " + ui.Command("javahome use 17 --global --shell fish"))
		case shellenv.Zsh:
			fmt.Println("  " + ui.Command("eval \"$(javahome use 17 --shell zsh)\""))
			fmt.Println("  " + ui.Command("javahome use 17 --global --shell zsh"))
		default:
			fmt.Println("  " + ui.Command("eval \"$(javahome use 17 --shell bash)\""))
			fmt.Println("  " + ui.Command("javahome use 17 --global --shell bash"))
		}
		fmt.Println("  " + ui.Command("javahome setup"))
		fmt.Println("  " + ui.Command("JAVA_HOME=\"$(javahome print 17)\""))
		fmt.Println()
		fmt.Println(ui.Bold("Linux discovery:"))
		fmt.Println("  Uses update-alternatives, update-java-alternatives, /usr/lib/jvm, /opt, SDKMAN, asdf, and mise paths.")
	}
}
