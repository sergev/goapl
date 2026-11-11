package apl

// ANSI styles used when Interp.Color is set.
const (
	styleResult = "\x1b[32m"   // green
	styleError  = "\x1b[1;31m" // bold red
	styleWhere  = "\x1b[33m"   // yellow
	styleCaret  = "\x1b[31m"   // red
	stylePrompt = "\x1b[2m"    // dim
	styleReset  = "\x1b[0m"
)

// paint wraps s in an ANSI style when color is on.
func (in *Interp) paint(style, s string) string {
	if !in.Color || s == "" {
		return s
	}
	return style + s + styleReset
}
