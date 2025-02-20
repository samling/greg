package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

type Model struct {
	// input fields
	inputs []textinput.Model

	// piped content
	pipedContent string

	// matches
	matches [][]int

	// capture groups
	hasCaptureGroups bool

	// active cell
	activeCell int

	// scroll offset
	scrollOffset int

	// terminal width
	width int

	// terminal height
	height int

	// input prompt height
	inputHeight int

	// default border width
	borderWidth int

	// results viewport
	resultsViewport viewport.Model

	// capture viewport
	captureViewport viewport.Model

	// groups content viewport
	contentViewport viewport.Model

	// result info
	footerText string
}
