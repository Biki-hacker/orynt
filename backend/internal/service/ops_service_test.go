package service

import (
	"context"
	"orynt/internal/models"
	"orynt/internal/repository"
	"testing"
	"time"
)

func TestOpsService_Tasks(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewOperationsService(dbRepo, pubSubRepo)

	tasks, err := svc.ListTasks(ctx, "Facilities")
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}
	initialCount := len(tasks)

	task, err := svc.CreateTask(ctx, "Clean Spill", "Spill near Stall A", "volunteer1", "high", "Facilities")
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if task.Title != "Clean Spill" || task.Status != "todo" {
		t.Errorf("Unexpected task status/details: %+v", task)
	}

	tasks, _ = svc.ListTasks(ctx, "Facilities")
	if len(tasks) != initialCount+1 {
		t.Errorf("Expected tasks count to be %d, got %d", initialCount+1, len(tasks))
	}

	err = svc.UpdateTaskStatus(ctx, task.ID, "in_progress")
	if err != nil {
		t.Fatalf("Failed to update task status: %v", err)
	}

	tasks, _ = svc.ListTasks(ctx, "")
	for _, tk := range tasks {
		if tk.ID == task.ID {
			if tk.Status != "in_progress" {
				t.Errorf("Expected status to be in_progress, got %s", tk.Status)
			}
		}
	}
}

func TestOpsService_LostFound(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewOperationsService(dbRepo, pubSubRepo)

	items, err := svc.ListLostFound(ctx)
	if err != nil {
		t.Fatalf("Failed to list lost/found: %v", err)
	}
	initialCount := len(items)

	item, err := svc.ReportLostFoundItem(ctx, "iPhone 13", "Black cover, cracked screen", "Electronics", "John Doe", "123-456")
	if err != nil {
		t.Fatalf("Failed to report item: %v", err)
	}

	if item.ItemName != "iPhone 13" || item.Status != "reported_lost" {
		t.Errorf("Unexpected item state: %+v", item)
	}

	items, _ = svc.ListLostFound(ctx)
	if len(items) != initialCount+1 {
		t.Errorf("Expected items count to be %d, got %d", initialCount+1, len(items))
	}

	err = svc.ClaimLostFoundItem(ctx, item.ID)
	if err != nil {
		t.Fatalf("Failed to claim item: %v", err)
	}

	items, _ = svc.ListLostFound(ctx)
	for _, itm := range items {
		if itm.ID == item.ID {
			if itm.Status != "claimed" {
				t.Errorf("Expected status 'claimed', got %s", itm.Status)
			}
		}
	}
}

func TestOpsService_FoodStalls(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewOperationsService(dbRepo, pubSubRepo)

	stalls, err := svc.ListFoodStalls(ctx)
	if err != nil {
		t.Fatalf("Failed to list food stalls: %v", err)
	}
	if len(stalls) == 0 {
		t.Fatal("Expected seeded food stalls, got 0")
	}

	stallID := stalls[0].ID
	err = svc.UpdateFoodStallWaitTime(ctx, stallID, 25)
	if err != nil {
		t.Fatalf("Failed to update wait time: %v", err)
	}

	stalls, _ = svc.ListFoodStalls(ctx)
	for _, fs := range stalls {
		if fs.ID == stallID {
			if fs.WaitTimeMinutes != 25 {
				t.Errorf("Expected wait time 25, got %d", fs.WaitTimeMinutes)
			}
		}
	}
}

func TestOpsService_Announcements(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewOperationsService(dbRepo, pubSubRepo)

	ann, err := svc.CreateAnnouncement(ctx, "Gate B Closed", "Use Gate C instead", "public")
	if err != nil {
		t.Fatalf("Failed to create announcement: %v", err)
	}

	if ann.Approved != false {
		t.Error("New announcement should be unapproved by default")
	}

	// Should not list in public announcements since it's not approved
	anns, _ := svc.ListAnnouncements(ctx, "public")
	for _, a := range anns {
		if a.ID == ann.ID {
			t.Error("Unapproved announcement should not show in public list")
		}
	}

	// Approve
	err = svc.ApproveAnnouncement(ctx, ann.ID, "admin")
	if err != nil {
		t.Fatalf("Failed to approve announcement: %v", err)
	}

	// Now it should list in public
	anns, _ = svc.ListAnnouncements(ctx, "public")
	var found bool
	for _, a := range anns {
		if a.ID == ann.ID {
			found = true
			if a.ApprovedBy != "admin" || !a.Approved {
				t.Errorf("Expected approved by admin, got approved=%t, approvedBy=%s", a.Approved, a.ApprovedBy)
			}
		}
	}
	if !found {
		t.Error("Approved announcement not found in public list")
	}
}

func TestOpsService_MedicalRequests(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewOperationsService(dbRepo, pubSubRepo)

	req, err := svc.CreateMedicalRequest(ctx, "Jane Smith", "Section 104", "Heat stroke symptoms")
	if err != nil {
		t.Fatalf("Failed to create medical request: %v", err)
	}

	if req.Status != "pending" || req.AssignedTo != "" {
		t.Errorf("Unexpected medical request state: %+v", req)
	}

	// Assign
	err = svc.AssignMedicalRequest(ctx, req.ID, "medical1")
	if err != nil {
		t.Fatalf("Failed to assign medical request: %v", err)
	}

	list, _ := svc.ListMedicalRequests(ctx)
	var found bool
	for _, r := range list {
		if r.ID == req.ID {
			found = true
			if r.Status != "assigned" || r.AssignedTo != "medical1" {
				t.Errorf("Unexpected status/assignee: status=%s, assignee=%s", r.Status, r.AssignedTo)
			}
		}
	}
	if !found {
		t.Error("Medical request not found in list")
	}

	// Resolve
	err = svc.ResolveMedicalRequest(ctx, req.ID)
	if err != nil {
		t.Fatalf("Failed to resolve medical request: %v", err)
	}

	list, _ = svc.ListMedicalRequests(ctx)
	for _, r := range list {
		if r.ID == req.ID {
			if r.Status != "resolved" {
				t.Errorf("Expected status resolved, got %s", r.Status)
			}
		}
	}
}

func TestOpsService_AuditLogs(t *testing.T) {
	ctx := context.Background()
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")
	svc := NewOperationsService(dbRepo, pubSubRepo)

	logEntry := &models.AuditLog{
		ID:        "log_123",
		UserID:    "admin",
		Username:  "admin",
		Action:    "CREATE_USER",
		Resource:  "users",
		Details:   "Created volunteer2",
		Timestamp: time.Now(),
	}

	err := svc.CreateAuditLog(ctx, logEntry)
	if err != nil {
		t.Fatalf("Failed to create audit log: %v", err)
	}

	logs, err := svc.ListAuditLogs(ctx)
	if err != nil {
		t.Fatalf("Failed to list audit logs: %v", err)
	}

	if len(logs) == 0 {
		t.Fatal("Expected at least 1 audit log")
	}

	if logs[0].ID != "log_123" {
		t.Errorf("Expected ID 'log_123', got %s", logs[0].ID)
	}
}
