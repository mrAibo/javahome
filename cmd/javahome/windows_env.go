package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/mrAibo/javahome/internal/termui"
	"github.com/mrAibo/javahome/internal/winenv"
)

func cmdWindowsEnv(args []string) error {
	fs := flag.NewFlagSet("windows-env", flag.ContinueOnError)
	vendor := fs.String("vendor", "", "filter by vendor text")
	dryRun := fs.Bool("dry-run", false, "show planned changes without writing Windows environment")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"vendor": true})); err != nil {
		return err
	}
	if fs.NArg() < 2 {
		return fmt.Errorf("usage: javahome windows-env user|machine <major-version> [--vendor text] [--dry-run]")
	}
	if runtime.GOOS != "windows" {
		return fmt.Errorf("windows-env is only available on Windows; use `javahome use <version> --global --shell <shell>` on this platform")
	}

	scope, err := winenv.ParseScope(fs.Arg(0))
	if err != nil {
		return err
	}
	versionArg := fs.Arg(1)
	inst, err := selectInstallation(versionArg, *vendor)
	if err != nil {
		return err
	}

	ui := termui.New(os.Stdout)
	message, err := winenv.Apply(scope, inst.Path, *dryRun)
	if err != nil {
		if strings.TrimSpace(message) != "" {
			fmt.Print(message)
		}
		return err
	}
	if *dryRun {
		fmt.Println(ui.Warning(message))
		return nil
	}
	if strings.TrimSpace(message) != "" {
		fmt.Print(message)
	}
	fmt.Println(ui.Bullet("Current terminals do not inherit persisted Windows environment changes retroactively."))
	fmt.Println(ui.Bullet("Open a new terminal. Restart IDEs, daemons, or Windows services that were already running."))
	if scope == winenv.Machine {
		fmt.Println(ui.Bullet("Machine scope usually requires an elevated Administrator terminal."))
	}
	return nil
}
