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

// GetExitRoutes evaluates crowd densities and suggests optimal exit pathways
func (s *stadiumService) GetExitRoutes(ctx context.Context) ([]models.ExitRoute, error) {
	zones, err := s.dbRepo.ListCrowdZones(ctx, "stadium_01")
	if err != nil {
		return nil, err
	}

	routes := make([]models.ExitRoute, 0, len(zones))

	// Find the density of each zone for route calculations
	zoneDensities := make(map[string]string)
	for _, z := range zones {
		zoneDensities[z.ID] = z.DensityLevel
	}

	for _, z := range zones {
		var recommendedExit string
		var exitName string
		var congestionLevel string
		var pathPoints []models.Point
		var reason string

		switch z.ID {
		case "z1": // North Stand (natural exit is e1 - Exit Gate A)
			if zoneDensities["z1"] == "high" {
				recommendedExit = "e2"
				exitName = "Exit Gate B (Accessible)"
				congestionLevel = "medium"
				pathPoints = []models.Point{
					{X: 250, Y: 80},
					{X: 100, Y: 80},
					{X: 80, Y: 180},
					{X: 80, Y: 350},
				}
				reason = "Exit Gate A concourse is heavily congested. Redirecting via West Concourse to Exit Gate B to prevent crowd crush."
			} else {
				recommendedExit = "e1"
				exitName = "Exit Gate A"
				congestionLevel = "low"
				pathPoints = []models.Point{
					{X: 250, Y: 80},
					{X: 300, Y: 50},
					{X: 300, Y: 15},
				}
				reason = "Route clear. Proceed to Exit Gate A (North Side)."
			}

		case "z2": // East Stand (natural exit is e1 - Exit Gate A)
			recommendedExit = "e1"
			exitName = "Exit Gate A"
			congestionLevel = "low"
			if zoneDensities["z1"] == "high" {
				congestionLevel = "medium"
				reason = "Concourse A experiencing high volume due to North Stand egress. Proceed with caution to Exit Gate A."
			} else {
				reason = "Route clear. Proceed to Exit Gate A (North Side)."
			}
			pathPoints = []models.Point{
				{X: 420, Y: 200},
				{X: 420, Y: 100},
				{X: 300, Y: 50},
				{X: 300, Y: 15},
			}

		case "z3": // South Stand (natural exit is e2 - Exit Gate B)
			recommendedExit = "e2"
			exitName = "Exit Gate B (Accessible)"
			congestionLevel = "low"
			if zoneDensities["z1"] == "high" {
				congestionLevel = "medium"
				reason = "Exit Gate B experiencing increased flow from North Stand rerouting. Expect minor queues at Exit Gate B."
			} else {
				reason = "Route clear. Proceed to Exit Gate B (South-West Side)."
			}
			pathPoints = []models.Point{
				{X: 250, Y: 300},
				{X: 150, Y: 320},
				{X: 80, Y: 350},
			}

		case "z4": // West Stand (natural exit is e2 - Exit Gate B)
			recommendedExit = "e2"
			exitName = "Exit Gate B (Accessible)"
			congestionLevel = "low"
			if zoneDensities["z1"] == "high" {
				congestionLevel = "medium"
				reason = "Exit Gate B experiencing increased flow from North Stand rerouting. Expect minor queues at Exit Gate B."
			} else {
				reason = "Route clear. Proceed to Exit Gate B (South-West Side)."
			}
			pathPoints = []models.Point{
				{X: 80, Y: 200},
				{X: 80, Y: 350},
			}

		case "z5": // VIP Concourse (natural exit is e1 - Exit Gate A)
			if zoneDensities["z5"] == "high" && zoneDensities["z1"] == "high" {
				recommendedExit = "e2"
				exitName = "Exit Gate B (Accessible)"
				congestionLevel = "medium"
				pathPoints = []models.Point{
					{X: 250, Y: 240},
					{X: 180, Y: 280},
					{X: 80, Y: 350},
				}
				reason = "Exit Gate A route heavily loaded. VIP exit path diverted to Exit Gate B."
			} else {
				recommendedExit = "e1"
				exitName = "Exit Gate A"
				congestionLevel = "low"
				pathPoints = []models.Point{
					{X: 250, Y: 160},
					{X: 300, Y: 120},
					{X: 300, Y: 15},
				}
				reason = "Route clear. Proceed to Exit Gate A."
			}
		}

		routes = append(routes, models.ExitRoute{
			ZoneID:          z.ID,
			ZoneName:        z.ZoneName,
			RecommendedExit: recommendedExit,
			ExitName:        exitName,
			CongestionLevel: congestionLevel,
			PathPoints:      pathPoints,
			Reason:          reason,
		})
	}

	return routes, nil
}

