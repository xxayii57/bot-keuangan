package sessions

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewSessionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"sess"},
		Short:   "Manage chat sessions",
		Long:    `List, view, and manage saved chat sessions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Sessions management for IntimClaw.")
			fmt.Println("")
			fmt.Println("Telegram commands:")
			fmt.Println("  /sessions list    — List all saved sessions")
			fmt.Println("  /sessions switch  — Switch to a different session")
			fmt.Println("  /sessions info    — Show current session details")
			fmt.Println("")
			fmt.Println("CLI usage:")
			fmt.Println("  intimclaw sessions list")
			return cmd.Help()
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all saved sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Session list is available via Telegram: /sessions list")
			fmt.Println("Use /sessions info to see current session.")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Show current session details",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Current session info is available via Telegram: /sessions info")
			return nil
		},
	})

	return cmd
}
