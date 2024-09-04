package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

var (
	Version string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		//nolint:errcheck
		io.WriteString(cmd.OutOrStdout(), Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
