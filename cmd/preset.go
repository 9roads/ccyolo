package cmd

import (
	"fmt"

	"github.com/9roads/ccyolo/internal/config"
	"github.com/9roads/ccyolo/internal/preset"
	"github.com/spf13/cobra"
)

var builtinPresets = []string{"strict", "balanced", "permissive"}

var presetCmd = &cobra.Command{
	Use:   "preset [name]",
	Short: "Set the safety preset",
	Long: `Set the safety preset for auto-approval decisions.

Built-in presets:
  strict      - Only auto-approve read operations
  balanced    - Auto-approve common dev tasks (default)
  permissive  - Auto-approve almost everything

Custom presets can be created in ~/.ccyolo/presets/`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()

		if len(args) == 0 {
			fmt.Printf("Current preset: %s\n\n", cfg.Preset)

			fmt.Println("Built-in presets:")
			for _, p := range builtinPresets {
				marker := "  "
				if p == cfg.Preset {
					marker = "* "
				}
				fmt.Printf("%s%s\n", marker, p)
			}

			// List custom presets
			custom, _ := preset.ListCustomPresets()
			if len(custom) > 0 {
				fmt.Println("\nCustom presets:")
				for _, p := range custom {
					marker := "  "
					if p == cfg.Preset {
						marker = "* "
					}
					fmt.Printf("%s%s\n", marker, p)
				}
			}
			return
		}

		presetName := args[0]

		// Check if valid (builtin or custom)
		valid := false
		for _, p := range builtinPresets {
			if p == presetName {
				valid = true
				break
			}
		}
		if !valid {
			custom, _ := preset.ListCustomPresets()
			for _, p := range custom {
				if p == presetName {
					valid = true
					break
				}
			}
		}

		if !valid {
			fmt.Printf("Invalid preset: %s\n", presetName)
			fmt.Println("Use 'ccyolo preset' to list available presets")
			fmt.Println("Use 'ccyolo preset create <name>' to create a custom preset")
			return
		}

		cfg.Preset = presetName
		if err := config.Save(cfg); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Preset set to: %s\n", presetName)
	},
}

var presetCreateCmd = &cobra.Command{
	Use:   "create <name> [base]",
	Short: "Create a custom preset",
	Long: `Create a custom preset based on a built-in preset.

Example:
  ccyolo preset create mypreset balanced

This creates ~/.ccyolo/presets/mypreset.json which you can edit.`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		base := "balanced"
		if len(args) > 1 {
			base = args[1]
		}

		// Check if already exists
		if _, err := preset.LoadCustomPreset(name); err == nil {
			fmt.Printf("Preset '%s' already exists\n", name)
			fmt.Printf("Edit: %s/%s.json\n", preset.CustomPresetsDir(), name)
			return
		}

		// Get base preset
		basePreset := preset.Get(base)
		basePreset.Name = name
		basePreset.Description = fmt.Sprintf("Custom preset based on %s", base)

		if err := preset.SaveCustomPreset(basePreset); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Created preset '%s' based on '%s'\n", name, base)
		fmt.Printf("Edit: %s/%s.json\n", preset.CustomPresetsDir(), name)
	},
}

var presetShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show preset details",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		name := cfg.Preset
		if len(args) > 0 {
			name = args[0]
		}

		p := preset.Get(name)
		fmt.Printf("Preset: %s\n", p.Name)
		fmt.Printf("Description: %s\n\n", p.Description)

		fmt.Println("Always Allow:")
		for _, r := range p.AlwaysAllow {
			fmt.Printf("  %s: %s\n", r.Tool, r.Pattern)
		}

		fmt.Println("\nAlways Deny:")
		for _, r := range p.AlwaysDeny {
			fmt.Printf("  %s: %s\n", r.Tool, r.Pattern)
		}
	},
}

func init() {
	presetCmd.AddCommand(presetCreateCmd)
	presetCmd.AddCommand(presetShowCmd)
}
