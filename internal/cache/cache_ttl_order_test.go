package cache

import (
	"strings"
	"testing"

	"autocache/internal/types"

	"github.com/sirupsen/logrus"
)

func TestCacheTTLOrdering(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	injector := NewCacheInjector(types.StrategyAggressive, "https://api.anthropic.com", "test-key", logger)

	// Create a test request where some content would normally get 5m TTL
	// but should be upgraded to 1h because there's 1h content later
	req := &types.AnthropicRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 100,
		
		// This normally gets 1h TTL (system)
		System: strings.Repeat("You are a helpful assistant. Instructions: Please follow these detailed guidelines carefully when responding to user queries. " +
			"Context: This is important background information that rarely changes. " +
			"Guidelines: Always be accurate, helpful, and concise in your responses. ", 50), // Much larger system prompt
		
		Messages: []types.Message{
			{
				Role: "user",
				Content: []types.ContentBlock{
					{
						Type: "text",
						// This normally would get 5m TTL, but should be upgraded to 1h 
						// because there's stable content (1h) later in the request
						Text: strings.Repeat("This is a large user message that would normally get 5m TTL because it doesn't contain stable patterns. " +
							"It's just a regular user query that could change frequently. However, since there's stable content later " +
							"in this request, this should be upgraded to 1h TTL for hierarchy consistency. " +
							"Here's more content to ensure it gets cached: Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 50),
					},
				},
			},
			{
				Role: "user", 
				Content: []types.ContentBlock{
					{
						Type: "text",
						// This should get 1h TTL (content with stable patterns) and trigger upgrades
						Text: strings.Repeat("You are an expert analyst. Instructions: Always provide comprehensive analysis. " +
							"Guidelines: Follow these detailed steps when analyzing data or information provided by users. " +
							"Context: This is additional context that provides important background information for your analysis. " +
							"Your role is to examine the information carefully and provide detailed insights and recommendations. " +
							"Please ensure your analysis is thorough, accurate, and well-structured. Consider multiple perspectives. ", 40),
					},
				},
			},
		},
	}

	// Inject cache control
	metadata, err := injector.InjectCacheControl(req)
	if err != nil {
		t.Fatalf("Failed to inject cache control: %v", err)
	}

	if !metadata.CacheInjected {
		t.Fatal("Expected cache to be injected")
	}

	// Verify that we have breakpoints
	if len(metadata.Breakpoints) < 2 {
		t.Fatalf("Expected at least 2 breakpoints, got %d", len(metadata.Breakpoints))
	}

	// All breakpoints should now be 1h due to upgrade logic
	for i, breakpoint := range metadata.Breakpoints {
		if breakpoint.TTL != "1h" {
			t.Errorf("Expected all breakpoints to be upgraded to 1h, but breakpoint %d (%s) has TTL %s", 
				i, breakpoint.Position, breakpoint.TTL)
		}
	}

	t.Logf("All cache breakpoints upgraded to 1h TTL as expected")
	for _, bp := range metadata.Breakpoints {
		t.Logf("- %s: %s (%d tokens)", bp.Position, bp.TTL, bp.Tokens)
	}
}

func TestTTLUpgradeLogic(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	injector := NewCacheInjector(types.StrategyModerate, "https://api.anthropic.com", "test-key", logger)

	// Test case 1: No 1h TTL, no upgrades should happen
	t.Run("No upgrades needed", func(t *testing.T) {
		candidates := []CacheCandidate{
			{Position: "message_0", TTL: "5m", Tokens: 1000, ContentType: "content"},
			{Position: "message_1", TTL: "5m", Tokens: 1500, ContentType: "content"},
		}

		injector.upgradeTTLHierarchy(candidates, "claude-3-5-sonnet-20241022")

		for i, candidate := range candidates {
			if candidate.TTL != "5m" {
				t.Errorf("Candidate %d should remain 5m, got %s", i, candidate.TTL)
			}
		}
	})

	// Test case 2: 1h TTL present, all preceding should upgrade
	t.Run("Upgrade preceding candidates", func(t *testing.T) {
		candidates := []CacheCandidate{
			{Position: "system", TTL: "5m", Tokens: 1000, ContentType: "content"},        // Should upgrade
			{Position: "message_0", TTL: "5m", Tokens: 1500, ContentType: "content"},    // Should upgrade  
			{Position: "message_1", TTL: "1h", Tokens: 1200, ContentType: "content"},    // Original 1h
			{Position: "message_2", TTL: "5m", Tokens: 800, ContentType: "content"},     // Should stay 5m
		}

		injector.upgradeTTLHierarchy(candidates, "claude-3-5-sonnet-20241022")

		expectedTTLs := []string{"1h", "1h", "1h", "5m"}
		for i, candidate := range candidates {
			if candidate.TTL != expectedTTLs[i] {
				t.Errorf("Candidate %d (%s) expected TTL %s, got %s", 
					i, candidate.Position, expectedTTLs[i], candidate.TTL)
			}
		}
	})

	// Test case 3: Multiple 1h TTLs, upgrade up to the last one
	t.Run("Multiple 1h TTLs", func(t *testing.T) {
		candidates := []CacheCandidate{
			{Position: "system", TTL: "5m", Tokens: 1000, ContentType: "content"},       // Should upgrade
			{Position: "message_0", TTL: "1h", Tokens: 1500, ContentType: "content"},   // Original 1h
			{Position: "message_1", TTL: "5m", Tokens: 1200, ContentType: "content"},   // Should upgrade (between 1h)
			{Position: "message_2", TTL: "1h", Tokens: 800, ContentType: "content"},    // Original 1h (last)
			{Position: "message_3", TTL: "5m", Tokens: 900, ContentType: "content"},    // Should stay 5m
		}

		injector.upgradeTTLHierarchy(candidates, "claude-3-5-sonnet-20241022")

		expectedTTLs := []string{"1h", "1h", "1h", "1h", "5m"}
		for i, candidate := range candidates {
			if candidate.TTL != expectedTTLs[i] {
				t.Errorf("Candidate %d (%s) expected TTL %s, got %s", 
					i, candidate.Position, expectedTTLs[i], candidate.TTL)
			}
		}
	})
}