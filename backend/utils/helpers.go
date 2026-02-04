package utils

import "strings"

func ExtractJSON(s string) string {

	s = strings.ReplaceAll(s, "`", "")

	start := strings.Index(s, "<json>")
	end := strings.Index(s, "</json>")

	if start == -1 || end == -1 {
		return ""
	}

	return strings.TrimSpace(s[start+6 : end])
}
