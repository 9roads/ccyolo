package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/9roads/ccyolo/internal/cache"
	"github.com/9roads/ccyolo/internal/claude"
	"github.com/9roads/ccyolo/internal/config"
	"github.com/9roads/ccyolo/internal/preset"
	"github.com/spf13/cobra"
)

var (
	logFile    *os.File
	logEnabled bool
)

func initLogging(enabled bool) {
	logEnabled = enabled
	if enabled && logFile == nil {
		home, _ := os.UserHomeDir()
		logDir := filepath.Join(home, ".ccyolo")
		os.MkdirAll(logDir, 0755)
		logPath := filepath.Join(logDir, "ccyolo.log")
		logFile, _ = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
}

func logMsg(format string, args ...interface{}) {
	if !logEnabled || logFile == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logFile, "[%s] %s\n", time.Now().Format("15:04:05"), msg)
}

type HookInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

type HookSpecificOutput struct {
	HookEventName      string `json:"hookEventName"`
	PermissionDecision string `json:"permissionDecision"`
	Reason             string `json:"reason,omitempty"`
}

type HookResponse struct {
	HookSpecificOutput HookSpecificOutput `json:"hookSpecificOutput"`
}

var hookCmd = &cobra.Command{
	Use:    "hook",
	Short:  "Handle Claude Code permission request (internal)",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		runHook()
	},
}

func runHook() {
	cfg := config.Load()
	initLogging(cfg.Logging)

	logMsg("=== hook called ===")
	logMsg("config: enabled=%v, preset=%s", cfg.Enabled, cfg.Preset)

	// If disabled, pass through
	if !cfg.Enabled {
		logMsg("disabled, passing through")
		fmt.Println("{}")
		return
	}

	// Read raw input first for logging
	rawInput, err := io.ReadAll(os.Stdin)
	if err != nil {
		logMsg("read error: %v", err)
		fmt.Println("{}")
		return
	}
	logMsg("raw input: %s", string(rawInput))

	// Parse input
	var input HookInput
	if err := json.Unmarshal(rawInput, &input); err != nil {
		logMsg("parse error: %v", err)
		fmt.Fprintln(os.Stderr, "[ccyolo] parse error:", err)
		fmt.Println("{}")
		return
	}

	toolName := input.ToolName
	toolInput := input.ToolInput
	logMsg("tool: %s, input: %+v", toolName, toolInput)

	// Load preset
	p := preset.Get(cfg.Preset)

	// Step 1: Check static rules
	ruleResult := preset.CheckRules(toolName, toolInput, p)
	logMsg("rule check result: %v", ruleResult)

	if ruleResult != nil {
		if *ruleResult {
			logMsg("rule ALLOW")
			respond(true, "rule", toolName, toolInput)
		} else {
			logMsg("rule DENY, asking user")
			fmt.Println("{}")
		}
		return
	}

	// Step 2: Check cache
	cachedResult := cache.Get(toolName, toolInput, cfg.Preset)
	logMsg("cache result: %v", cachedResult)
	if cachedResult != nil {
		if *cachedResult {
			logMsg("cache ALLOW")
			respond(true, "cached", toolName, toolInput)
		} else {
			logMsg("cache DENY")
			fmt.Println("{}")
		}
		return
	}

	// Step 3: Ask Claude API
	apiKey := config.GetAPIKey()
	if apiKey == "" {
		logMsg("no API key")
		fmt.Fprintln(os.Stderr, "[ccyolo] no API key configured")
		fmt.Println("{}")
		return
	}
	logMsg("calling Claude API...")

	result, reason, err := claude.EvaluateSafety(apiKey, cfg.Model, p.Prompt, toolName, toolInput)
	if err != nil {
		logMsg("API error: %v", err)
		fmt.Fprintln(os.Stderr, "[ccyolo] API error:", err)
		fmt.Println("{}")
		return
	}
	logMsg("API result: %v, reason: %s", result, reason)

	// Cache the result
	if result != nil {
		cache.Set(toolName, toolInput, cfg.Preset, *result)
	}

	if result != nil && *result {
		logMsg("API ALLOW")
		respond(true, "AI: "+reason, toolName, toolInput)
	} else {
		logMsg("API DENY or nil")
		fmt.Println("{}")
	}
}

func respond(allow bool, reason, toolName string, toolInput map[string]interface{}) {
	summary := getOperationSummary(toolName, toolInput)

	decision := "deny"
	if allow {
		decision = "allow"
	}

	msg := fmt.Sprintf("[YOLO] %s (%s)", summary, reason)
	logMsg("respond: %s - %s", decision, msg)

	response := HookResponse{
		HookSpecificOutput: HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: decision,
			Reason:             msg,
		},
	}

	data, _ := json.Marshal(response)
	logMsg("respond: %s", string(data))
	fmt.Println(string(data))
	// Exit 0 for success (approval or denial handled via JSON)
}

func getOperationSummary(toolName string, toolInput map[string]interface{}) string {
	switch toolName {
	case "Bash":
		if cmd, ok := toolInput["command"].(string); ok {
			if len(cmd) > 60 {
				cmd = cmd[:57] + "..."
			}
			return "Bash: " + cmd
		}
	case "Read", "Write", "Edit", "Glob":
		path := ""
		if p, ok := toolInput["file_path"].(string); ok {
			path = p
		} else if p, ok := toolInput["path"].(string); ok {
			path = p
		}
		if idx := strings.LastIndex(path, "/"); idx >= 0 {
			path = "..." + path[idx:]
		}
		return toolName + ": " + path
	case "Grep":
		if pattern, ok := toolInput["pattern"].(string); ok {
			if len(pattern) > 30 {
				pattern = pattern[:27] + "..."
			}
			return "Grep: " + pattern
		}
	}
	return toolName
}
