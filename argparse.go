package main

import "strings"

// parseArg parses restic arguments and splits them by spaces or quoted parts
func parseArg(arg string) []string {
	// res is the assembled result
	var res []string
	// sb is assembling the current argument
	sb := strings.Builder{}
	// isEscaped is true if the previous character was a backslash
	isEscaped := false
	// isQuoted is true if the current string is contained in double quotes
	isQuoted := false
	// iterate over each individual rune
	for _, x := range arg {
		if isEscaped {
			// last character was a backslash
			isEscaped = false
			if x == '"' {
				sb.WriteRune(x)
				continue
			} else if x == '\\' {
				sb.WriteRune(x)
				continue
			} else {
				// not a backslash, output the previously omitted backslash
				sb.WriteRune('\\')
			}
		}
		if x == '\\' {
			// indicate backslash and check upcoming rune to decide on proceeding
			isEscaped = true
			continue
		}
		if x == '"' {
			if isQuoted {
				// quoted string is closed
				isQuoted = false
				if sb.Len() > 0 {
					res = append(res, sb.String())
					sb.Reset()
				}
			} else {
				// start quoted string
				isQuoted = true
			}
			continue
		}
		if x == ' ' && !isQuoted {
			// space indentified as separator (not part of a quoted string)
			if sb.Len() > 0 {
				res = append(res, sb.String())
				sb.Reset()
			}
			continue
		}
		// other character, just append
		sb.WriteRune(x)
	}
	if sb.Len() > 0 {
		res = append(res, sb.String())
	}
	return res
}
