package cmd

import (
	"fmt"

	"github.com/9roads/ccyolo/internal/config"
	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable ccyolo auto-approval",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		cfg.Enabled = true
		if err := config.Save(cfg); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("ccyolo: ENABLED")
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable ccyolo auto-approval",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		cfg.Enabled = false
		if err := config.Save(cfg); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("ccyolo: DISABLED")
	},
}
