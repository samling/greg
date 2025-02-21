package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/rivo/tview"

	_ "github.com/samling/greg/internal/style"
	"github.com/samling/greg/internal/tui"
	"github.com/spf13/cobra"
)

var (
	app *tview.Application
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

			// Initialize the application
			app = tview.NewApplication()

			// Create the TUI
			err := tui.SetupTUI(app, pipedContent)
			if err != nil {
				fmt.Println(err)
			}

			// Run the application
			err = app.Run()
			if err != nil {
				fmt.Println(err)
			}
		},
	}

	rootCmd.Flags().StringVarP(&filename, "file", "f", "", "file to read from")

	return rootCmd
}
