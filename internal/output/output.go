// Package output provides colored terminal output helpers.
package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Printer handles formatted console output with optional colors.
type Printer struct {
	label   *color.Color
	value   *color.Color
	path    *color.Color
	unset   *color.Color
	success *color.Color
	warning *color.Color
	err     *color.Color
}

// New creates a new Printer with color settings.
func New(noColor bool) *Printer {
	// Respect NO_COLOR environment variable and --no-color flag
	if noColor || os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}

	return &Printer{
		label:   color.New(color.FgCyan),
		value:   color.New(color.FgGreen),
		path:    color.New(color.Faint),
		unset:   color.New(color.FgYellow),
		success: color.New(color.FgGreen),
		warning: color.New(color.FgYellow),
		err:     color.New(color.FgRed),
	}
}

// Label prints a cyan label.
func (p *Printer) Label(format string, a ...interface{}) {
	p.label.Printf(format, a...)
}

// Value prints a green value.
func (p *Printer) Value(format string, a ...interface{}) {
	p.value.Printf(format, a...)
}

// Path prints a dimmed path.
func (p *Printer) Path(format string, a ...interface{}) {
	p.path.Printf(format, a...)
}

// Unset prints a yellow "unset" indicator.
func (p *Printer) Unset(format string, a ...interface{}) {
	p.unset.Printf(format, a...)
}

// Success prints a green success message.
func (p *Printer) Success(format string, a ...interface{}) {
	p.success.Printf(format, a...)
}

// Warning prints a yellow warning message.
func (p *Printer) Warning(format string, a ...interface{}) {
	p.warning.Printf(format, a...)
}

// Error prints a red error message.
func (p *Printer) Error(format string, a ...interface{}) {
	p.err.Printf(format, a...)
}

// ReloadOK prints a success message for config reload.
func (p *Printer) ReloadOK() {
	p.Success("Reloaded aerospace config\n")
}

// ReloadSkipped prints a warning that reload was skipped.
func (p *Printer) ReloadSkipped() {
	p.Warning("Skipped config reload (--no-reload)\n")
}

// ReloadFailed prints an error for failed config reload.
func (p *Printer) ReloadFailed(err error) {
	p.Error("Failed to reload config: %v\n", err)
}

// DryRun prints a notice that this is a dry run.
func (p *Printer) DryRun() {
	p.Warning("[dry-run] ")
}

// PrintKeyValue prints a key-value pair with formatting.
func (p *Printer) PrintKeyValue(key string, value interface{}) {
	p.Label("  %s: ", key)
	if value == nil {
		p.Unset("(not set)\n")
		return
	}
	p.Value("%v\n", value)
}

// PrintHeader prints a section header.
func (p *Printer) PrintHeader(title string) {
	p.Label("%s\n", title)
}

// PrintPath prints a path with its label.
func (p *Printer) PrintPath(label, path string) {
	p.Label("  %s: ", label)
	p.Path("%s\n", path)
}

// Printf prints formatted output without color.
func (p *Printer) Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}
