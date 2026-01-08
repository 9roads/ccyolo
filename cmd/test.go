package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/9roads/ccyolo/internal/claude"
	"github.com/9roads/ccyolo/internal/config"
	"github.com/9roads/ccyolo/internal/preset"
	"github.com/spf13/cobra"
)

var (
	testRulesOnly bool
	testVerbose   bool
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run preset test cases",
	Long: `Run the test cases defined in the current preset.

Tests each simulated hook input against rules and optionally the LLM,
comparing results with expected outcomes.

Use --rules-only to skip LLM evaluation and only test static rules.
Use --verbose to see details of each test.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTests()
	},
}

func init() {
	testCmd.Flags().BoolVar(&testRulesOnly, "rules-only", false, "Only test static rules, skip LLM evaluation")
	testCmd.Flags().BoolVarP(&testVerbose, "verbose", "v", false, "Show details for each test")
	rootCmd.AddCommand(testCmd)
}

func runTests() {
	cfg := config.Load()
	p := preset.Get(cfg.Preset)

	fmt.Printf("Testing preset: %s\n", p.Name)
	fmt.Printf("Model: %s\n", cfg.Model)
	if testRulesOnly {
		fmt.Println("Mode: rules-only (skipping LLM)")
	} else {
		fmt.Println("Mode: full (rules + LLM)")
	}
	fmt.Println()

	if len(p.Tests) == 0 {
		fmt.Println("No test cases defined for this preset.")
		return
	}

	// Check API key if not rules-only
	apiKey := ""
	if !testRulesOnly {
		apiKey = config.GetAPIKey()
		if apiKey == "" {
			fmt.Println("Warning: No API key configured, falling back to rules-only mode")
			testRulesOnly = true
		}
	}

	passed := 0
	failed := 0

	for i, tc := range p.Tests {
		result, source := evaluateTestCase(tc, p, apiKey, cfg.Model)

		// Determine actual result
		actualStr := "ask"
		if result != nil && *result {
			actualStr = "allow"
		}

		// Compare with expected
		status := ""
		if actualStr == tc.Expect {
			status = "PASS"
			passed++
		} else {
			status = "FAIL"
			failed++
		}

		expectedStr := strings.ToUpper(tc.Expect)

		if testVerbose || status == "FAIL" {
			icon := "✓"
			if status == "FAIL" {
				icon = "✗"
			}
			fmt.Printf("%s [%d] %s\n", icon, i+1, tc.Name)
			fmt.Printf("    Tool: %s\n", tc.Tool)
			fmt.Printf("    Expected: %s, Got: %s (%s)\n", expectedStr, strings.ToUpper(actualStr), source)
			if status == "FAIL" {
				fmt.Printf("    Input: %v\n", tc.Input)
			}
			fmt.Println()
		} else {
			fmt.Printf("✓ %s\n", tc.Name)
		}
	}

	fmt.Println()
	fmt.Printf("Results: %d passed, %d failed (total: %d)\n", passed, failed, len(p.Tests))

	if failed > 0 {
		os.Exit(1)
	}
}

// evaluateTestCase runs a test case through the hook logic
// Returns: result (nil = ask user, true = allow, false = deny), source string
func evaluateTestCase(tc preset.TestCase, p preset.Preset, apiKey, model string) (*bool, string) {
	// Step 1: Check static rules
	ruleResult := preset.CheckRules(tc.Tool, tc.Input, p)
	if ruleResult != nil {
		return ruleResult, "rule"
	}

	// Step 2: Check cache (skip for tests - we want fresh evaluation)
	// In test mode, we clear cache to ensure fresh LLM evaluation

	// Step 3: If rules-only mode, return nil (would ask user)
	if testRulesOnly {
		return nil, "no-rule"
	}

	// Step 4: Ask Claude API
	if apiKey == "" {
		return nil, "no-api-key"
	}

	result, reason, err := claude.EvaluateSafety(apiKey, model, p.Prompt, tc.Tool, tc.Input)
	if err != nil {
		return nil, "api-error"
	}

	if result == nil {
		return nil, "api-uncertain"
	}

	// Don't cache test results
	_ = reason
	return result, "llm"
}
