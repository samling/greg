package tui

import (
	"fmt"
	"regexp"
	"strings"
)

func (m *Model) parseRegexContent() string {
	// Debug information
	// resultsContent := fmt.Sprintf(
	// 	"Debug Info:\n"+
	// 		"Input Height: %d\n"+
	// 		"Full Height: %d\n"+
	// 		"Full Width: %d\n"+
	// 		"Input Width: %d\n\n"+
	// 		"Active Cell: %d\n"+
	// 		"Results lines: %d\n",
	// 	m.inputHeight,
	// 	m.height,
	// 	m.width,
	// 	m.inputs[0].Width,
	// 	m.activeCell,
	// 	strings.Count(m.resultsViewport.View(), "\n"),
	// )

	var resultsContent string

	if pattern := m.inputs[0].Value(); pattern != "" {
		resultsContent += metaStyle.Render("Pattern breakdown:") + "\n\n"
		groupLevel := 0
		groupCount := 0

		for i := 0; i < len(pattern); i++ {
			char := rune(pattern[i])
			indent := strings.Repeat("  ", groupLevel)
			prefix := fmt.Sprintf("%s%d: %c - ", indent, i+1, char)

			if char == '\\' && i < len(pattern)-1 {
				i++
				nextChar := pattern[i]
				resultsContent += fmt.Sprintf("%s%d: \\%c - ", indent, i, nextChar)
				desc := ""
				switch nextChar {
				case 'd':
					desc = "Digit (0-9)"
				case 'w':
					desc = "Word character"
				case 's':
					desc = "Whitespace"
				case 'b':
					desc = "Word boundary"
				default:
					desc = fmt.Sprintf("Escaped character '%c'", nextChar)
				}
				resultsContent += escapeStyle.Render(desc) + "\n"
				continue
			}

			resultsContent += prefix
			switch char {
			case '(':
				groupLevel++
				groupCount++
				resultsContent += groupStyle.Render(fmt.Sprintf("Start group %d ↓", groupCount)) + "\n"
			case ')':
				resultsContent = strings.TrimSuffix(resultsContent, prefix)
				groupLevel = max(0, groupLevel-1)
				indent = strings.Repeat("  ", groupLevel)
				resultsContent += fmt.Sprintf("%s%d: %c - ", indent, i+1, char) +
					groupStyle.Render(fmt.Sprintf("End group %d ↑", groupLevel+1)) + "\n"
			case '^':
				resultsContent += metaStyle.Render("Start of line") + "\n"
			case '$':
				resultsContent += metaStyle.Render("End of line") + "\n"
			case '.':
				resultsContent += metaStyle.Render("Any character") + "\n"
			case '*', '+', '?':
				resultsContent += quantStyle.Render("Zero or more of previous") + "\n"
			case '[':
				resultsContent += metaStyle.Render("Start character class") + "\n"
			case ']':
				resultsContent += metaStyle.Render("End character class") + "\n"
			case '{':
				resultsContent += quantStyle.Render("Start quantifier") + "\n"
			case '}':
				resultsContent += quantStyle.Render("End quantifier") + "\n"
			case '|':
				resultsContent += metaStyle.Render("Alternation (OR)") + "\n"
			default:
				resultsContent += literalStyle.Render("Literal character") + "\n"
			}
		}

		if groupCount > 0 {
			m.hasCaptureGroups = true
		} else {
			m.hasCaptureGroups = false
		}

	} else {
		resultsContent = literalStyle.Render("Enter a pattern to see explanation")
	}

	return resultsContent
}

func (m *Model) parseCaptureGroups(pattern string) string {
	var captureContent strings.Builder
	captureContent.WriteString(metaStyle.Render("Capture groups:") + "\n\n")

	re, err := regexp.Compile(pattern)
	if err == nil {
		match := re.FindStringSubmatch(m.pipedContent)
		for i := 1; i < len(match); i++ {
			groupNum := fmt.Sprintf("%d", i)
			captureContent.WriteString(groupStyle.Render(fmt.Sprintf("Group %-1s: ", groupNum)))
			captureContent.WriteString(literalStyle.Render(fmt.Sprintf("%q", match[i])) + "\n")
		}
	}

	return captureContent.String()
}
