package tui

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// General styles
	defaultStyle           = lipgloss.NewStyle()
	defaultWithBorderStyle = lipgloss.NewStyle().Inherit(defaultStyle).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240"))

	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()

	// Regex colors
	groupStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("87"))  // Cyan
	metaStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("213")) // Pink
	quantStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
	escapeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("148")) // Green
	literalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250")) // Light gray
)

func InitialModel() Model {
	var pipedContent string
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		bytes, err := io.ReadAll(os.Stdin)
		if err == nil {
			pipedContent = string(bytes)
		}
	}

	contentViewport := viewport.New(0, 0)
	contentViewport.SetContent(pipedContent)

	m := Model{
		inputs:           make([]textinput.Model, 1),
		pipedContent:     pipedContent,
		activeCell:       0,
		scrollOffset:     0,
		hasCaptureGroups: false,
		resultsViewport:  viewport.New(0, 0),
		contentViewport:  contentViewport,
		captureViewport:  viewport.New(0, 0),
		inputHeight:      1,
		borderWidth:      1,
		width:            0,
		height:           0,
		footerText:       "",
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 256 // todo: what should this actually be?
		t.Placeholder = "Pattern"
		t.Focus()
		t.PromptStyle = focusedStyle
		t.TextStyle = focusedStyle

		m.inputs[i] = t
	}

	m.footerText = m.buildFooterText()

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("greg"),
		textinput.Blink,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Cycle between cells
		case "tab", "shift+tab":
			m.activeCell = (m.activeCell + 1) % 4
			for i := range m.inputs {
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}
			if m.activeCell < len(m.inputs) {
				m.inputs[m.activeCell].Focus()
				m.inputs[m.activeCell].PromptStyle = focusedStyle
				m.inputs[m.activeCell].TextStyle = focusedStyle
			}
			return m, nil

		case "up", "k":
			if m.activeCell == 1 {
				m.contentViewport.LineUp(1)
				return m, nil
			} else if m.activeCell == 2 {
				m.resultsViewport.LineUp(1)
				return m, nil
			}

		case "ctrl+u":
			if m.activeCell == 1 {
				m.contentViewport.LineUp(8)
				return m, nil
			} else if m.activeCell == 2 {
				m.resultsViewport.LineUp(8)
				return m, nil
			}

		case "down", "j":
			if m.activeCell == 1 {
				m.contentViewport.LineDown(1)
				return m, nil
			} else if m.activeCell == 2 {
				m.resultsViewport.LineDown(1)
				return m, nil
			}

		case "ctrl+d":
			if m.activeCell == 1 {
				m.contentViewport.LineDown(8)
				return m, nil
			} else if m.activeCell == 2 {
				m.resultsViewport.LineDown(8)
				return m, nil
			}
		}

		if m.activeCell == 0 {
			cmd := m.updateInputs(msg)
			pattern := m.inputs[0].Value()
			if pattern != "" {
				if re, err := regexp.Compile(pattern); err == nil {
					m.matches = re.FindAllStringIndex(m.pipedContent, -1)
					m.hasCaptureGroups = strings.Count(pattern, "(") > strings.Count(pattern, "(?:")
					if m.hasCaptureGroups {
						m.captureViewport.SetContent(m.parseCaptureGroups(pattern))
					}
				} else {
					m.matches = nil
					m.hasCaptureGroups = false
					m.footerText = fmt.Sprintf("Invalid pattern: %s", err.Error())
					return m, cmd
				}
			} else {
				m.matches = nil
				m.hasCaptureGroups = false
			}

			m.resultsViewport.SetContent(m.parseRegexContent())
			m.footerText = m.buildFooterText()
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.resultsViewport, cmd = m.resultsViewport.Update(msg)
	if m.hasCaptureGroups {
		m.captureViewport, _ = m.captureViewport.Update(msg)
	}
	return m, cmd
}

func (m Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's
	// safe to simply update all of them here.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	// Render header
	header := m.headerRender()

	// Render primary content
	content := m.contentRender()

	// Render footer
	footer := m.footerRender()

	return lipgloss.JoinVertical(lipgloss.Top, header, content, footer)
}

// Render the pattern input box
func (m Model) headerRender() string {
	columnWidth := m.width
	rowHeight := 1

	inputBoxStyle := lipgloss.NewStyle().
		BorderForeground(lipgloss.Color("240")).
		Align(lipgloss.Left).
		Height(rowHeight).
		Width(columnWidth)

	if m.activeCell == 0 {
		inputBoxStyle = inputBoxStyle.BorderForeground(lipgloss.Color("205"))
	} else {
		inputBoxStyle = inputBoxStyle.BorderForeground(lipgloss.Color("240"))
	}

	return inputBoxStyle.Render(m.inputs[0].View())
}

// Render the primary content between the header and footer
func (m Model) contentRender() string {
	rowHeight := m.height - lipgloss.Height(m.headerRender()) - lipgloss.Height(m.footerRender())

	content := lipgloss.JoinHorizontal(0, m.resultsRender(), m.pipedContentRender())

	return defaultStyle.Height(rowHeight).Render(content)
}

// Render the pattern evaluation results viewport
func (m Model) resultsRender() string {
	columnWidth := m.width / 2
	rowHeight := m.height - lipgloss.Height(m.headerRender()) - lipgloss.Height(m.footerRender()) - defaultWithBorderStyle.GetVerticalBorderSize()

	if !m.hasCaptureGroups {
		m.resultsViewport.Width = columnWidth
		m.resultsViewport.Height = rowHeight

		resultsBoxStyle := defaultWithBorderStyle.Width(columnWidth)

		if m.activeCell == 2 {
			resultsBoxStyle = resultsBoxStyle.BorderForeground(lipgloss.Color("205"))
		}

		return resultsBoxStyle.Render(m.resultsViewport.View())
	}

	// Split view for capture groups
	// Account for the shared border between panes (subtract 1)
	halfWidth := columnWidth / 2
	m.resultsViewport.Width = halfWidth
	m.resultsViewport.Height = rowHeight
	m.captureViewport.Width = halfWidth
	m.captureViewport.Height = rowHeight

	resultsBoxStyle := defaultWithBorderStyle.Width(halfWidth - 1) // Subtract 1 for the shared border

	if m.activeCell == 2 {
		resultsBoxStyle = resultsBoxStyle.BorderForeground(lipgloss.Color("205"))
	}

	captureBoxStyle := defaultWithBorderStyle.Width(halfWidth)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		resultsBoxStyle.Render(m.resultsViewport.View()),
		captureBoxStyle.Render(m.captureViewport.View()),
	)
}

// Render the content piped into the application
func (m Model) pipedContentRender() string {
	columnWidth := m.width / 2
	rowHeight := m.height - lipgloss.Height(m.headerRender()) - lipgloss.Height(m.footerRender()) - defaultWithBorderStyle.GetVerticalBorderSize()

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
	m.contentViewport.Height = rowHeight
	m.contentViewport.SetContent(highlightedContent)

	// Create the right column with additional top padding
	rightColumnStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Align(lipgloss.Left).
		Width(columnWidth)

	if m.activeCell == 1 {
		rightColumnStyle = rightColumnStyle.BorderForeground(lipgloss.Color("205"))
	}

	return rightColumnStyle.Render(m.contentViewport.View())
}

// Render the footer
func (m Model) footerRender() string {
	footer := defaultStyle.
		Align(lipgloss.Left).
		Width(m.width).
		Render(m.footerText)
	return footer
}

func (m Model) buildFooterText() string {
	var info []string

	matchCount := 0
	if m.matches != nil {
		matchCount = len(m.matches)
	}
	info = append(info, fmt.Sprintf("Total matches: %d", matchCount))

	return strings.Join(info, " ")
}
