package service

import (
	"context"
	"orynt/internal/models"
	"orynt/internal/repository"
	"testing"
)

func TestStadiumService_GetStadiumDetails(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewStadiumService(dbRepo, pubSubRepo)

	stadium, err := svc.GetStadiumDetails(ctx, "stadium_01")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if stadium == nil {
		t.Fatal("Expected stadium details to be not nil")
	}

	if stadium.Name != "Orynt Arena" {
		t.Errorf("Expected stadium name 'Orynt Arena', got %s", stadium.Name)
	}
}

func TestStadiumService_CrowdZonesAndOccupancy(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewStadiumService(dbRepo, pubSubRepo)

	zones, err := svc.GetCrowdZones(ctx, "stadium_01")
	if err != nil {
		t.Fatalf("Expected no error listing crowd zones, got %v", err)
	}

	if len(zones) == 0 {
		t.Fatal("Expected crowd zones to be seeded, got 0")
	}

	// Update occupancy of z1 (North Stand) to 9000
	// Max capacity of z1 is 10000 in seedData.
	// 9000/10000 = 90% which is >= 80%, so DensityLevel should be "high".
	err = svc.UpdateZoneOccupancy(ctx, "z1", 9000)
	if err != nil {
		t.Fatalf("Failed to update zone occupancy: %v", err)
	}

	// Fetch again to verify
	updatedZones, _ := svc.GetCrowdZones(ctx, "stadium_01")
	var found bool
	for _, z := range updatedZones {
		if z.ID == "z1" {
			found = true
			if z.CurrentOccupancy != 9000 {
				t.Errorf("Expected occupancy 9000, got %d", z.CurrentOccupancy)
			}
			if z.DensityLevel != "high" {
				t.Errorf("Expected density level 'high', got '%s'", z.DensityLevel)
			}
		}
	}
	if !found {
		t.Error("Zone z1 not found in list")
	}

	// Update to 2000 (20%) -> should become low
	err = svc.UpdateZoneOccupancy(ctx, "z1", 2000)
	if err != nil {
		t.Fatalf("Failed to update zone occupancy: %v", err)
	}
	updatedZones, _ = svc.GetCrowdZones(ctx, "stadium_01")
	for _, z := range updatedZones {
		if z.ID == "z1" {
			if z.DensityLevel != "low" {
				t.Errorf("Expected density level 'low', got '%s'", z.DensityLevel)
			}
		}
	}
}

func TestStadiumService_ParkingAndTransport(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewStadiumService(dbRepo, pubSubRepo)

	// Parking (seeded with p1, p2, p3, p4)
	lots, err := svc.GetParkingZones(ctx)
	if err != nil {
		t.Fatalf("Failed to get parking zones: %v", err)
	}
	if len(lots) == 0 {
		t.Fatal("Expected parking zones, got 0")
	}

	err = svc.UpdateParking(ctx, "p1", 480)
	if err != nil {
		t.Fatalf("Failed to update parking spots: %v", err)
	}

	lots, _ = svc.GetParkingZones(ctx)
	for _, l := range lots {
		if l.ID == "p1" {
			if l.OccupiedSpots != 480 {
				t.Errorf("Expected 480 occupied spots, got %d", l.OccupiedSpots)
			}
		}
	}

	// Transport (seeded with t_line_1, t_line_2, t_line_3)
	lines, err := svc.GetTransportLines(ctx)
	if err != nil {
		t.Fatalf("Failed to get transport lines: %v", err)
	}
	if len(lines) == 0 {
		t.Fatal("Expected transport lines, got 0")
	}

	err = svc.UpdateTransport(ctx, "t_line_1", 15, "Delayed")
	if err != nil {
		t.Fatalf("Failed to update transport: %v", err)
	}

	lines, _ = svc.GetTransportLines(ctx)
	for _, r := range lines {
		if r.ID == "t_line_1" {
			if r.DelayMinutes != 15 || r.Status != "Delayed" {
				t.Errorf("Expected delay 15 and status Delayed, got delay=%d status=%s", r.DelayMinutes, r.Status)
			}
		}
	}
}

func TestStadiumService_Alerts(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewStadiumService(dbRepo, pubSubRepo)

	alert, err := svc.BroadcastAlert(ctx, "Severe Storm", "Seek shelter immediately", "weather", "critical")
	if err != nil {
		t.Fatalf("Failed to broadcast alert: %v", err)
	}

	if alert.Title != "Severe Storm" || !alert.Active {
		t.Errorf("Alert not initialized correctly: %+v", alert)
	}

	alerts, err := svc.GetAlerts(ctx, true)
	if err != nil {
		t.Fatalf("Failed to get active alerts: %v", err)
	}

	var found bool
	for _, a := range alerts {
		if a.ID == alert.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Broadcast alert was not found in active alerts list")
	}

	err = svc.ResolveAlert(ctx, alert.ID)
	if err != nil {
		t.Fatalf("Failed to resolve alert: %v", err)
	}

	alerts, _ = svc.GetAlerts(ctx, true)
	for _, a := range alerts {
		if a.ID == alert.ID {
			t.Error("Resolved alert should not show up in active-only list")
		}
	}
}

func TestStadiumService_GetExitRoutes(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewStadiumService(dbRepo, pubSubRepo)

	// Step 1: Initial seeded state has z1 (North Stand) in high density.
	// So z1 routes to e2. z5 (VIP Concourse) has z5 high & z1 high, so it also routes to e2.
	routes, err := svc.GetExitRoutes(ctx)
	if err != nil {
		t.Fatalf("Failed to get exit routes: %v", err)
	}

	var z1Route models.ExitRoute
	var z5Route models.ExitRoute
	for _, r := range routes {
		if r.ZoneID == "z1" {
			z1Route = r
		}
		if r.ZoneID == "z5" {
			z5Route = r
		}
	}

	if z1Route.RecommendedExit != "e2" {
		t.Errorf("Expected initial z1 exit to route to e2 (due to seeded high density), got %s", z1Route.RecommendedExit)
	}
	if z5Route.RecommendedExit != "e2" {
		t.Errorf("Expected initial z5 exit to route to e2, got %s", z5Route.RecommendedExit)
	}

	// Step 2: Update z1 to 2000 (low density, ratio 20%)
	err = svc.UpdateZoneOccupancy(ctx, "z1", 2000)
	if err != nil {
		t.Fatalf("Failed to update z1 occupancy: %v", err)
	}

	// Get routes again to check rerouting
	routes, err = svc.GetExitRoutes(ctx)
	if err != nil {
		t.Fatalf("Failed to get exit routes: %v", err)
	}

	for _, r := range routes {
		if r.ZoneID == "z1" {
			// Under low density, z1 routes to its natural exit e1
			if r.RecommendedExit != "e1" {
				t.Errorf("Expected low-density z1 exit to reroute to e1, got %s", r.RecommendedExit)
			}
		}
		if r.ZoneID == "z5" {
			// Since z1 is now low density, z5 routes to e1
			if r.RecommendedExit != "e1" {
				t.Errorf("Expected z5 exit to revert to e1 since z1 is clear, got %s", r.RecommendedExit)
			}
		}
	}
}
