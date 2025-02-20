package internal

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

type Model struct {
	inputs          []textinput.Model
	pipedContent    string
	matches         [][]int
	activeCell      int
	scrollOffset    int
	fullWidth       int
	fullHeight      int
	inputHeight     int
	resultsViewport viewport.Model
	contentViewport viewport.Model
}
