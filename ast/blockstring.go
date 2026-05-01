package ast

import "strings"

// BlockStringValue applies the GraphQL spec's BlockStringValue algorithm to
// the raw text between """ delimiters: it removes the common leading
// whitespace from all lines except the first, trims leading and trailing
// blank lines, and joins the result with '\n'. Line terminators (\n, \r, \r\n)
// are normalized to '\n' in the output.
func BlockStringValue(raw string) string {
	lines := splitBlockLines(raw)

	// Find the common indent across all lines except the first.
	commonIndent := -1
	for i := 1; i < len(lines); i++ {
		indent := leadingWhitespaceLen(lines[i])
		if indent < len(lines[i]) { // non-blank line
			if commonIndent == -1 || indent < commonIndent {
				commonIndent = indent
				if commonIndent == 0 {
					break
				}
			}
		}
	}

	// Strip common indent from non-first lines.
	if commonIndent > 0 {
		for i := 1; i < len(lines); i++ {
			line := lines[i]
			if commonIndent < len(line) {
				lines[i] = line[commonIndent:]
			} else {
				lines[i] = ""
			}
		}
	}

	// Trim leading and trailing blank lines.
	for len(lines) > 0 && isBlankLine(lines[0]) {
		lines = lines[1:]
	}
	for len(lines) > 0 && isBlankLine(lines[len(lines)-1]) {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}

// splitBlockLines splits s on any of \n, \r, \r\n.
func splitBlockLines(s string) []string {
	var out []string
	start, i := 0, 0
	for i < len(s) {
		switch s[i] {
		case '\n':
			out = append(out, s[start:i])
			i++
			start = i
		case '\r':
			out = append(out, s[start:i])
			i++
			if i < len(s) && s[i] == '\n' {
				i++
			}
			start = i
		default:
			i++
		}
	}
	out = append(out, s[start:])
	return out
}

// leadingWhitespaceLen returns the byte length of the leading run of spaces
// and tabs in line.
func leadingWhitespaceLen(line string) int {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	return i
}

// isBlankLine reports whether line consists entirely of spaces and tabs.
func isBlankLine(line string) bool {
	return leadingWhitespaceLen(line) == len(line)
}
