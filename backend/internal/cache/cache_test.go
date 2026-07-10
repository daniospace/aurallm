package cache

import (
	"context"
	"gateway/internal/model"
	"testing"
)

func TestMemoryCache(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	// 1. Test Gateway Key Cache
	keyHash := "some_key_hash"
	key := &model.GatewayKey{
		KeyHash: keyHash,
		TeamID:  "team-1",
		Status:  "active",
	}

	// Get key (not found)
	_, found, err := c.GetGatewayKey(ctx, keyHash)
	if err != nil {
		t.Fatalf("GetGatewayKey failed: %v", err)
	}
	if found {
		t.Errorf("Expected key to not be found initially")
	}

	// Set key
	err = c.SetGatewayKey(ctx, keyHash, key)
	if err != nil {
		t.Fatalf("SetGatewayKey failed: %v", err)
	}

	// Get key (found)
	gotKey, found, err := c.GetGatewayKey(ctx, keyHash)
	if err != nil {
		t.Fatalf("GetGatewayKey failed: %v", err)
	}
	if !found {
		t.Errorf("Expected key to be found")
	}
	if gotKey.TeamID != "team-1" || gotKey.Status != "active" {
		t.Errorf("Retrieved key mismatch: %+v", gotKey)
	}

	// Invalidate key
	err = c.InvalidateGatewayKey(ctx, keyHash)
	if err != nil {
		t.Fatalf("InvalidateGatewayKey failed: %v", err)
	}

	// Get key (not found after invalidation)
	_, found, err = c.GetGatewayKey(ctx, keyHash)
	if err != nil {
		t.Fatalf("GetGatewayKey failed: %v", err)
	}
	if found {
		t.Errorf("Expected key to be deleted after invalidation")
	}

	// 2. Test Team Budget Cache & Atomic Incrementation
	teamID := "team-abc"
	limit := 100.00
	used := 5.00

	// Get budget (not found)
	_, _, found, err = c.GetTeamBudget(ctx, teamID)
	if err != nil {
		t.Fatalf("GetTeamBudget failed: %v", err)
	}
	if found {
		t.Errorf("Expected team budget to not be found initially")
	}

	// Set budget
	err = c.SetTeamBudget(ctx, teamID, limit, used)
	if err != nil {
		t.Fatalf("SetTeamBudget failed: %v", err)
	}

	// Get budget (found)
	gotLimit, gotUsed, found, err := c.GetTeamBudget(ctx, teamID)
	if err != nil {
		t.Fatalf("GetTeamBudget failed: %v", err)
	}
	if !found {
		t.Errorf("Expected team budget to be found")
	}
	if gotLimit != limit || gotUsed != used {
		t.Errorf("Expected limit %f and used %f, got limit %f and used %f", limit, used, gotLimit, gotUsed)
	}

	// Increment budget used
	increment := 2.50
	newUsed, err := c.IncrementTeamBudgetUsed(ctx, teamID, increment)
	if err != nil {
		t.Fatalf("IncrementTeamBudgetUsed failed: %v", err)
	}
	if newUsed != 7.50 {
		t.Errorf("Expected new used to be 7.50, got %f", newUsed)
	}

	// Verify budget in cache after increment
	_, gotUsed, _, _ = c.GetTeamBudget(ctx, teamID)
	if gotUsed != 7.50 {
		t.Errorf("Expected cached used to be 7.50, got %f", gotUsed)
	}

	// Invalidate Team
	err = c.InvalidateTeam(ctx, teamID)
	if err != nil {
		t.Fatalf("InvalidateTeam failed: %v", err)
	}

	// Verify not found after invalidation
	_, _, found, _ = c.GetTeamBudget(ctx, teamID)
	if found {
		t.Errorf("Expected team budget to be invalidated")
	}
}
