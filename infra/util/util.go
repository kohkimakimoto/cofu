package util

import "strings"

func ShellEscape(s string) string {
	return "'" + strings.Replace(s, "'", "'\"'\"'", -1) + "'"
}
