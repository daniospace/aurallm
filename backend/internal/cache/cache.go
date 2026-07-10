package cache

import (
	"context"
	"gateway/internal/model"
)

type Cache interface {
	GetTeamBudget(ctx context.Context, teamID string) (limit float64, used float64, found bool, err error)
	SetTeamBudget(ctx context.Context, teamID string, limit float64, used float64) error
	IncrementTeamBudgetUsed(ctx context.Context, teamID string, amount float64) (newUsed float64, err error)

	GetGatewayKey(ctx context.Context, keyHash string) (*model.GatewayKey, bool, error)
	SetGatewayKey(ctx context.Context, keyHash string, key *model.GatewayKey) error

	InvalidateGatewayKey(ctx context.Context, keyHash string) error
	InvalidateTeam(ctx context.Context, teamID string) error
}
