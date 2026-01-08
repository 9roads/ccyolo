package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/9roads/ccyolo/internal/config"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Manage logging",
}

var logEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable logging",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		cfg.Logging = true
		if err := config.Save(cfg); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("Logging enabled")
		fmt.Printf("Log file: %s\n", logFilePath())
	},
}

var logDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable logging",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		cfg.Logging = false
		if err := config.Save(cfg); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("Logging disabled")
	},
}

var logShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show log file contents",
	Run: func(cmd *cobra.Command, args []string) {
		path := logFilePath()
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No log file exists yet")
				return
			}
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Print(string(data))
	},
}

var logClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear log file",
	Run: func(cmd *cobra.Command, args []string) {
		path := logFilePath()
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("Log cleared")
	},
}

func logFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ccyolo", "ccyolo.log")
}

func init() {
	logCmd.AddCommand(logEnableCmd)
	logCmd.AddCommand(logDisableCmd)
	logCmd.AddCommand(logShowCmd)
	logCmd.AddCommand(logClearCmd)
}
