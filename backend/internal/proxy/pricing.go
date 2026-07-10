package proxy

import "strings"

type ModelPrice struct {
	InputPricePerMillion  float64
	OutputPricePerMillion float64
}

var modelPricing = map[string]ModelPrice{
	"gpt-4o": {
		InputPricePerMillion:  5.00,
		OutputPricePerMillion: 15.00,
	},
	"gpt-4-turbo": {
		InputPricePerMillion:  10.00,
		OutputPricePerMillion: 30.00,
	},
	"gpt-3.5-turbo": {
		InputPricePerMillion:  0.50,
		OutputPricePerMillion: 1.50,
	},
	"claude-3-5-sonnet-20240620": {
		InputPricePerMillion:  3.00,
		OutputPricePerMillion: 15.00,
	},
	"claude-3-5-sonnet": {
		InputPricePerMillion:  3.00,
		OutputPricePerMillion: 15.00,
	},
	"claude-3-opus-20240229": {
		InputPricePerMillion:  15.00,
		OutputPricePerMillion: 75.00,
	},
	"claude-3-haiku-20240307": {
		InputPricePerMillion:  0.25,
		OutputPricePerMillion: 1.25,
	},
}

func CalculateCost(model string, promptTokens, completionTokens int) float64 {
	// Standardize key check
	key := strings.ToLower(model)
	price, exists := modelPricing[key]
	if !exists {
		// Fallback price for unknown models (e.g. standard average price)
		price = ModelPrice{
			InputPricePerMillion:  1.00,
			OutputPricePerMillion: 5.00,
		}
	}

	inputCost := (float64(promptTokens) / 1000000.0) * price.InputPricePerMillion
	outputCost := (float64(completionTokens) / 1000000.0) * price.OutputPricePerMillion
	return inputCost + outputCost
}
