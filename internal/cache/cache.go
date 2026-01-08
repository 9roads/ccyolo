package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/9roads/ccyolo/internal/config"
)

type Entry struct {
	Approve   bool  `json:"approve"`
	Timestamp int64 `json:"timestamp"`
}

func getCacheKey(toolName string, toolInput map[string]interface{}, preset string) string {
	// Normalize input for caching
	var normalized string

	if toolName == "Bash" {
		if cmd, ok := toolInput["command"].(string); ok {
			normalized = normalizeCommand(cmd)
		}
	} else {
		data, _ := json.Marshal(toolInput)
		normalized = string(data)
	}

	input := preset + ":" + toolName + ":" + normalized
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:8])
}

func normalizeCommand(cmd string) string {
	patterns := []struct {
		regex   string
		replace string
	}{
		{`^(npm|yarn|pnpm)\s+(install|add|remove)\s+.+`, "$1 $2 *"},
		{`^pip3?\s+install\s+.+`, "pip install *"},
		{`^git\s+commit\s+.+`, "git commit *"},
		{`^rm\s+[^-].*`, "rm *"},      // rm without flags
		{`^rm\s+-[^r].*`, "rm -* *"},  // rm with flags but not -r
		{`^mkdir\s+.+`, "mkdir *"},
		{`^(cat|head|tail)\s+.+`, "$1 *"},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.regex)
		if re.MatchString(cmd) {
			return re.ReplaceAllString(cmd, p.replace)
		}
	}

	return cmd
}

func Get(toolName string, toolInput map[string]interface{}, preset string) *bool {
	cfg := config.Load()
	key := getCacheKey(toolName, toolInput, preset)
	cacheFile := filepath.Join(config.CacheDir(), key+".json")

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}

	// Check expiration
	if time.Now().Unix()-entry.Timestamp > int64(cfg.CacheTTL) {
		os.Remove(cacheFile)
		return nil
	}

	return &entry.Approve
}

func Set(toolName string, toolInput map[string]interface{}, preset string, approve bool) {
	key := getCacheKey(toolName, toolInput, preset)

	if err := os.MkdirAll(config.CacheDir(), 0755); err != nil {
		return
	}

	entry := Entry{
		Approve:   approve,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	cacheFile := filepath.Join(config.CacheDir(), key+".json")
	os.WriteFile(cacheFile, data, 0644)
}

func Clear() error {
	return os.RemoveAll(config.CacheDir())
}
