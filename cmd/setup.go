package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/9roads/ccyolo/internal/claude"
	"github.com/9roads/ccyolo/internal/config"
	"github.com/spf13/cobra"
)

const apiKeysURL = "https://console.anthropic.com/settings/keys"

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup for ccyolo",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ccyolo Setup")
		fmt.Println("============")
		fmt.Println()

		// Check if already has API key
		if config.HasAPIKey() {
			fmt.Println("API key is already configured.")
			fmt.Print("Replace it? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Keeping existing API key.")
				return
			}
		}

		reader := bufio.NewReader(os.Stdin)
		var key string

		// Offer to open browser
		fmt.Printf("Open %s in browser? [Y/n]: ", apiKeysURL)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "n" && answer != "no" {
			if err := openBrowser(apiKeysURL); err != nil {
				fmt.Printf("Could not open browser: %v\n", err)
				fmt.Printf("Please visit: %s\n", apiKeysURL)
			}
		}
		fmt.Println()

		for {
			fmt.Println("Enter your Anthropic API key:")
			fmt.Print("> ")
			key, _ = reader.ReadString('\n')
			key = strings.TrimSpace(key)

			if key == "" {
				fmt.Println("No key provided. Aborting.")
				return
			}

			if !strings.HasPrefix(key, "sk-ant-") {
				fmt.Println("Warning: Key doesn't look like an Anthropic API key (should start with sk-ant-)")
				fmt.Print("Continue anyway? [y/N]: ")
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println()
					continue
				}
			}

			// Validate the key
			fmt.Print("Validating API key... ")
			if err := claude.ValidateAPIKey(key); err != nil {
				fmt.Println("FAILED")
				fmt.Printf("Error: %v\n\n", err)
				fmt.Print("Try again? [Y/n]: ")
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer == "n" || answer == "no" {
					fmt.Println("Aborting.")
					return
				}
				fmt.Println()
				continue
			}
			fmt.Println("OK")
			break
		}

		if err := config.SetAPIKey(key); err != nil {
			fmt.Printf("Error storing key in keychain: %v\n", err)
			fmt.Println("\nAlternative: set ANTHROPIC_API_KEY environment variable")
			return
		}

		fmt.Println("API key stored in system keychain.")
		fmt.Println()
		fmt.Println("Setup complete! Restart Claude Code for changes to take effect.")
	},
}

func RunSetup() {
	setupCmd.Run(nil, nil)
}
