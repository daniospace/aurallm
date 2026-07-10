package proxy

import "testing"

func TestScrubText(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedScrub string
		expectedCount int
	}{
		{
			name:          "Scrub SSN",
			input:         "My SSN is 123-45-6789.",
			expectedScrub: "My SSN is [REDACTED_SSN_1].",
			expectedCount: 1,
		},
		{
			name:          "Scrub Email",
			input:         "Contact support@aurallm.ai for help.",
			expectedScrub: "Contact [REDACTED_EMAIL_1] for help.",
			expectedCount: 1,
		},
		{
			name:          "Scrub OpenAI API Secret Key",
			input:         "Here is my key: sk-proj-1234567890abcdefghijklmnopqrstuvwxyzabcdef.",
			expectedScrub: "Here is my key: [REDACTED_SECRET_1].",
			expectedCount: 1,
		},
		{
			name:          "Scrub Visa Credit Card",
			input:         "My credit card is 4111-2222-3333-4444.",
			expectedScrub: "My credit card is [REDACTED_CARD_1].",
			expectedCount: 1,
		},
		{
			name:          "Scrub Mixed Content",
			input:         "User alex@gmail.com submitted card 1234-5678-1234-5678 and SSN 999-12-3456.",
			expectedScrub: "User [REDACTED_EMAIL_2] submitted card [REDACTED_CARD_3] and SSN [REDACTED_SSN_1].",
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ScrubText(tt.input)
			if res.Text != tt.expectedScrub {
				t.Errorf("ScrubText() got = %q; want %q", res.Text, tt.expectedScrub)
			}
			if res.RedactedCount != tt.expectedCount {
				t.Errorf("ScrubText() redacted count got = %d; want %d", res.RedactedCount, tt.expectedCount)
			}

			// Verify Unscrub recovers original content perfectly!
			unscrubbed := UnscrubText(res.Text, res.PlaceholderMap)
			if unscrubbed != tt.input {
				t.Errorf("UnscrubText() did not recover original text. got = %q; want %q", unscrubbed, tt.input)
			}
		})
	}
}
