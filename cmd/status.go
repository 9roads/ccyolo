package cmd

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show ccyolo status",
	Run: func(cmd *cobra.Command, args []string) {
		showStatus()
	},
}
