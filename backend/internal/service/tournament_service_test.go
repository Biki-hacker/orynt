package service

import (
	"context"
	"orynt/internal/repository"
	"testing"
	"time"
)

func TestTournamentService_CreateAndListTournaments(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewTournamentService(dbRepo, pubSubRepo)

	tournaments, err := svc.ListTournaments(ctx)
	if err != nil {
		t.Fatalf("Failed to list tournaments: %v", err)
	}
	initialCount := len(tournaments)

	tourn, err := svc.CreateTournament(ctx, "Copa America 2026")
	if err != nil {
		t.Fatalf("Failed to create tournament: %v", err)
	}

	if tourn.Name != "Copa America 2026" || tourn.Status != "planned" {
		t.Errorf("Unexpected tournament metadata: %+v", tourn)
	}

	tournaments, err = svc.ListTournaments(ctx)
	if err != nil {
		t.Fatalf("Failed to list tournaments: %v", err)
	}

	if len(tournaments) != initialCount+1 {
		t.Errorf("Expected tournament list size to be %d, got %d", initialCount+1, len(tournaments))
	}
}

func TestTournamentService_Matches(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewTournamentService(dbRepo, pubSubRepo)

	// Create a tournament first
	tourn, err := svc.CreateTournament(ctx, "World Cup 2026")
	if err != nil {
		t.Fatalf("Failed to create tournament: %v", err)
	}

	// Create Match
	nowStr := time.Now().Format(time.RFC3339)
	match, err := svc.CreateMatch(ctx, tourn.ID, "USA", "Mexico", nowStr)
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	if match.HomeTeam != "USA" || match.AwayTeam != "Mexico" || match.Status != "scheduled" {
		t.Errorf("Unexpected match status/teams: %+v", match)
	}

	// Update Score -> Live
	updated, err := svc.UpdateMatchScore(ctx, match.ID, 1, 0, "12:34")
	if err != nil {
		t.Fatalf("Failed to update score: %v", err)
	}
	if updated.Status != "live" || updated.HomeScore != 1 || updated.AwayScore != 0 {
		t.Errorf("Expected match status 'live' and score 1-0, got status=%s score=%d-%d", updated.Status, updated.HomeScore, updated.AwayScore)
	}

	// Add Match Event
	updated, err = svc.AddMatchEvent(ctx, match.ID, "12'", "goal", "Christian Pulisic (USA) scores!")
	if err != nil {
		t.Fatalf("Failed to add match event: %v", err)
	}

	if len(updated.Events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(updated.Events))
	}
	if updated.Events[0].Type != "goal" || updated.Events[0].Detail != "Christian Pulisic (USA) scores!" {
		t.Errorf("Unexpected event detail: %+v", updated.Events[0])
	}

	// Update Score -> Completed (using "FT")
	updated, err = svc.UpdateMatchScore(ctx, match.ID, 2, 1, "FT")
	if err != nil {
		t.Fatalf("Failed to complete match: %v", err)
	}
	if updated.Status != "completed" || updated.TimeElapsed != "FT" {
		t.Errorf("Expected completed match, got status=%s, elapsed=%s", updated.Status, updated.TimeElapsed)
	}

	// List matches
	matches, err := svc.ListMatches(ctx, tourn.ID)
	if err != nil {
		t.Fatalf("Failed to list matches: %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(matches))
	}
}
