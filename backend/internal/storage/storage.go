package storage

import (
	"context"
	"gateway/internal/model"
)

type Storage interface {
	GetTeam(ctx context.Context, id string) (*model.Team, error)
	CreateTeam(ctx context.Context, team *model.Team) error
	UpdateTeamBudget(ctx context.Context, id string, amount float64) error
	ListTeams(ctx context.Context) ([]*model.Team, error)
	DeleteTeam(ctx context.Context, id string) error
	
	GetGatewayKey(ctx context.Context, keyHash string) (*model.GatewayKey, error)
	CreateGatewayKey(ctx context.Context, key *model.GatewayKey) error
	ListGatewayKeys(ctx context.Context) ([]*model.GatewayKey, error)
	DeleteGatewayKey(ctx context.Context, keyHash string) error
	
	GetProviderConfig(ctx context.Context, name string) (*model.ProviderConfig, error)
	CreateProviderConfig(ctx context.Context, config *model.ProviderConfig) error
	
	CreateUsageLog(ctx context.Context, log *model.UsageLog) error
	GetUsageLogs(ctx context.Context, teamID string) ([]*model.UsageLog, error)
	
	Close() error
}
