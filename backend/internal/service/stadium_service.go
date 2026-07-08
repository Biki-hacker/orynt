package service

import (
	"context"
	"fmt"
	"time"
	"orynt/internal/models"
	"orynt/internal/repository"
)

type stadiumService struct {
	dbRepo     repository.DBRepository
	pubSubRepo repository.PubSubRepository
}

// NewStadiumService creates concrete stadium service
func NewStadiumService(dbRepo repository.DBRepository, pubSubRepo repository.PubSubRepository) StadiumService {
	return &stadiumService{
		dbRepo:     dbRepo,
		pubSubRepo: pubSubRepo,
	}
}

func (s *stadiumService) GetStadiumDetails(ctx context.Context, id string) (*models.Stadium, error) {
	return s.dbRepo.GetStadium(ctx, id)
}

func (s *stadiumService) GetCrowdZones(ctx context.Context, stadiumID string) ([]models.CrowdZone, error) {
	return s.dbRepo.ListCrowdZones(ctx, stadiumID)
}

func (s *stadiumService) UpdateZoneOccupancy(ctx context.Context, zoneID string, currentOccupancy int) error {
	err := s.dbRepo.UpdateCrowdZoneOccupancy(ctx, zoneID, currentOccupancy)
	if err != nil {
		return err
	}

	// Fetch updated zone details to broadcast
	zones, err := s.dbRepo.ListCrowdZones(ctx, "")
	if err == nil {
		var updatedZone models.CrowdZone
		for _, z := range zones {
			if z.ID == zoneID {
				updatedZone = z
				break
			}
		}
		// Publish event
		msg := models.WSMessage{
			Type:    "crowd_density",
			Payload: updatedZone,
		}
		_ = s.pubSubRepo.Publish(ctx, "stadium", msg)
	}

	return nil
}

func (s *stadiumService) GetParkingZones(ctx context.Context) ([]models.Parking, error) {
	return s.dbRepo.ListParkingZones(ctx)
}

func (s *stadiumService) UpdateParking(ctx context.Context, lotID string, occupied int) error {
	err := s.dbRepo.UpdateParkingSpots(ctx, lotID, occupied)
	if err != nil {
		return err
	}

	lots, err := s.dbRepo.ListParkingZones(ctx)
	if err == nil {
		var updatedLot models.Parking
		for _, l := range lots {
			if l.ID == lotID {
				updatedLot = l
				break
			}
		}
		msg := models.WSMessage{
			Type:    "parking",
			Payload: updatedLot,
		}
		_ = s.pubSubRepo.Publish(ctx, "parking", msg)
	}

	return nil
}

func (s *stadiumService) GetTransportLines(ctx context.Context) ([]models.Transport, error) {
	return s.dbRepo.ListTransportLines(ctx)
}

func (s *stadiumService) UpdateTransport(ctx context.Context, routeID string, delay int, status string) error {
	err := s.dbRepo.UpdateTransportArrival(ctx, routeID, delay, status)
	if err != nil {
		return err
	}

	lines, err := s.dbRepo.ListTransportLines(ctx)
	if err == nil {
		var updatedRoute models.Transport
		for _, r := range lines {
			if r.ID == routeID {
				updatedRoute = r
				break
			}
		}
		msg := models.WSMessage{
			Type:    "transport",
			Payload: updatedRoute,
		}
		_ = s.pubSubRepo.Publish(ctx, "transport", msg)
	}
	return nil
}

func (s *stadiumService) BroadcastAlert(ctx context.Context, title, content, alertType, severity string) (*models.Alert, error) {
	alert := &models.Alert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Title:     title,
		Content:   content,
		Type:      alertType,
		Severity:  severity,
		CreatedAt: time.Now(),
		Active:    true,
	}

	err := s.dbRepo.CreateAlert(ctx, alert)
	if err != nil {
		return nil, err
	}

	// Broadcast alert via PubSub
	msg := models.WSMessage{
		Type:    "alert",
		Payload: alert,
	}
	_ = s.pubSubRepo.Publish(ctx, "alerts", msg)

	return alert, nil
}

func (s *stadiumService) GetAlerts(ctx context.Context, activeOnly bool) ([]models.Alert, error) {
	return s.dbRepo.ListAlerts(ctx, activeOnly)
}

func (s *stadiumService) ResolveAlert(ctx context.Context, id string) error {
	err := s.dbRepo.UpdateAlertStatus(ctx, id, false)
	if err != nil {
		return err
	}

	// Notify resolver
	msg := models.WSMessage{
		Type: "alert_resolved",
		Payload: map[string]string{
			"id": id,
		},
	}
	_ = s.pubSubRepo.Publish(ctx, "alerts", msg)

	return nil
}
