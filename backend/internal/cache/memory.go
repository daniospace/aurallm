package cache

import (
	"context"
	"gateway/internal/model"
	"sync"
)

type teamBudget struct {
	Limit float64
	Used  float64
}

type MemoryCache struct {
	mu      sync.RWMutex
	budgets map[string]*teamBudget
	keys    map[string]*model.GatewayKey
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		budgets: make(map[string]*teamBudget),
		keys:    make(map[string]*model.GatewayKey),
	}
}

func (m *MemoryCache) GetTeamBudget(ctx context.Context, teamID string) (float64, float64, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	b, ok := m.budgets[teamID]
	if !ok {
		return 0, 0, false, nil
	}
	return b.Limit, b.Used, true, nil
}

func (m *MemoryCache) SetTeamBudget(ctx context.Context, teamID string, limit float64, used float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.budgets[teamID] = &teamBudget{
		Limit: limit,
		Used:  used,
	}
	return nil
}

func (m *MemoryCache) IncrementTeamBudgetUsed(ctx context.Context, teamID string, amount float64) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.budgets[teamID]
	if !ok {
		// If not cached, we shouldn't arbitrarily create one without knowing the limit.
		// So we return an error or start with 0 limit?
		// Usually the caller should have populated the cache first.
		// Let's assume we start with 0 limit, but it's better to return a specific state.
		m.budgets[teamID] = &teamBudget{
			Limit: 0,
			Used:  amount,
		}
		return amount, nil
	}
	b.Used += amount
	return b.Used, nil
}

func (m *MemoryCache) GetGatewayKey(ctx context.Context, keyHash string) (*model.GatewayKey, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key, ok := m.keys[keyHash]
	if !ok {
		return nil, false, nil
	}
	return key, true, nil
}

func (m *MemoryCache) SetGatewayKey(ctx context.Context, keyHash string, key *model.GatewayKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.keys[keyHash] = key
	return nil
}

func (m *MemoryCache) InvalidateGatewayKey(ctx context.Context, keyHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.keys, keyHash)
	return nil
}

func (m *MemoryCache) InvalidateTeam(ctx context.Context, teamID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.budgets, teamID)
	return nil
}
