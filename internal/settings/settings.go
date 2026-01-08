package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Hook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type HookMatcher struct {
	Matcher string `json:"matcher"`
	Hooks   []Hook `json:"hooks"`
}

type HooksConfig struct {
	PermissionRequest []HookMatcher `json:"PermissionRequest,omitempty"`
}

type Settings struct {
	Hooks   HooksConfig            `json:"hooks,omitempty"`
	Unknown map[string]interface{} `json:"-"` // preserve other fields
}

func ClaudeSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func loadRaw() (map[string]interface{}, error) {
	data, err := os.ReadFile(ClaudeSettingsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	return raw, nil
}

func saveRaw(raw map[string]interface{}) error {
	dir := filepath.Dir(ClaudeSettingsPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ClaudeSettingsPath(), data, 0644)
}

func AddHook(command string) error {
	raw, err := loadRaw()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	// Ensure "hooks" wrapper exists
	hooks, ok := raw["hooks"].(map[string]interface{})
	if !ok {
		hooks = make(map[string]interface{})
		raw["hooks"] = hooks
	}

	// Check if ccyolo hook already exists in PreToolUse
	permReq, _ := hooks["PreToolUse"].([]interface{})
	for _, item := range permReq {
		if matcher, ok := item.(map[string]interface{}); ok {
			if hooksList, ok := matcher["hooks"].([]interface{}); ok {
				for _, h := range hooksList {
					if hook, ok := h.(map[string]interface{}); ok {
						if cmd, ok := hook["command"].(string); ok {
							if cmd == command || containsCCYolo(cmd) {
								return fmt.Errorf("ccyolo hook already installed")
							}
						}
					}
				}
			}
		}
	}

	// Add new hook to PreToolUse inside "hooks" wrapper
	newHook := map[string]interface{}{
		"matcher": "*",
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": command,
			},
		},
	}

	permReq = append(permReq, newHook)
	hooks["PreToolUse"] = permReq

	return saveRaw(raw)
}

func RemoveHook() error {
	raw, err := loadRaw()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	hooks, ok := raw["hooks"].(map[string]interface{})
	if !ok {
		return nil // No hooks section
	}

	permReq, ok := hooks["PreToolUse"].([]interface{})
	if !ok {
		return nil // No PreToolUse hooks
	}

	// Filter out ccyolo hooks
	var filtered []interface{}
	for _, item := range permReq {
		keep := true
		if matcher, ok := item.(map[string]interface{}); ok {
			if hooksList, ok := matcher["hooks"].([]interface{}); ok {
				for _, h := range hooksList {
					if hook, ok := h.(map[string]interface{}); ok {
						if cmd, ok := hook["command"].(string); ok {
							if containsCCYolo(cmd) {
								keep = false
								break
							}
						}
					}
				}
			}
		}
		if keep {
			filtered = append(filtered, item)
		}
	}

	if len(filtered) == 0 {
		delete(hooks, "PreToolUse")
	} else {
		hooks["PreToolUse"] = filtered
	}

	return saveRaw(raw)
}

func containsCCYolo(s string) bool {
	return len(s) >= 6 && (s[:6] == "ccyolo" ||
		(len(s) > 6 && (s[len(s)-6:] == "ccyolo" ||
		 contains(s, "ccyolo "))))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// IsHookInstalled checks if ccyolo hook is registered with Claude Code
func IsHookInstalled() bool {
	raw, err := loadRaw()
	if err != nil {
		return false
	}

	hooks, ok := raw["hooks"].(map[string]interface{})
	if !ok {
		return false
	}

	permReq, ok := hooks["PreToolUse"].([]interface{})
	if !ok {
		return false
	}

	for _, item := range permReq {
		if matcher, ok := item.(map[string]interface{}); ok {
			if hooksList, ok := matcher["hooks"].([]interface{}); ok {
				for _, h := range hooksList {
					if hook, ok := h.(map[string]interface{}); ok {
						if cmd, ok := hook["command"].(string); ok {
							if containsCCYolo(cmd) {
								return true
							}
						}
					}
				}
			}
		}
	}

	return false
}
