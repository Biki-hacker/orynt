package repository

import (
	"context"
	"orynt/internal/models"
	"time"
)

// DBRepository defines storage APIs
// ... (rest of file remains same)
type DBRepository interface {
	// Users
	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	ListUsers(ctx context.Context) ([]models.User, error)

	// Tournaments & Matches
	GetTournament(ctx context.Context, id string) (*models.Tournament, error)
	ListTournaments(ctx context.Context) ([]models.Tournament, error)
	CreateTournament(ctx context.Context, tournament *models.Tournament) error
	GetMatch(ctx context.Context, id string) (*models.Match, error)
	ListMatches(ctx context.Context, tournamentID string) ([]models.Match, error)
	CreateMatch(ctx context.Context, match *models.Match) error
	UpdateMatch(ctx context.Context, match *models.Match) error

	// Stadium & Crowd Zones
	GetStadium(ctx context.Context, id string) (*models.Stadium, error)
	CreateStadium(ctx context.Context, stadium *models.Stadium) error
	ListCrowdZones(ctx context.Context, stadiumID string) ([]models.CrowdZone, error)
	UpdateCrowdZoneOccupancy(ctx context.Context, zoneID string, occupancy int) error
	CreateCrowdZone(ctx context.Context, zone *models.CrowdZone) error

	// Transport & Parking
	ListParkingZones(ctx context.Context) ([]models.Parking, error)
	UpdateParkingSpots(ctx context.Context, lotID string, occupied int) error
	CreateParkingZone(ctx context.Context, lot *models.Parking) error
	ListTransportLines(ctx context.Context) ([]models.Transport, error)
	UpdateTransportArrival(ctx context.Context, routeID string, delay int, status string) error
	CreateTransportLine(ctx context.Context, route *models.Transport) error

	// Alerts & Announcements
	ListAlerts(ctx context.Context, activeOnly bool) ([]models.Alert, error)
	CreateAlert(ctx context.Context, alert *models.Alert) error
	UpdateAlertStatus(ctx context.Context, id string, active bool) error
	ListAnnouncements(ctx context.Context, role string) ([]models.Announcement, error)
	CreateAnnouncement(ctx context.Context, announcement *models.Announcement) error
	ApproveAnnouncement(ctx context.Context, id string, adminUsername string) error

	// Tasks & Operational Workflows
	ListTasks(ctx context.Context, department string) ([]models.Task, error)
	CreateTask(ctx context.Context, task *models.Task) error
	UpdateTaskStatus(ctx context.Context, id string, status string) error
	ListLostFound(ctx context.Context) ([]models.LostFoundItem, error)
	CreateLostFound(ctx context.Context, item *models.LostFoundItem) error
	UpdateLostFoundStatus(ctx context.Context, id string, status string) error
	ListFoodStalls(ctx context.Context) ([]models.FoodStall, error)
	CreateFoodStall(ctx context.Context, stall *models.FoodStall) error
	UpdateFoodStallWaitTime(ctx context.Context, id string, minutes int) error

	// Medical
	ListMedicalRequests(ctx context.Context) ([]models.MedicalRequest, error)
	CreateMedicalRequest(ctx context.Context, req *models.MedicalRequest) error
	UpdateMedicalRequestStatus(ctx context.Context, id string, status string, assignedTo string) error

	// Audit Logs & Settings
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
	ListAuditLogs(ctx context.Context) ([]models.AuditLog, error)
}

// PubSubRepository handles real-time message broadcasting triggers (Redis PubSub equivalent)
type PubSubRepository interface {
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channel string, handler func(msg string)) error
	IncrLimit(ctx context.Context, key string, expiration time.Duration) (int64, error)
}
