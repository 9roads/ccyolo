package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type ClaudeSettings struct {
	Permissions struct {
		Allow []string `json:"allow"`
		Deny  []string `json:"deny"`
	} `json:"permissions"`
	DisableAllHooks bool `json:"disableAllHooks"`
}

func claudeSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.local.json")
}

func loadClaudeSettings() (*ClaudeSettings, error) {
	data, err := os.ReadFile(claudeSettingsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &ClaudeSettings{}, nil
		}
		return nil, err
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

func saveClaudeSettings(settings *ClaudeSettings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(claudeSettingsPath(), data, 0644)
}

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage Claude Code allow rules",
	Long:  `View and suggest additions to Claude Code's ~/.claude/settings.local.json`,
	Run: func(cmd *cobra.Command, args []string) {
		listRules()
	},
}

var rulesSuggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Analyze logs and suggest rules to add",
	Run: func(cmd *cobra.Command, args []string) {
		suggestRules()
	},
}

var rulesInstallCommandCmd = &cobra.Command{
	Use:   "install-command",
	Short: "Install /ccyolo-rules slash command for Claude Code",
	Run: func(cmd *cobra.Command, args []string) {
		installSlashCommand()
	},
}

func listRules() {
	settings, err := loadClaudeSettings()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading settings: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Claude Code allow rules (%s):\n\n", claudeSettingsPath())

	if len(settings.Permissions.Allow) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, rule := range settings.Permissions.Allow {
			fmt.Printf("  %s\n", rule)
		}
	}
}

func suggestRules() {
	// Read log file
	home, _ := os.UserHomeDir()
	logPath := filepath.Join(home, ".ccyolo", "ccyolo.log")

	data, err := os.ReadFile(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading logs: %v\n", err)
		fmt.Println("No logs found. Run some commands with ccyolo enabled first.")
		os.Exit(1)
	}

	// Load existing rules
	settings, _ := loadClaudeSettings()
	existingMap := make(map[string]bool)
	for _, r := range settings.Permissions.Allow {
		existingMap[r] = true
	}

	suggestions := extractSuggestions(string(data), existingMap)

	if len(suggestions) == 0 {
		fmt.Println("No new rules to suggest.")
		return
	}

	fmt.Println("Suggested additions to Claude Code allow list:")
	for _, s := range suggestions {
		fmt.Printf("  %s\n", s)
	}
	fmt.Println("\nTo add these, edit ~/.claude/settings.local.json")
	fmt.Println("or use /ccyolo-rules in Claude Code.")
}

func extractSuggestions(logContent string, existing map[string]bool) []string {
	suggestions := make(map[string]bool)

	lines := strings.Split(logContent, "\n")
	for _, line := range lines {
		// Look for API-allowed Bash commands
		if strings.Contains(line, "respond: allow") && strings.Contains(line, "(AI:") && strings.Contains(line, "Bash:") {
			rule := extractBashRule(line)
			if rule != "" && !existing[rule] {
				suggestions[rule] = true
			}
		}
	}

	result := make([]string, 0, len(suggestions))
	for rule := range suggestions {
		result = append(result, rule)
	}
	return result
}

func extractBashRule(line string) string {
	// Extract command from: "respond: allow - [YOLO] Bash: ls -la /path (AI: ...)"
	idx := strings.Index(line, "Bash: ")
	if idx == -1 {
		return ""
	}

	cmd := line[idx+6:]
	if i := strings.Index(cmd, " (AI:"); i > 0 {
		cmd = strings.TrimSpace(cmd[:i])
	}

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}

	base := parts[0]

	// Generate Claude Code format rules
	switch base {
	case "ls", "cat", "head", "tail", "find", "which", "wc", "file", "stat", "du", "df", "pwd", "echo", "grep":
		return fmt.Sprintf("Bash(%s:*)", base)
	case "git":
		if len(parts) > 1 {
			switch parts[1] {
			case "status", "log", "diff", "branch", "show", "remote":
				return fmt.Sprintf("Bash(git %s:*)", parts[1])
			}
		}
	}

	return ""
}

func installSlashCommand() {
	home, _ := os.UserHomeDir()
	destDir := filepath.Join(home, ".claude", "commands")
	destPath := filepath.Join(destDir, "ccyolo-rules.md")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	content := `# Curate ccyolo Allow Rules

Analyze ccyolo logs and help curate Claude Code's allow list to skip the hook entirely for safe patterns.

## Step 1: Gather Info

` + "```bash" + `
ccyolo rules          # Current Claude Code allow rules
ccyolo rules suggest  # Suggestions based on logs
` + "```" + `

Also read ` + "`~/.ccyolo/ccyolo.log`" + ` to see all hook decisions.

## Step 2: Analyze

Look at the logs and identify:
- Patterns that were API-allowed repeatedly (candidates for Claude Code's allow list)
- Patterns that were correctly denied (validate the AI is working)

## Step 3: Update Settings

Edit ` + "`~/.claude/settings.local.json`" + ` and add to the permissions.allow array.

Format: ` + "`Bash(command:*)`" + ` where ` + "`*`" + ` is wildcard.

Examples:
` + "```json" + `
{
  "permissions": {
    "allow": [
      "Bash(ls:*)",
      "Bash(cat:*)",
      "Bash(git status:*)",
      "Bash(git log:*)"
    ]
  }
}
` + "```" + `

Rules in Claude Code's allow list skip the hook entirely = faster + no API cost.
`

	if err := os.WriteFile(destPath, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Installed: %s\n", destPath)
	fmt.Println("\nUse /ccyolo-rules in Claude Code to analyze logs and curate rules.")
}

func init() {
	rulesCmd.AddCommand(rulesSuggestCmd)
	rulesCmd.AddCommand(rulesInstallCommandCmd)
	rootCmd.AddCommand(rulesCmd)
}
