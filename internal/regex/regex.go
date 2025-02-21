package regex

import (
	"fmt"
	"strings"
)

func ExplainRegexPattern(pattern string) string {
	if pattern == "" {
		return ""
	}

	var result strings.Builder
	inCaptureGroup := 0

	for i := 0; i < len(pattern); i++ {
		char := pattern[i]
		var styledChar, explanation string

		// Handle escape sequences
		if char == '\\' && i < len(pattern)-1 {
			nextChar := pattern[i+1]
			styledChar = fmt.Sprintf("[green]\\%c[white]", nextChar)

			// Expanded escape sequence explanations
			switch nextChar {
			case 'w':
				explanation = "Match any word character [a-zA-Z0-9_]"
			case 'W':
				explanation = "Match any non-word character"
			case 'd':
				explanation = "Match any digit [0-9]"
			case 'D':
				explanation = "Match any non-digit"
			case 's':
				explanation = "Match any whitespace character (space, tab, newline)"
			case 'S':
				explanation = "Match any non-whitespace character"
			case 'b':
				explanation = "Match a word boundary"
			case 'B':
				explanation = "Match a non-word boundary"
			case 'A':
				explanation = "Match start of string"
			case 'z':
				explanation = "Match end of string"
			case 't':
				explanation = "Match tab character"
			case 'n':
				explanation = "Match newline character"
			case 'r':
				explanation = "Match carriage return"
			default:
				explanation = fmt.Sprintf("Escape sequence for '%c'", nextChar)
			}
			i++ // Skip the next character
		} else if isMetaChar(char) {
			styledChar = fmt.Sprintf("[#ff69b4]%c[white]", char) // Pink for meta characters
			explanation = getCharacterExplanation(pattern, i, rune(char))
		} else if char == '(' || char == ')' {
			styledChar = fmt.Sprintf("[#5fd7ff]%c[white]", char) // Cyan for groups
			explanation = getCharacterExplanation(pattern, i, rune(char))
			// Update capture group nesting level
			if char == '(' && (i >= len(pattern)-2 || pattern[i+1] != '?' || pattern[i+2] != ':') {
				inCaptureGroup++
			} else if char == ')' {
				inCaptureGroup = max(0, inCaptureGroup-1)
			}
		} else if isQuantifier(char) {
			styledChar = fmt.Sprintf("[#ff8c00]%c[white]", char) // Orange for quantifiers
			explanation = getCharacterExplanation(pattern, i, rune(char))
		} else {
			styledChar = fmt.Sprintf("[#808080]%c[white]", char) // Gray for literals
			explanation = "Match this character literally"
		}

		line := fmt.Sprintf("%s: %s\n", styledChar, explanation)
		// Add indentation if we're inside a capture group (but not for the group markers themselves)
		if inCaptureGroup > 0 && char != '(' && char != ')' {
			line = "  " + line
		}
		result.WriteString(line)
	}
	return result.String()
}

// Helper function to get maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func isMetaChar(c byte) bool {
	return c == '^' || c == '$' || c == '.' || c == '[' || c == ']' || c == '|'
}

func isQuantifier(c byte) bool {
	return c == '*' || c == '+' || c == '?' || c == '{' || c == '}'
}

func getCharacterExplanation(pattern string, pos int, char rune) string {
	switch char {
	case '^':
		return "Match start of line"
	case '$':
		return "Match end of line"
	case '.':
		return "Match any character except newline"
	case '*':
		if pos > 0 {
			return "Match previous character zero or more times"
		}
		return "Match previous character zero or more times"
	case '+':
		if pos > 0 {
			return "Match previous character one or more times"
		}
		return "Match previous character one or more times"
	case '?':
		if pos > 0 {
			return "Match previous character zero or one time"
		}
		return "Match previous character zero or one time"
	case '[':
		return "Start character class"
	case ']':
		return "End character class"
	case '(':
		if pos < len(pattern)-2 && pattern[pos+1] == '?' && pattern[pos+2] == ':' {
			return "Start non-capturing group"
		}
		return "Start capturing group"
	case ')':
		return "End group"
	case '{':
		return "Start quantifier"
	case '}':
		return "End quantifier"
	case '|':
		return "Alternation (OR)"
	default:
		return "Match this character literally"
	}
}
