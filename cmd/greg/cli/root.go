package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/samling/greg/internal/tui"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "greg",
		Run: func(cmd *cobra.Command, args []string) {
			p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			}
		},
	}

	return rootCmd
}
