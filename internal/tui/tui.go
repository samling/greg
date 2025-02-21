package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/samling/greg/internal/regex"
)

var (
	pages      *tview.Pages
	focusViews []tview.Primitive
	leftSide   *tview.Flex

	matches          [][]int
	inputField       *tview.InputField
	resultsView      *tview.TextView
	contentView      *tview.TextView
	captureView      *tview.TextView
	captureViewAdded bool
)

func SetupTUI(app *tview.Application, rawContent string) error {
	captureViewAdded = false

	resultsView = newResultsView()
	contentView = newContentView(rawContent)
	captureView = newCaptureView()
	inputField = newInputField(rawContent)
	focusViews = []tview.Primitive{inputField, resultsView, contentView}

	// Create left side with results only initially
	leftSide = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(resultsView, 0, 1, false)

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
			cycleFocus(app, focusViews, false)
		case tcell.KeyBacktab:
			cycleFocus(app, focusViews, true)
		case tcell.KeyEsc:
			app.Stop()
		case tcell.KeyCtrlD:
			if contentView.HasFocus() {
				_, _, _, height := contentView.GetInnerRect()
				row, _ := contentView.GetScrollOffset()
				contentView.ScrollTo(row+height/2, 0)
				return nil
			}
		case tcell.KeyCtrlU:
			if contentView.HasFocus() {
				_, _, _, height := contentView.GetInnerRect()
				row, _ := contentView.GetScrollOffset()
				contentView.ScrollTo(max(0, row-height/2), 0)
				return nil
			}
		}
		return event
	})

	return nil
}

func newInputField(content string) *tview.InputField {
	inputField = tview.NewInputField().
		SetLabel("pattern: ").
		SetFieldWidth(30).
		SetChangedFunc(func(text string) {
			matches = nil

			if text == "" {
				contentView.SetText(content)
				resultsView.SetText("")
				captureView.SetText("")
				if captureViewAdded {
					leftSide.RemoveItem(captureView)
					captureViewAdded = false
					focusViews = []tview.Primitive{inputField, resultsView, contentView}
				}
				return
			}

			re, err := regexp.Compile(text)
			if err != nil {
				contentView.SetText(content)
				resultsView.SetText(fmt.Sprintf("[red]Invalid regex: %v[white]", err))
				captureView.SetText("")
				if captureViewAdded {
					leftSide.RemoveItem(captureView)
					captureViewAdded = false
					focusViews = []tview.Primitive{inputField, resultsView, contentView}
				}
				return
			}
			matches = re.FindAllStringIndex(content, -1)

			// Highlight matches
			highlightedContent := highlightMatches(content, matches)
			contentView.SetText(highlightedContent)

			// Show regex pattern explanation
			resultsView.SetText(regex.ExplainRegexPattern(text))

			// Handle capture groups
			hasCaptureGroups := strings.Count(text, "(") > strings.Count(text, "(?:")
			captureText := formatCaptureGroups(content, re)

			// Only show capture view if there are actual capture groups with content
			if hasCaptureGroups && captureText != "" {
				captureView.SetText(captureText)
				if !captureViewAdded {
					leftSide.AddItem(captureView, 0, 1, false)
					captureViewAdded = true
					focusViews = []tview.Primitive{inputField, resultsView, captureView, contentView}
				}
			} else if captureViewAdded {
				leftSide.RemoveItem(captureView)
				captureViewAdded = false
				focusViews = []tview.Primitive{inputField, resultsView, contentView}
			}
		})
	inputField.
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 0).
		SetTitle("input")

	return inputField
}

func newResultsView() *tview.TextView {
	resultsView = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetRegions(true)
	resultsView.SetBorder(true).
		SetBorderPadding(1, 1, 2, 2).
		SetTitle("pattern explanation")

	return resultsView
}

func newContentView(content string) *tview.TextView {
	contentView = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetRegions(true).
		SetText(content)
	contentView.SetBorder(true).
		SetBorderPadding(1, 1, 2, 2).
		SetTitle("content")

	return contentView
}

func newCaptureView() *tview.TextView {
	captureView = tview.NewTextView().
		SetWrap(false).
		SetDynamicColors(true)
	captureView.
		SetBorder(true).
		SetBorderPadding(1, 1, 2, 2).
		SetTitle("capture groups")

	return captureView
}

func cycleFocus(app *tview.Application, elements []tview.Primitive, reverse bool) {
	for i, el := range elements {
		if !el.HasFocus() {
			continue
		}

		if reverse {
			i = i - 1
			if i < 0 {
				i = len(elements) - 1
			}
		} else {
			i = i + 1
			i = i % len(elements)
		}

		app.SetFocus(elements[i])
		return
	}
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

func formatCaptureGroups(content string, re *regexp.Regexp) string {
	matches := re.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return ""
	}

	var result strings.Builder
	for groupNum := 1; groupNum < len(matches[0]); groupNum++ {
		result.WriteString(fmt.Sprintf("Group %d:\n", groupNum))

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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
