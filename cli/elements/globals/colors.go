package globals

import (
	"runtime"
	"strings"

	"github.com/mgutz/ansi"
)

var (
	// TitleStyle can be used to format a string; suitable for titles
	TitleStyle = ansi.ColorFunc("default+bh")

	// PrimaryTextStyle can be used to format a string; suitable for emphasis
	PrimaryTextStyle = ansi.ColorFunc("blue+b")

	// CTATextStyle can be used to format a string; important text
	CTATextStyle = ansi.ColorFunc("red+b:white+h") // Call To Action

	// ErrorTextStyle can be used to format a string; important text
	ErrorTextStyle = ansi.ColorFunc("red+b")

	// WarningStyle is used to format strings to indicate a warning
	WarningStyle = ansi.ColorFunc("yellow+bh")

	// SuccessStyle is used to format strings to indicate a successfull operation
	SuccessStyle = ansi.ColorFunc("green+bh")

	// BooleanTrueStyle is used to format strings that indicate a boolean true value
	BooleanTrueStyle = ansi.ColorFunc("blue")

	// BooleanFalseStyle is used to format strings that indicate a boolean false value
	BooleanFalseStyle = ansi.ColorFunc("red")
)

func init() {
	if runtime.GOOS == "windows" {
		TitleStyle = ansi.ColorFunc("default+b")
		PrimaryTextStyle = ansi.ColorFunc("default+bh")
		CTATextStyle = ansi.ColorFunc("red+bh:white+h")
		ErrorTextStyle = ansi.ColorFunc("red+bh:black")
		WarningStyle = ansi.ColorFunc("yellow+bh:black")
		BooleanTrueStyle = ansi.ColorFunc("cyan+h")
		BooleanFalseStyle = ansi.ColorFunc("red+h")
	}
}

// BooleanStyle colors instances of 'true' & 'false' blue & red respectively
func BooleanStyle(in string) string {
	replacer := strings.NewReplacer(
		"false", BooleanFalseStyle("false"),
		"true", BooleanTrueStyle("true"),
	)

	return replacer.Replace(in)
}

// BooleanStyleP is like BooleanStyle except takes a boolean as it's argument
func BooleanStyleP(in bool) string {
	if in {
		return BooleanStyle("true")
	}
	return BooleanStyle("false")
}
