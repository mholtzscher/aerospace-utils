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
	//nolint:reassign // We intentionally modify the global color setting
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
//
//nolint:goprintffuncname // These are semantic output methods, not printf variants
func (p *Printer) Label(format string, a ...any) {
	_, _ = p.label.Printf(format, a...)
}

// Value prints a green value.
//
//nolint:goprintffuncname // These are semantic output methods, not printf variants
func (p *Printer) Value(format string, a ...any) {
	_, _ = p.value.Printf(format, a...)
}

// Path prints a dimmed path.
//
//nolint:goprintffuncname // These are semantic output methods, not printf variants
func (p *Printer) Path(format string, a ...any) {
	_, _ = p.path.Printf(format, a...)
}

// Unset prints a yellow "unset" indicator.
//
//nolint:goprintffuncname // These are semantic output methods, not printf variants
func (p *Printer) Unset(format string, a ...any) {
	_, _ = p.unset.Printf(format, a...)
}

// Success prints a green success message.
//
//nolint:goprintffuncname // These are semantic output methods, not printf variants
func (p *Printer) Success(format string, a ...any) {
	_, _ = p.success.Printf(format, a...)
}

// Warning prints a yellow warning message.
//
//nolint:goprintffuncname // These are semantic output methods, not printf variants
func (p *Printer) Warning(format string, a ...any) {
	_, _ = p.warning.Printf(format, a...)
}

// Error prints a red error message.
//
//nolint:goprintffuncname // These are semantic output methods, not printf variants
func (p *Printer) Error(format string, a ...any) {
	_, _ = p.err.Printf(format, a...)
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
func (p *Printer) PrintKeyValue(key string, value any) {
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
//
//nolint:forbidigo // This is the raw printf passthrough method
func (p *Printer) Printf(format string, a ...any) {
	fmt.Printf(format, a...)
}

// Println prints a blank line.
//
//nolint:forbidigo // Need to print blank line, no color needed
func (p *Printer) Println() {
	fmt.Println()
}
