package internal

import (
	"io"
	"os"
	"regexp"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		inputs:          make([]textinput.Model, 1),
		pipedContent:    pipedContent,
		activeCell:      0,
		scrollOffset:    0,
		resultsViewport: viewport.New(0, 0),
		contentViewport: contentViewport,
		inputHeight:     1,
		fullWidth:       0,
		fullHeight:      0,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 256 // todo: what should this actually be?

		switch i {
		case 0:
			t.Placeholder = "Pattern"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		}

		m.inputs[i] = t
	}

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
		m.fullWidth = msg.Width
		m.fullHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Cycle between cells
		case "tab", "shift+tab":
			m.activeCell = (m.activeCell + 1) % 3
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
				} else {
					m.matches = nil
				}
			} else {
				m.matches = nil
			}

			m.resultsViewport.SetContent(m.parseRegexContent())
			return m, cmd
		}
	}

	// cmd := m.updateInputs(msg)
	var cmd tea.Cmd
	m.resultsViewport, cmd = m.resultsViewport.Update(msg)
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
	// Render the left column sections
	inputSection := m.inputRender()
	resultsSection := m.resultsRender()

	// Join the left column sections vertically
	leftColumn := lipgloss.JoinVertical(lipgloss.Left,
		inputSection,
		resultsSection,
	)
	leftColumn = columnStyle.Render(leftColumn)

	// Render the right column section
	rightColumn := m.contentRender()

	// Join columns horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top,
		leftColumn,
		rightColumn,
	)
}
