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
)

const version = "0.1.0"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
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
	if len(installs) == 0 {
		fmt.Println("No Java installations found.")
		return nil
	}

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
		fmt.Println("JAVA_HOME is not set.")
		return nil
	}

	inst, ok := javaenv.InstallationFromHome(javaHome, "JAVA_HOME", javaHome)
	if *jsonOut {
		if ok {
			return printJSON(inst)
		}
		return printJSON(map[string]any{"java_home": javaHome, "valid": false})
	}

	if !ok {
		fmt.Printf("JAVA_HOME=%s\n", javaHome)
		fmt.Println("Status: invalid Java home")
		return nil
	}
	fmt.Printf("JAVA_HOME=%s\n", inst.Path)
	fmt.Printf("Version: %s\n", emptyDash(inst.Version))
	fmt.Printf("Vendor:  %s\n", emptyDash(inst.Vendor))
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
	shell, err := shellenv.ParseShell(*shellName)
	if err != nil {
		return err
	}
	newPath := shellenv.CleanPathForJava(os.Getenv("PATH"), inst.Path)
	script := shellenv.ActivationScript(shell, inst.Path, newPath)

	if *project {
		content := projectConfig(inst, *vendor)
		if *dryRun {
			fmt.Print(content)
			return nil
		}
		return os.WriteFile(".javahome.toml", []byte(content), 0o644)
	}

	if *global {
		path, err := shellenv.ProfilePath(shell)
		if err != nil {
			return err
		}
		block := "# >>> javahome >>>\n" + strings.TrimRight(script, "\n") + "\n# <<< javahome <<<\n"
		if *dryRun {
			fmt.Printf("Would update %s with:\n\n%s", path, block)
			return nil
		}
		if err := shellenv.UpsertMarkedBlock(path, block); err != nil {
			return err
		}
		fmt.Printf("Updated %s\n", path)
		fmt.Println("Open a new shell or reload your profile to apply the change.")
		return nil
	}

	if *shellName != "" {
		fmt.Print(script)
		return nil
	}

	fmt.Printf("Selected Java %s: %s\n", strconv.Itoa(inst.Major), inst.Path)
	fmt.Println()
	fmt.Printf("For the current shell, run:\n  eval \"$(javahome use %s --shell %s)\"\n", versionArg, shell)
	fmt.Println()
	fmt.Printf("To make it permanent, run:\n  javahome use %s --global --shell %s\n", versionArg, shell)
	return nil
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

	if *jsonOut {
		return printJSON(checks)
	}
	for _, check := range checks {
		mark := "OK "
		if !check.OK {
			mark = "WARN"
		}
		fmt.Printf("%-5s %-12s %s\n", mark, check.Name, check.Message)
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
	script := shellenv.InitScript(shell)
	if *global {
		path, err := shellenv.ProfilePath(shell)
		if err != nil {
			return err
		}
		if *dryRun {
			fmt.Printf("Would update %s with:\n\n%s", path, script)
			return nil
		}
		if err := shellenv.UpsertMarkedBlock(path, script); err != nil {
			return err
		}
		fmt.Printf("Updated %s\n", path)
		return nil
	}
	fmt.Print(script)
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

func printHelp() {
	fmt.Println(`javahome ` + version + `

Switch Java versions without fragile shell hacks.

Usage:
  javahome <command> [options]

Most useful commands:
  javahome list                         List discovered JDKs
  javahome current                      Show active JAVA_HOME
  javahome doctor                       Diagnose JAVA_HOME, java, javac, and discovery
  javahome print 17                     Print the path for Java 17
  javahome use 17                       Show activation instructions for your shell

Current-shell activation:
  eval "$(javahome use 17 --shell bash)"
  eval "$(javahome use 17 --shell zsh)"
  javahome use 17 --shell fish | source
  javahome use 17 --shell powershell | Invoke-Expression

Permanent profile update:
  javahome use 17 --global --shell bash
  javahome use 17 --global --shell zsh
  javahome use 17 --global --shell fish
  javahome use 17 --global --shell powershell

Project and automation:
  javahome use 17 --project             Write .javahome.toml
  javahome list --json                  JSON output for scripts
  javahome use 17 --global --dry-run    Preview profile changes

All commands:
  javahome list [--json]
  javahome current [--json]
  javahome print [version] [--vendor text] [--json]
  javahome use <version> [--vendor text] [--shell bash|zsh|fish|powershell|cmd]
  javahome use <version> --global [--shell bash|zsh|fish|powershell]
  javahome use <version> --project
  javahome doctor [--json]
  javahome init [bash|zsh|fish|powershell] [--global]
  javahome version

Notes:
  An external process cannot directly change the already-running parent shell.
  Use eval/source/Invoke-Expression for current-shell activation, or --global
  for profile updates.`)
}
