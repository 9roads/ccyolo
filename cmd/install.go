package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/9roads/ccyolo/internal/config"
	"github.com/9roads/ccyolo/internal/settings"
	"github.com/spf13/cobra"
)

func ccyoloDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ccyolo")
}

func createCCYoloDirs() error {
	dir := ccyoloDir()
	if err := os.MkdirAll(filepath.Join(dir, "presets"), 0755); err != nil {
		return err
	}
	return nil
}

func removeCCYoloDir() error {
	return os.RemoveAll(ccyoloDir())
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Register ccyolo hook with Claude Code",
	Run: func(cmd *cobra.Command, args []string) {
		// Create ~/.ccyolo directories
		if err := createCCYoloDirs(); err != nil {
			fmt.Printf("Warning: could not create config directory: %v\n", err)
		}

		// Find ccyolo binary path
		binaryPath, err := exec.LookPath("ccyolo")
		if err != nil {
			// Try to get the current executable
			binaryPath, err = os.Executable()
			if err != nil {
				fmt.Println("Error: ccyolo binary not found in PATH")
				fmt.Println("Make sure ccyolo is installed and in your PATH")
				return
			}
		}

		hookCmd := binaryPath + " hook"

		err = settings.AddHook(hookCmd)
		alreadyInstalled := err != nil && err.Error() == "ccyolo hook already installed"

		if err != nil && !alreadyInstalled {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if alreadyInstalled {
			fmt.Println("ccyolo hook already registered")
		} else {
			fmt.Println("ccyolo hook registered with Claude Code")
		}
		fmt.Println()

		// Auto-run setup if no API key configured
		if !config.HasAPIKey() {
			fmt.Println("No API key configured. Starting setup...")
			fmt.Println()
			RunSetup()
		} else {
			fmt.Println("API key already configured.")
			fmt.Println("Restart Claude Code for changes to take effect.")
		}
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove ccyolo hook from Claude Code",
	Run: func(cmd *cobra.Command, args []string) {
		if err := settings.RemoveHook(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("ccyolo hook removed from Claude Code")

		// Ask to remove API key from keychain
		if config.HasAPIKey() {
			fmt.Print("\nRemove API key from keychain? [Y/n]: ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer == "" || answer == "y" || answer == "yes" {
				if err := config.DeleteAPIKey(); err != nil {
					fmt.Printf("Warning: could not remove API key: %v\n", err)
				} else {
					fmt.Println("API key removed from keychain")
				}
			}
		}

		// Remove ~/.ccyolo directory
		if err := removeCCYoloDir(); err != nil {
			fmt.Printf("Warning: could not remove config directory: %v\n", err)
		} else {
			fmt.Println("Config directory removed")
		}

		fmt.Println("\nccyolo uninstalled successfully")
	},
}
