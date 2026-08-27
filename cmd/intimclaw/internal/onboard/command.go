package onboard

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	intimclaw "github.com/xxayii57/intimclaw"
	"github.com/xxayii57/intimclaw/cmd/intimclaw/internal/model"
)

var embeddedFiles = intimclaw.OnboardWorkspace

func NewOnboardCommand() *cobra.Command {
	var encrypt bool

	cmd := &cobra.Command{
		Use:     "onboard",
		Aliases: []string{"o"},
		Short:   "Initialize intimclaw configuration and workspace",
		// Run without subcommands → original onboard flow
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				onboard(encrypt)
				if offerModelSetup(cmd.InOrStdin(), cmd.OutOrStdout()) {
					fmt.Fprintln(cmd.OutOrStdout())
					if err := model.RunSetupWizard(cmd.InOrStdin(), cmd.OutOrStdout()); err != nil {
						fmt.Fprintf(cmd.OutOrStdout(),
							"Model setup did not complete: %v\nYou can retry any time with:\n  intimclaw model setup\n", err)
					}
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "\nSkipped. Run 'intimclaw model setup' when you are ready.")
				}
			} else {
				_ = cmd.Help()
			}
		},
	}

	cmd.Flags().BoolVar(&encrypt, "enc", false,
		"Enable credential encryption (generates SSH key and prompts for passphrase)")

	return cmd
}

// offerModelSetup asks whether to launch the interactive model setup wizard
// right after onboarding. Defaults to yes; EOF (non-interactive) skips.
func offerModelSetup(stdin interface{ Read([]byte) (int, error) }, stdout interface{ Write([]byte) (int, error) }) bool {
	scanner := bufio.NewScanner(stdin)
	for {
		fmt.Fprint(stdout, "\nSet up your AI model now? [Y/n]: ")
		if !scanner.Scan() {
			return false
		}
		switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
		case "", "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Fprintln(stdout, "Please answer y or n.")
		}
	}
}
