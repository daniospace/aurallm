package storage

import (
	"context"
	"database/sql"
	"gateway/internal/model"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) (*PostgresStorage, error) {
	s := &PostgresStorage{db: db}
	if err := s.InitDB(); err != nil {
		return nil, err
	}
	return s, nil
}

func (p *PostgresStorage) InitDB() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS teams (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			budget_limit DOUBLE PRECISION NOT NULL,
			budget_used DOUBLE PRECISION NOT NULL DEFAULT 0.0
		);`,
		`CREATE TABLE IF NOT EXISTS gateway_keys (
			key_hash TEXT PRIMARY KEY,
			team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			status TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS provider_configs (
			provider_name TEXT PRIMARY KEY,
			api_key TEXT NOT NULL,
			routing_rules TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS usage_logs (
			id TEXT PRIMARY KEY,
			team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
			model TEXT NOT NULL,
			prompt_tokens INT NOT NULL,
			completion_tokens INT NOT NULL,
			cost DOUBLE PRECISION NOT NULL,
			latency_ms BIGINT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		);`,
		`ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS is_shadow BOOLEAN DEFAULT FALSE;`,
		`ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS primary_log_id TEXT;`,
		`ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS pii_redacted_count INT DEFAULT 0;`,
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStorage) GetTeam(ctx context.Context, id string) (*model.Team, error) {
	row := p.db.QueryRowContext(ctx, "SELECT id, name, budget_limit, budget_used FROM teams WHERE id = $1", id)
	var t model.Team
	err := row.Scan(&t.ID, &t.Name, &t.BudgetLimit, &t.BudgetUsed)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (p *PostgresStorage) CreateTeam(ctx context.Context, team *model.Team) error {
	_, err := p.db.ExecContext(ctx,
		"INSERT INTO teams (id, name, budget_limit, budget_used) VALUES ($1, $2, $3, $4)",
		team.ID, team.Name, team.BudgetLimit, team.BudgetUsed,
	)
	return err
}

func (p *PostgresStorage) ListTeams(ctx context.Context) ([]*model.Team, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, name, budget_limit, budget_used FROM teams ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*model.Team
	for rows.Next() {
		var t model.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.BudgetLimit, &t.BudgetUsed); err != nil {
			return nil, err
		}
		teams = append(teams, &t)
	}
	return teams, nil
}

func (p *PostgresStorage) DeleteTeam(ctx context.Context, id string) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM teams WHERE id = $1", id)
	return err
}

func (p *PostgresStorage) UpdateTeamBudget(ctx context.Context, id string, amount float64) error {
	_, err := p.db.ExecContext(ctx, "UPDATE teams SET budget_used = budget_used + $1 WHERE id = $2", amount, id)
	return err
}

func (p *PostgresStorage) GetGatewayKey(ctx context.Context, keyHash string) (*model.GatewayKey, error) {
	row := p.db.QueryRowContext(ctx, "SELECT key_hash, team_id, created_at, status FROM gateway_keys WHERE key_hash = $1", keyHash)
	var k model.GatewayKey
	err := row.Scan(&k.KeyHash, &k.TeamID, &k.CreatedAt, &k.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (p *PostgresStorage) CreateGatewayKey(ctx context.Context, key *model.GatewayKey) error {
	_, err := p.db.ExecContext(ctx,
		"INSERT INTO gateway_keys (key_hash, team_id, created_at, status) VALUES ($1, $2, $3, $4)",
		key.KeyHash, key.TeamID, key.CreatedAt, key.Status,
	)
	return err
}

func (p *PostgresStorage) ListGatewayKeys(ctx context.Context) ([]*model.GatewayKey, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT key_hash, team_id, created_at, status FROM gateway_keys ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*model.GatewayKey
	for rows.Next() {
		var k model.GatewayKey
		if err := rows.Scan(&k.KeyHash, &k.TeamID, &k.CreatedAt, &k.Status); err != nil {
			return nil, err
		}
		keys = append(keys, &k)
	}
	return keys, nil
}

func (p *PostgresStorage) DeleteGatewayKey(ctx context.Context, keyHash string) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM gateway_keys WHERE key_hash = $1", keyHash)
	return err
}

func (p *PostgresStorage) GetProviderConfig(ctx context.Context, name string) (*model.ProviderConfig, error) {
	row := p.db.QueryRowContext(ctx, "SELECT provider_name, api_key, routing_rules FROM provider_configs WHERE provider_name = $1", name)
	var c model.ProviderConfig
	err := row.Scan(&c.ProviderName, &c.APIKey, &c.RoutingRules)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (p *PostgresStorage) CreateProviderConfig(ctx context.Context, config *model.ProviderConfig) error {
	_, err := p.db.ExecContext(ctx,
		"INSERT INTO provider_configs (provider_name, api_key, routing_rules) VALUES ($1, $2, $3) ON CONFLICT (provider_name) DO UPDATE SET api_key = EXCLUDED.api_key, routing_rules = EXCLUDED.routing_rules",
		config.ProviderName, config.APIKey, config.RoutingRules,
	)
	return err
}

func (p *PostgresStorage) CreateUsageLog(ctx context.Context, log *model.UsageLog) error {
	_, err := p.db.ExecContext(ctx,
		"INSERT INTO usage_logs (id, team_id, model, prompt_tokens, completion_tokens, cost, latency_ms, created_at, is_shadow, primary_log_id, pii_redacted_count) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		log.ID, log.TeamID, log.Model, log.PromptTokens, log.CompletionTokens, log.Cost, log.LatencyMS, log.CreatedAt, log.IsShadow, log.PrimaryLogID, log.PIIRedactedCount,
	)
	return err
}

func (p *PostgresStorage) GetUsageLogs(ctx context.Context, teamID string) ([]*model.UsageLog, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, team_id, model, prompt_tokens, completion_tokens, cost, latency_ms, created_at, is_shadow, primary_log_id, pii_redacted_count FROM usage_logs WHERE team_id = $1 ORDER BY created_at DESC", teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.UsageLog
	for rows.Next() {
		var l model.UsageLog
		err := rows.Scan(&l.ID, &l.TeamID, &l.Model, &l.PromptTokens, &l.CompletionTokens, &l.Cost, &l.LatencyMS, &l.CreatedAt, &l.IsShadow, &l.PrimaryLogID, &l.PIIRedactedCount)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, nil
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}
