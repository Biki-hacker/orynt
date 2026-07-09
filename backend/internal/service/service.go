package service

import (
	"context"
	"orynt/internal/models"
)

// AuthService handles authentication/authorization workflows
type AuthService interface {
	Login(ctx context.Context, username, password string) (*models.TokenResponse, error)
	ValidateToken(tokenStr string) (*models.User, error)
	CreateUser(ctx context.Context, username, password, role, department string) (*models.User, error)
	ListUsers(ctx context.Context) ([]models.User, error)
}

// TournamentService manages tournament structure and live matches
type TournamentService interface {
	CreateTournament(ctx context.Context, name string) (*models.Tournament, error)
	ListTournaments(ctx context.Context) ([]models.Tournament, error)
	ListMatches(ctx context.Context, tournamentID string) ([]models.Match, error)
	CreateMatch(ctx context.Context, tournamentID, homeTeam, awayTeam string, scheduledTime string) (*models.Match, error)
	UpdateMatchScore(ctx context.Context, matchID string, homeScore, awayScore int, elapsed string) (*models.Match, error)
	AddMatchEvent(ctx context.Context, matchID string, time, eventType, detail string) (*models.Match, error)
}

// StadiumService handles stadium settings and live telemetry
type StadiumService interface {
	GetStadiumDetails(ctx context.Context, id string) (*models.Stadium, error)
	GetCrowdZones(ctx context.Context, stadiumID string) ([]models.CrowdZone, error)
	UpdateZoneOccupancy(ctx context.Context, zoneID string, currentOccupancy int) error
	GetParkingZones(ctx context.Context) ([]models.Parking, error)
	UpdateParking(ctx context.Context, lotID string, occupied int) error
	GetTransportLines(ctx context.Context) ([]models.Transport, error)
	UpdateTransport(ctx context.Context, routeID string, delay int, status string) error
	BroadcastAlert(ctx context.Context, title, content, alertType, severity string) (*models.Alert, error)
	GetAlerts(ctx context.Context, activeOnly bool) ([]models.Alert, error)
	ResolveAlert(ctx context.Context, id string) error
	GetExitRoutes(ctx context.Context) ([]models.ExitRoute, error)
}

// OperationsService manages internal staff workloads and public utilities
type OperationsService interface {
	ListTasks(ctx context.Context, department string) ([]models.Task, error)
	CreateTask(ctx context.Context, title, description, assignedTo, priority, department string) (*models.Task, error)
	UpdateTaskStatus(ctx context.Context, taskID, status string) error

	ListLostFound(ctx context.Context) ([]models.LostFoundItem, error)
	ReportLostFoundItem(ctx context.Context, name, description, category, contactName, contactPhone string) (*models.LostFoundItem, error)
	ClaimLostFoundItem(ctx context.Context, itemID string) error

	ListFoodStalls(ctx context.Context) ([]models.FoodStall, error)
	UpdateFoodStallWaitTime(ctx context.Context, stallID string, waitTime int) error

	CreateAnnouncement(ctx context.Context, title, content, targetAudience string) (*models.Announcement, error)
	ApproveAnnouncement(ctx context.Context, id string, adminUsername string) error
	ListAnnouncements(ctx context.Context, role string) ([]models.Announcement, error)

	ListMedicalRequests(ctx context.Context) ([]models.MedicalRequest, error)
	CreateMedicalRequest(ctx context.Context, requester, location, description string) (*models.MedicalRequest, error)
	AssignMedicalRequest(ctx context.Context, id string, assignedTo string) error
	ResolveMedicalRequest(ctx context.Context, id string) error

	ListAuditLogs(ctx context.Context) ([]models.AuditLog, error)
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
}

// AIService processes natural language queries grounded in Firestore data (RAG)
type AIService interface {
	Chat(ctx context.Context, req *models.AIChatRequest) (*models.AIChatResponse, error)
	ChatStream(ctx context.Context, req *models.AIChatRequest, onChunk func(chunk *models.AIChatResponse) error) error
	GetSustainabilityRecommendations(ctx context.Context, energyUsage float64, waterConsumption float64, wasteGenerated float64, attendance int) (string, error)
}
