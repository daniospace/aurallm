package model

import "time"

type Team struct {
	ID          string  `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	BudgetLimit float64 `json:"budget_limit" db:"budget_limit"`
	BudgetUsed  float64 `json:"budget_used" db:"budget_used"`
}

type GatewayKey struct {
	KeyHash   string    `json:"key_hash" db:"key_hash"`
	TeamID    string    `json:"team_id" db:"team_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Status    string    `json:"status" db:"status"` // "active", "revoked"
}

type ProviderConfig struct {
	ProviderName string `json:"provider_name" db:"provider_name"`
	APIKey       string `json:"api_key" db:"api_key"` // Encrypted
	RoutingRules string `json:"routing_rules" db:"routing_rules"`
}

type UsageLog struct {
	ID               string    `json:"id" db:"id"`
	TeamID           string    `json:"team_id" db:"team_id"`
	Model            string    `json:"model" db:"model"`
	PromptTokens     int       `json:"prompt_tokens" db:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens" db:"completion_tokens"`
	Cost             float64   `json:"cost" db:"cost"`
	LatencyMS        int64     `json:"latency_ms" db:"latency_ms"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	IsShadow         bool      `json:"is_shadow" db:"is_shadow"`
	PrimaryLogID     string    `json:"primary_log_id" db:"primary_log_id"`
	PIIRedactedCount int       `json:"pii_redacted_count" db:"pii_redacted_count"`
}
