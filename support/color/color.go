package color

import (
	"fmt"
	"github.com/fatih/color"
)

var FgBold = color.New(color.Bold).SprintfFunc()

var FgG = color.New(color.FgGreen).SprintfFunc()

var FgGB = color.New(color.FgGreen).Add(color.Bold).SprintfFunc()

var FgY = color.New(color.FgYellow).SprintfFunc()

var FgYB = color.New(color.FgYellow).Add(color.Bold).SprintfFunc()

var FgM = color.New(color.FgMagenta).SprintfFunc()

var FgMB = color.New(color.FgMagenta).Add(color.Bold).SprintfFunc()

var FgC = color.New(color.FgCyan).SprintfFunc()

var FgCB = color.New(color.FgCyan).Add(color.Bold).SprintfFunc()

var FgR = color.New(color.FgRed).SprintfFunc()

var FgRB = color.New(color.FgRed).Add(color.Bold).SprintfFunc()

func PrintFgBold(format string, a ...interface{}) (n int, err error) {
	// this is an example implementation for test coverage.
	return fmt.Printf(FgBold(format), a...)
}
