package termui

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

type UI struct {
	Enabled bool
}

func New(file *os.File) UI {
	return UI{Enabled: colorEnabled(file)}
}

func (ui UI) Paint(code string, value string) string {
	if !ui.Enabled || value == "" {
		return value
	}
	return "\x1b[" + code + "m" + value + "\x1b[0m"
}

func (ui UI) Bold(value string) string    { return ui.Paint("1", value) }
func (ui UI) Dim(value string) string     { return ui.Paint("2", value) }
func (ui UI) Red(value string) string     { return ui.Paint("31", value) }
func (ui UI) Green(value string) string   { return ui.Paint("32", value) }
func (ui UI) Yellow(value string) string  { return ui.Paint("33", value) }
func (ui UI) Blue(value string) string    { return ui.Paint("34", value) }
func (ui UI) Magenta(value string) string { return ui.Paint("35", value) }
func (ui UI) Cyan(value string) string    { return ui.Paint("36", value) }
func (ui UI) Gray(value string) string    { return ui.Paint("90", value) }

func (ui UI) Success(value string) string { return ui.Green(value) }
func (ui UI) Warning(value string) string { return ui.Yellow(value) }
func (ui UI) Error(value string) string   { return ui.Red(value) }
func (ui UI) Command(value string) string { return ui.Cyan(value) }
func (ui UI) Path(value string) string    { return ui.Blue(value) }
func (ui UI) Key(value string) string     { return ui.Bold(value) }

func (ui UI) Check(ok bool) string {
	if ok {
		return ui.Green("OK")
	}
	return ui.Yellow("WARN")
}

func (ui UI) Bullet(value string) string {
	return fmt.Sprintf("%s %s", ui.Cyan("›"), value)
}

func colorEnabled(file *os.File) bool {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("JAVAHOME_COLOR")))
	if mode == "always" || os.Getenv("CLICOLOR_FORCE") == "1" {
		return true
	}
	if mode == "never" || os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	if runtime.GOOS == "windows" && os.Getenv("ANSICON") == "" && os.Getenv("WT_SESSION") == "" && os.Getenv("TERM_PROGRAM") == "" {
		// Modern Windows terminals support ANSI escape sequences. Older console hosts may not.
		// Keep auto mode conservative unless a known ANSI-capable host is detected.
		return false
	}
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}
