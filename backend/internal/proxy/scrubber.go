package proxy

import (
	"fmt"
	"regexp"
)

var (
	ssnRegex        = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	creditCardRegex = regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`)
	emailRegex      = regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`)
	apiKeyRegex     = regexp.MustCompile(`\b(?:sk-proj-[a-zA-Z0-9]{30,60}|sk-[a-zA-Z0-9]{48}|AIzaSy[a-zA-Z0-9_-]{33})\b`)
)

type ScrubResult struct {
	Text          string
	PlaceholderMap map[string]string
	RedactedCount int
}

func ScrubText(text string) *ScrubResult {
	placeholderMap := make(map[string]string)
	redactedCount := 0

	// 1. Scrub SSNs
	text = ssnRegex.ReplaceAllStringFunc(text, func(match string) string {
		redactedCount++
		placeholder := fmt.Sprintf("[REDACTED_SSN_%d]", redactedCount)
		placeholderMap[placeholder] = match
		return placeholder
	})

	// 2. Scrub Emails
	text = emailRegex.ReplaceAllStringFunc(text, func(match string) string {
		redactedCount++
		placeholder := fmt.Sprintf("[REDACTED_EMAIL_%d]", redactedCount)
		placeholderMap[placeholder] = match
		return placeholder
	})

	// 3. Scrub API Keys & Secrets
	text = apiKeyRegex.ReplaceAllStringFunc(text, func(match string) string {
		redactedCount++
		placeholder := fmt.Sprintf("[REDACTED_SECRET_%d]", redactedCount)
		placeholderMap[placeholder] = match
		return placeholder
	})

	// 4. Scrub Credit Cards (Verify length is indeed card-like to avoid false positives on random zip codes)
	text = creditCardRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Strip spaces/dashes to verify length
		stripped := ""
		for _, r := range match {
			if r >= '0' && r <= '9' {
				stripped += string(r)
			}
		}
		if len(stripped) >= 13 && len(stripped) <= 16 {
			redactedCount++
			placeholder := fmt.Sprintf("[REDACTED_CARD_%d]", redactedCount)
			placeholderMap[placeholder] = match
			return placeholder
		}
		return match
	})

	return &ScrubResult{
		Text:          text,
		PlaceholderMap: placeholderMap,
		RedactedCount: redactedCount,
	}
}

func UnscrubText(text string, placeholderMap map[string]string) string {
	for placeholder, originalValue := range placeholderMap {
		// Simple direct replace
		text = regexp.MustCompile(regexp.QuoteMeta(placeholder)).ReplaceAllString(text, originalValue)
	}
	return text
}
