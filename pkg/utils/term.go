package utils

import "github.com/muesli/termenv"

var (
	ColorProfile = termenv.ColorProfile()
)

func MakeFgStyle(color string) func(string) string {
	if !IsColorSupported() {
		return func(s string) string { return s }
	}
	return termenv.Style{}.Foreground(ColorProfile.Color(color)).Styled
}

// IsColorSupported returns true if the terminal supports colors
func IsColorSupported() bool {
	return !termenv.EnvNoColor() && ColorProfile != termenv.Ascii
}
