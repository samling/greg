package internal

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	// blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle = focusedStyle
	noStyle     = lipgloss.NewStyle()

	// focusedButton = focusedStyle.Render("[ Submit ]")
	// blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))

	// Layout styles
	columnStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1)

	resultsBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1)

	// Add new styles for pattern breakdown
	groupStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("87"))  // Cyan
	metaStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("213")) // Pink
	quantStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
	escapeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("148")) // Green
	literalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250")) // Light gray
)

type Model struct {
	focusedIndex int
	inputs       []textinput.Model
	cursorMode   cursor.Mode
	pipedContent string
	width        int
	height       int
	matches      [][]int
}

func InitialModel() Model {
	var pipedContent string
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		bytes, err := io.ReadAll(os.Stdin)
		if err == nil {
			pipedContent = string(bytes)
		}
	}

	m := Model{
		inputs:       make([]textinput.Model, 1),
		pipedContent: pipedContent,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

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
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Quit on enter press when submit is highlighted
			if s == "enter" && m.focusedIndex == len(m.inputs) {
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusedIndex--
			} else {
				m.focusedIndex++
			}

			if m.focusedIndex > len(m.inputs) {
				m.focusedIndex = 0
			} else if m.focusedIndex < 0 {
				m.focusedIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusedIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}

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

		return m, cmd
	}

	cmd := m.updateInputs(msg)
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
	// Get terminal dimensions
	columnWidth := m.width / 2

	// var buttonText string
	// if m.focusedIndex == len(m.inputs) {
	// 	buttonText = focusedButton
	// } else {
	// 	buttonText = blurredButton
	// }

	// Create input section (top-left)
	inputSection := lipgloss.JoinVertical(lipgloss.Left,
		m.inputs[0].View(),
		// buttonText,
	)
	inputSection = inputBoxStyle.Width(columnWidth - 4).Render(inputSection)

	var resultsContent string
	if pattern := m.inputs[0].Value(); pattern != "" {
		resultsContent = metaStyle.Render("Pattern breakdown:") + "\n"
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
			case '^', '$':
				resultsContent += metaStyle.Render("Start of line") + "\n"
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

	// Create the results section (bottom-left)
	resultsSection := resultsBoxStyle.
		Width(columnWidth - 4).
		Height(m.height - lipgloss.Height(inputSection) - 4).
		Render(resultsContent)

	// Create the left column
	leftColumn := lipgloss.JoinVertical(lipgloss.Left,
		inputSection,
		resultsSection,
	)
	leftColumn = columnStyle.Render(leftColumn)

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

	// Create the right column
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(columnWidth - 4)

	contentHeight := m.height - 6

	content := "Piped content:\n" + highlightedContent
	rightColumn := containerStyle.
		Height(m.height - 4).
		Render(lipgloss.NewStyle().
			Height(contentHeight).
			MaxHeight(contentHeight).
			Render(content))

	// Join columns horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
}
