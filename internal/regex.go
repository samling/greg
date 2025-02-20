package internal

import (
	"fmt"
	"regexp"
	"strings"
)

func (m *Model) parseRegexContent() string {
	// Debug information
	resultsContent := fmt.Sprintf(
		"Debug Info:\n"+
			"Input Height: %d\n"+
			"Full Height: %d\n"+
			"Full Width: %d\n"+
			"Input Width: %d\n\n"+
			"Active Cell: %d\n"+
			"Results lines: %d\n",
		m.inputHeight,
		m.fullHeight,
		m.fullWidth,
		m.inputs[0].Width,
		m.activeCell,
		strings.Count(m.resultsViewport.View(), "\n"),
	)

	if pattern := m.inputs[0].Value(); pattern != "" {
		resultsContent += metaStyle.Render("Pattern breakdown:") + "\n"
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
			resultsContent += "\n" + metaStyle.Render("Capture groups:") + "\n"
			for i := 0; i < groupCount; i++ {
				re, err := regexp.Compile(pattern)
				if err == nil {
					match := re.FindStringSubmatch(m.pipedContent)
					if len(match) > i+1 {
						resultsContent += groupStyle.Render(fmt.Sprintf("Group %d: %q\n", i+1, match[i+1]))
					}
				}
			}
		}

		resultsContent += "\n" + metaStyle.Render(fmt.Sprintf("Total matches: %d", len(m.matches)))
	} else {
		resultsContent = literalStyle.Render("Enter a pattern to see explanation")
	}

	return resultsContent
}
