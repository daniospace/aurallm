package proxy

import "testing"

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		model            string
		promptTokens     int
		completionTokens int
		expectedCost     float64
	}{
		{
			model:            "gpt-4o",
			promptTokens:     100000, // 0.1 Million
			completionTokens: 200000, // 0.2 Million
			expectedCost:     (0.1 * 5.00) + (0.2 * 15.00), // 0.5 + 3.0 = 3.5
		},
		{
			model:            "gpt-3.5-turbo",
			promptTokens:     1000000, // 1 Million
			completionTokens: 1000000, // 1 Million
			expectedCost:     (1.0 * 0.50) + (1.0 * 1.50), // 0.5 + 1.5 = 2.0
		},
		{
			model:            "claude-3-5-sonnet",
			promptTokens:     10000,  // 0.01 Million
			completionTokens: 50000,  // 0.05 Million
			expectedCost:     (0.01 * 3.00) + (0.05 * 15.00), // 0.03 + 0.75 = 0.78
		},
		{
			model:            "unknown-model-fallback",
			promptTokens:     1000000, // 1 Million
			completionTokens: 1000000, // 1 Million
			expectedCost:     (1.0 * 1.00) + (1.0 * 5.00), // 1.0 + 5.0 = 6.0
		},
	}

	for _, tt := range tests {
		got := CalculateCost(tt.model, tt.promptTokens, tt.completionTokens)
		if got != tt.expectedCost {
			t.Errorf("CalculateCost(%q, %d, %d) = %f; want %f", tt.model, tt.promptTokens, tt.completionTokens, got, tt.expectedCost)
		}
	}
}
