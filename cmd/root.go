package cmd

import (
	"fmt"
	"os"

	"github.com/9roads/ccyolo/internal/config"
	"github.com/9roads/ccyolo/internal/settings"
	"github.com/spf13/cobra"
)

var Version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "ccyolo",
	Short: "Smart permission filter for Claude Code",
	Long: `ccyolo - Claude Code YOLO Mode

Auto-approves safe operations using Claude API evaluation.
USE AT YOUR OWN RISK.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default to status
		showStatus()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(enableCmd)
	rootCmd.AddCommand(disableCmd)
	rootCmd.AddCommand(presetCmd)
	rootCmd.AddCommand(hookCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(checkCmd)
}

func showStatus() {
	cfg := config.Load()

	fmt.Printf("ccyolo %s\n\n", Version)

	// Check if hook is installed
	installed := settings.IsHookInstalled()
	if !installed {
		fmt.Println("Status:  NOT INSTALLED")
		fmt.Println("\nRun 'ccyolo install' to get started.")
		return
	}

	status := "DISABLED"
	if cfg.Enabled {
		status = "ENABLED"
	}

	fmt.Printf("Status:  %s\n", status)
	fmt.Printf("Preset:  %s\n", cfg.Preset)
	fmt.Printf("Model:   %s\n", cfg.Model)

	// Check API key
	hasKey := config.HasAPIKey()
	keyStatus := "NOT SET"
	if hasKey {
		keyStatus = "configured"
	}
	fmt.Printf("API Key: %s\n", keyStatus)

	fmt.Printf("Cache:   %ds TTL\n", cfg.CacheTTL)

	logStatus := "disabled"
	if cfg.Logging {
		logStatus = "enabled"
	}
	fmt.Printf("Logging: %s\n", logStatus)
}
