package internal

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	borderWidth = 2
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()

	// Layout styles
	columnStyle = lipgloss.NewStyle()

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	resultsBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	// Add new styles for pattern breakdown
	groupStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("87"))  // Cyan
	metaStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("213")) // Pink
	quantStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
	escapeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("148")) // Green
	literalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250")) // Light gray
)

func (m Model) inputRender() string {
	columnWidth := m.fullWidth/2 - borderWidth

	inputSection := lipgloss.JoinVertical(lipgloss.Left,
		m.inputs[0].View(),
	)

	// Highlight input box when active
	currentInputBoxStyle := inputBoxStyle.
		Width(columnWidth).
		Height(m.inputHeight)

	if m.activeCell == 0 {
		currentInputBoxStyle = currentInputBoxStyle.BorderForeground(lipgloss.Color("205"))
	} else {
		currentInputBoxStyle = currentInputBoxStyle.BorderForeground(lipgloss.Color("240"))
	}

	return currentInputBoxStyle.Render(inputSection)
}

func (m Model) resultsRender() string {
	columnWidth := m.fullWidth/2 - borderWidth
	resultsHeight := m.fullHeight - (lipgloss.Height(m.inputRender()) + borderWidth)

	m.resultsViewport.Width = columnWidth
	m.resultsViewport.Height = resultsHeight

	// Create the results section with conditional highlighting
	currentResultsBoxStyle := resultsBoxStyle.
		Width(columnWidth).
		Height(resultsHeight)
	if m.activeCell == 2 {
		currentResultsBoxStyle = currentResultsBoxStyle.BorderForeground(lipgloss.Color("205"))
	} else {
		currentResultsBoxStyle = currentResultsBoxStyle.BorderForeground(lipgloss.Color("240"))
	}

	return currentResultsBoxStyle.Render(m.resultsViewport.View())
}

func (m Model) contentRender() string {
	columnWidth := m.fullWidth/2 - borderWidth
	contentHeight := m.fullHeight - borderWidth

	// Highlight matches in piped content
	var highlightedContent string
	if len(m.matches) > 0 {
		lastIdx := 0
		for _, match := range m.matches {
			// Add non-matching text
			highlightedContent += m.pipedContent[lastIdx:match[0]]
			// Add highlighted matching text
			matchText := m.pipedContent[match[0]:match[1]]
			highlightedContent += focusedStyle.Render(matchText)
			lastIdx = match[1]
		}
		// Add remaining text
		highlightedContent += m.pipedContent[lastIdx:]
	} else {
		highlightedContent = m.pipedContent
	}

	// Set content for the viewport
	m.contentViewport.Width = columnWidth
	m.contentViewport.Height = contentHeight
	m.contentViewport.SetContent(highlightedContent)

	// Create the right column with additional top padding
	rightColumnStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(columnWidth).
		Height(contentHeight)

	if m.activeCell == 1 {
		rightColumnStyle = rightColumnStyle.BorderForeground(lipgloss.Color("205"))
	}

	rightColumn := rightColumnStyle.Render(m.contentViewport.View())

	return rightColumn
}
