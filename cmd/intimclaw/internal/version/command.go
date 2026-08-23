package version

import (
	"github.com/spf13/cobra"

	"github.com/xxayii57/bot-keuangan/cmd/intimclaw/internal"
	"github.com/xxayii57/bot-keuangan/cmd/intimclaw/internal/cliui"
	"github.com/xxayii57/bot-keuangan/pkg/config"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Show version information",
		Run: func(_ *cobra.Command, _ []string) {
			printVersion()
		},
	}

	return cmd
}

func printVersion() {
	build, goVer := config.FormatBuildInfo()
	cliui.PrintVersion(internal.Logo, "intimclaw "+config.FormatVersion(), build, goVer)
}
