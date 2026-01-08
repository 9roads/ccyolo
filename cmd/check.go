package cmd

import (
	"fmt"

	"github.com/9roads/ccyolo/internal/claude"
	"github.com/9roads/ccyolo/internal/config"
	"github.com/9roads/ccyolo/internal/preset"
	"github.com/9roads/ccyolo/internal/settings"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check ccyolo setup and configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ccyolo Self-Check")
		fmt.Println("=================")
		fmt.Println()

		allGood := true

		// 1. Check hook registration
		fmt.Print("Hook registered:    ")
		if settings.IsHookInstalled() {
			fmt.Println("OK")
		} else {
			fmt.Println("MISSING")
			fmt.Println("  Run: ccyolo install")
			allGood = false
		}

		// 2. Check API key
		fmt.Print("API key:            ")
		apiKey := config.GetAPIKey()
		if apiKey != "" {
			fmt.Println("configured")

			// 3. Validate API key
			fmt.Print("API key valid:      ")
			if err := claude.ValidateAPIKey(apiKey); err != nil {
				fmt.Printf("FAILED (%v)\n", err)
				fmt.Println("  Run: ccyolo setup")
				allGood = false
			} else {
				fmt.Println("OK")
			}
		} else {
			fmt.Println("MISSING")
			fmt.Println("  Run: ccyolo setup")
			allGood = false
		}

		// 4. Check config
		cfg := config.Load()
		fmt.Print("Enabled:            ")
		if cfg.Enabled {
			fmt.Println("yes")
		} else {
			fmt.Println("no (run 'ccyolo enable' to enable)")
		}

		// 5. Check preset
		fmt.Printf("Preset:             %s\n", cfg.Preset)
		p := preset.Get(cfg.Preset)
		if p.Name == "" || p.Name == "balanced" && cfg.Preset != "balanced" {
			fmt.Println("  Warning: preset not found, using 'balanced'")
		}

		// 6. Check model
		fmt.Printf("Model:              %s\n", cfg.Model)

		// 7. Check logging
		fmt.Print("Logging:            ")
		if cfg.Logging {
			fmt.Println("enabled")
		} else {
			fmt.Println("disabled")
		}

		// 8. Check Claude Code settings path
		fmt.Printf("Claude settings:    %s\n", settings.ClaudeSettingsPath())

		fmt.Println()
		if allGood {
			fmt.Println("All checks passed!")
		} else {
			fmt.Println("Some checks failed. Please fix the issues above.")
		}
	},
}

func init() {
	// Will be added in root.go
}
