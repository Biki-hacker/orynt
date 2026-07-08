package service

import (
	"context"
	"fmt"
	"time"
	"orynt/internal/models"
	"orynt/internal/repository"
)

type tournamentService struct {
	dbRepo     repository.DBRepository
	pubSubRepo repository.PubSubRepository
}

// NewTournamentService creates concrete tournament service
func NewTournamentService(dbRepo repository.DBRepository, pubSubRepo repository.PubSubRepository) TournamentService {
	return &tournamentService{
		dbRepo:     dbRepo,
		pubSubRepo: pubSubRepo,
	}
}

func (s *tournamentService) CreateTournament(ctx context.Context, name string) (*models.Tournament, error) {
	t := &models.Tournament{
		ID:        fmt.Sprintf("t_%d", time.Now().UnixNano()),
		Name:      name,
		Status:    "planned",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(7 * 24 * time.Hour),
	}
	err := s.dbRepo.CreateTournament(ctx, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *tournamentService) ListTournaments(ctx context.Context) ([]models.Tournament, error) {
	return s.dbRepo.ListTournaments(ctx)
}

func (s *tournamentService) ListMatches(ctx context.Context, tournamentID string) ([]models.Match, error) {
	return s.dbRepo.ListMatches(ctx, tournamentID)
}

func (s *tournamentService) CreateMatch(ctx context.Context, tournamentID, homeTeam, awayTeam string, scheduledTime string) (*models.Match, error) {
	parsedTime, err := time.Parse(time.RFC3339, scheduledTime)
	if err != nil {
		parsedTime = time.Now().Add(2 * time.Hour) // fallback
	}

	m := &models.Match{
		ID:           fmt.Sprintf("m_%d", time.Now().UnixNano()),
		TournamentID: tournamentID,
		HomeTeam:     homeTeam,
		AwayTeam:     awayTeam,
		HomeScore:    0,
		AwayScore:    0,
		Status:       "scheduled",
		ScheduledAt:  parsedTime,
		TimeElapsed:  "00:00",
		Events:       make([]models.MatchComment, 0),
	}

	err = s.dbRepo.CreateMatch(ctx, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *tournamentService) UpdateMatchScore(ctx context.Context, matchID string, homeScore, awayScore int, elapsed string) (*models.Match, error) {
	m, err := s.dbRepo.GetMatch(ctx, matchID)
	if err != nil {
		return nil, err
	}

	m.HomeScore = homeScore
	m.AwayScore = awayScore
	m.TimeElapsed = elapsed
	if elapsed == "FT" || elapsed == "90:00" {
		m.Status = "completed"
	} else if elapsed != "00:00" {
		m.Status = "live"
	}

	err = s.dbRepo.UpdateMatch(ctx, m)
	if err != nil {
		return nil, err
	}

	// Publish to match updates channel
	msg := models.WSMessage{
		Type:    "match_score",
		Payload: m,
	}
	_ = s.pubSubRepo.Publish(ctx, "matches", msg)

	return m, nil
}

func (s *tournamentService) AddMatchEvent(ctx context.Context, matchID string, timeVal, eventType, detail string) (*models.Match, error) {
	m, err := s.dbRepo.GetMatch(ctx, matchID)
	if err != nil {
		return nil, err
	}

	event := models.MatchComment{
		Time:   timeVal,
		Type:   eventType,
		Detail: detail,
	}

	m.Events = append(m.Events, event)
	err = s.dbRepo.UpdateMatch(ctx, m)
	if err != nil {
		return nil, err
	}

	// Publish match updates
	msg := models.WSMessage{
		Type:    "match_event",
		Payload: m,
	}
	_ = s.pubSubRepo.Publish(ctx, "matches", msg)

	return m, nil
}
