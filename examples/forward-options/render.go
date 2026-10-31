package render

import "strings"

// Options controls rendering.
type Options struct {
	Verbose bool `json:"verbose"`
	Limit   int  `json:"limit"`
}

// Render joins items and optionally labels the output.
func Render(items []string, verbose bool) string {
	output := strings.Join(items, ", ")
	if verbose {
		return "items: " + output
	}
	return output
}
