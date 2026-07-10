package storage

import (
	"context"
	"database/sql"
	"gateway/internal/model"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestPostgresStorageIntegration(t *testing.T) {
	// Read DATABASE_URL from system environment or use default port 5435
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Use port 5435 as configured in our compose/env
		dbURL = "postgres://postgres:postgres@localhost:5435/gateway?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("Skipping integration test: failed to open postgres connection: %v", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test: database not reachable: %v", err)
		return
	}

	s, err := NewPostgresStorage(db)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	ctx := context.Background()

	// 1. Test Create Team
	testTeam := &model.Team{
		ID:          "test-integration-team-id",
		Name:        "Test Integration Team",
		BudgetLimit: 500.00,
		BudgetUsed:  0.0,
	}

	// Delete team first if leftover
	_, _ = db.Exec("DELETE FROM teams WHERE id = $1", testTeam.ID)

	err = s.CreateTeam(ctx, testTeam)
	if err != nil {
		t.Fatalf("CreateTeam failed: %v", err)
	}

	// 2. Test Get Team
	gotTeam, err := s.GetTeam(ctx, testTeam.ID)
	if err != nil {
		t.Fatalf("GetTeam failed: %v", err)
	}
	if gotTeam == nil {
		t.Fatalf("Expected team to be found, got nil")
	}
	if gotTeam.Name != testTeam.Name || gotTeam.BudgetLimit != testTeam.BudgetLimit {
		t.Errorf("GetTeam data mismatch: %+v", gotTeam)
	}

	// 3. Test List Teams
	teams, err := s.ListTeams(ctx)
	if err != nil {
		t.Fatalf("ListTeams failed: %v", err)
	}
	found := false
	for _, team := range teams {
		if team.ID == testTeam.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected test team to be included in ListTeams result")
	}

	// 4. Test Update Budget
	err = s.UpdateTeamBudget(ctx, testTeam.ID, 50.00)
	if err != nil {
		t.Fatalf("UpdateTeamBudget failed: %v", err)
	}

	gotTeam, _ = s.GetTeam(ctx, testTeam.ID)
	if gotTeam.BudgetUsed != 50.00 {
		t.Errorf("Expected budget used to be 50.00, got %f", gotTeam.BudgetUsed)
	}

	// Cleanup
	_, err = db.Exec("DELETE FROM teams WHERE id = $1", testTeam.ID)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}
