package cli

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/spf13/cobra"
)

var (
	app         *tview.Application
	pages       *tview.Pages
	finderFocus tview.Primitive
	leftSide    *tview.Flex

	originalContent string

	matches          [][]int
	inputField       *tview.InputField
	resultsView      *tview.TextView
	contentView      *tview.TextView
	captureView      *tview.TextView
	captureViewAdded bool
)

func NewRootCommand() *cobra.Command {
	var filename string

	rootCmd := &cobra.Command{
		Use: "greg",
		Run: func(cmd *cobra.Command, args []string) {
			var pipedContent string

			if filename != "" {
				bytes, err := os.ReadFile(filename)
				if err == nil {
					pipedContent = string(bytes)
				}
			} else {
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) == 0 {
					bytes, err := io.ReadAll(os.Stdin)
					if err == nil {
						pipedContent = string(bytes)
					}
				}
			}

			app = tview.NewApplication()
			content(pipedContent)
			app.Run()
		},
	}

	rootCmd.Flags().StringVarP(&filename, "file", "f", "", "file to read from")

	return rootCmd
}

func content(initialContent string) {
	originalContent = initialContent
	captureViewAdded = false

	resultsView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)
	resultsView.SetBorder(true).SetTitle("pattern explanation")

	contentView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetText(initialContent)
	contentView.SetBorder(true).SetTitle("content")

	captureView = tview.NewTextView().
		SetDynamicColors(true)
	captureView.SetBorder(true).SetTitle("capture")

	// Create left side with results only initially
	leftSide = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(resultsView, 0, 1, false)

	inputField = tview.NewInputField().
		SetLabel("Enter a pattern ").
		SetFieldWidth(30).
		SetChangedFunc(func(text string) {
			matches = nil

			if text == "" {
				contentView.SetText(originalContent)
				resultsView.SetText("")
				captureView.SetText("")
				if captureViewAdded {
					leftSide.RemoveItem(captureView)
					captureViewAdded = false
				}
				return
			}

			re, err := regexp.Compile(text)
			if err != nil {
				contentView.SetText(originalContent)
				resultsView.SetText(fmt.Sprintf("[red]Invalid regex: %v[white]", err))
				captureView.SetText("")
				if captureViewAdded {
					leftSide.RemoveItem(captureView)
					captureViewAdded = false
				}
				return
			}
			matches = re.FindAllStringIndex(originalContent, -1)

			// Highlight matches
			highlightedContent := highlightMatches(originalContent, matches)
			contentView.SetText(highlightedContent)

			// Show regex pattern explanation
			resultsView.SetText(explainRegexPattern(text))

			// Handle capture groups
			hasCaptureGroups := strings.Count(text, "(") > strings.Count(text, "(?:")
			captureText := formatCaptureGroups(text, originalContent, re)

			// Only show capture view if there are actual capture groups with content
			if hasCaptureGroups && captureText != "" {
				captureView.SetText(captureText)
				if !captureViewAdded {
					leftSide.AddItem(captureView, 0, 1, false)
					captureViewAdded = true
				}
			} else if captureViewAdded {
				leftSide.RemoveItem(captureView)
				captureViewAdded = false
			}
		})
	inputField.SetBorder(true).SetTitle("input")

	// Create main content area
	contentFlex := tview.NewFlex().
		AddItem(leftSide, 0, 1, false).
		AddItem(contentView, 0, 1, false)

	// Create main layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(inputField, 3, 0, true).
		AddItem(contentFlex, 0, 1, false)

	pages = tview.NewPages().AddPage("main", flex, true, true)
	app.SetRoot(pages, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			// Implement cycling logic
		case tcell.KeyEsc:
			app.Stop()
		}
		return event
	})
}

func highlightMatches(content string, matches [][]int) string {
	if len(matches) == 0 {
		return content
	}

	var result strings.Builder
	lastIdx := 0

	for _, match := range matches {
		result.WriteString(content[lastIdx:match[0]])
		result.WriteString("[red]")
		result.WriteString(content[match[0]:match[1]])
		result.WriteString("[reset]")
		lastIdx = match[1]
	}
	result.WriteString(content[lastIdx:])
	return result.String()
}

func formatMatches(content string, matches [][]int) string {
	var result strings.Builder
	for _, match := range matches {
		result.WriteString(content[match[0]:match[1]])
		result.WriteString("\n")
	}
	return result.String()
}

func formatCaptureGroups(pattern string, content string, re *regexp.Regexp) string {
	matches := re.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return ""
	}

	var result strings.Builder
	for groupNum := 1; groupNum < len(matches[0]); groupNum++ {
		result.WriteString(fmt.Sprintf("Group %d: ", groupNum))

		// Collect unique captures for this group
		var captures []string
		for _, match := range matches {
			if !contains(captures, match[groupNum]) {
				captures = append(captures, match[groupNum])
			}
		}

		// Print the unique captures with indentation
		for _, capture := range captures {
			result.WriteString(fmt.Sprintf("\t%s\n", capture))
		}
	}
	return result.String()
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func explainRegexPattern(pattern string) string {
	if pattern == "" {
		return ""
	}

	var result strings.Builder
	inCaptureGroup := 0 // Track nesting level of capture groups

	for i := 0; i < len(pattern); i++ {
		char := pattern[i]
		var styledChar, explanation string

		// Handle escape sequences
		if char == '\\' && i < len(pattern)-1 {
			styledChar = fmt.Sprintf("[green]\\%c[white]", pattern[i+1])
			explanation = fmt.Sprintf("Escape sequence for '%c'", pattern[i+1])
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
			return fmt.Sprintf("Match previous character zero or more times")
		}
		return "Match previous character zero or more times"
	case '+':
		if pos > 0 {
			return fmt.Sprintf("Match previous character one or more times")
		}
		return "Match previous character one or more times"
	case '?':
		if pos > 0 {
			return fmt.Sprintf("Match previous character zero or one time")
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
