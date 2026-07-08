package service

import (
	"context"
	"fmt"
	"time"
	"orynt/internal/models"
	"orynt/internal/repository"
)

type opsService struct {
	dbRepo     repository.DBRepository
	pubSubRepo repository.PubSubRepository
}

// NewOperationsService creates concrete operations service
func NewOperationsService(dbRepo repository.DBRepository, pubSubRepo repository.PubSubRepository) OperationsService {
	return &opsService{
		dbRepo:     dbRepo,
		pubSubRepo: pubSubRepo,
	}
}

// Task management
func (s *opsService) ListTasks(ctx context.Context, department string) ([]models.Task, error) {
	return s.dbRepo.ListTasks(ctx, department)
}

func (s *opsService) CreateTask(ctx context.Context, title, description, assignedTo, priority, department string) (*models.Task, error) {
	task := &models.Task{
		ID:          fmt.Sprintf("task_%d", time.Now().UnixNano()),
		Title:       title,
		Description: description,
		AssignedTo:  assignedTo,
		Status:      "todo",
		Priority:    priority,
		Department:  department,
		CreatedAt:   time.Now(),
	}

	err := s.dbRepo.CreateTask(ctx, task)
	if err != nil {
		return nil, err
	}

	// Publish to tasks channel
	msg := models.WSMessage{
		Type:    "task_update",
		Payload: task,
	}
	_ = s.pubSubRepo.Publish(ctx, "tasks", msg)

	return task, nil
}

func (s *opsService) UpdateTaskStatus(ctx context.Context, taskID, status string) error {
	err := s.dbRepo.UpdateTaskStatus(ctx, taskID, status)
	if err != nil {
		return err
	}

	// Fetch all tasks and find the updated one or construct payload
	tasks, err := s.dbRepo.ListTasks(ctx, "")
	if err == nil {
		var updatedTask models.Task
		for _, t := range tasks {
			if t.ID == taskID {
				updatedTask = t
				break
			}
		}
		msg := models.WSMessage{
			Type:    "task_update",
			Payload: updatedTask,
		}
		_ = s.pubSubRepo.Publish(ctx, "tasks", msg)
	}

	return nil
}

// Lost & Found
func (s *opsService) ListLostFound(ctx context.Context) ([]models.LostFoundItem, error) {
	return s.dbRepo.ListLostFound(ctx)
}

func (s *opsService) ReportLostFoundItem(ctx context.Context, name, description, category, contactName, contactPhone string) (*models.LostFoundItem, error) {
	item := &models.LostFoundItem{
		ID:           fmt.Sprintf("lf_%d", time.Now().UnixNano()),
		ItemName:     name,
		Description:  description,
		Category:     category,
		Status:       "reported_lost",
		ContactName:  contactName,
		ContactPhone: contactPhone,
		ReportedAt:   time.Now(),
	}

	err := s.dbRepo.CreateLostFound(ctx, item)
	if err != nil {
		return nil, err
	}

	msg := models.WSMessage{
		Type:    "lost_found_report",
		Payload: item,
	}
	_ = s.pubSubRepo.Publish(ctx, "lost_found", msg)

	return item, nil
}

func (s *opsService) ClaimLostFoundItem(ctx context.Context, itemID string) error {
	err := s.dbRepo.UpdateLostFoundStatus(ctx, itemID, "claimed")
	if err != nil {
		return err
	}

	msg := models.WSMessage{
		Type: "lost_found_claim",
		Payload: map[string]string{
			"id": itemID,
		},
	}
	_ = s.pubSubRepo.Publish(ctx, "lost_found", msg)

	return nil
}

// Food Stalls
func (s *opsService) ListFoodStalls(ctx context.Context) ([]models.FoodStall, error) {
	return s.dbRepo.ListFoodStalls(ctx)
}

func (s *opsService) UpdateFoodStallWaitTime(ctx context.Context, stallID string, waitTime int) error {
	err := s.dbRepo.UpdateFoodStallWaitTime(ctx, stallID, waitTime)
	if err != nil {
		return err
	}

	stalls, err := s.dbRepo.ListFoodStalls(ctx)
	if err == nil {
		var updatedStall models.FoodStall
		for _, fs := range stalls {
			if fs.ID == stallID {
				updatedStall = fs
				break
			}
		}
		msg := models.WSMessage{
			Type:    "food_stall",
			Payload: updatedStall,
		}
		_ = s.pubSubRepo.Publish(ctx, "stalls", msg)
	}

	return nil
}

// Announcements
func (s *opsService) CreateAnnouncement(ctx context.Context, title, content, targetAudience string) (*models.Announcement, error) {
	ann := &models.Announcement{
		ID:             fmt.Sprintf("ann_%d", time.Now().UnixNano()),
		Title:          title,
		Content:        content,
		TargetAudience: targetAudience,
		Approved:       false, // Admin must approve
		ApprovedBy:     "",
		CreatedAt:      time.Now(),
	}

	err := s.dbRepo.CreateAnnouncement(ctx, ann)
	if err != nil {
		return nil, err
	}

	// Notify admins for approval
	msg := models.WSMessage{
		Type:    "announcement_pending",
		Payload: ann,
	}
	_ = s.pubSubRepo.Publish(ctx, "announcements", msg)

	return ann, nil
}

func (s *opsService) ApproveAnnouncement(ctx context.Context, id string, adminUsername string) error {
	err := s.dbRepo.ApproveAnnouncement(ctx, id, adminUsername)
	if err != nil {
		return err
	}

	// Fetch all approved announcements
	anns, err := s.dbRepo.ListAnnouncements(ctx, "")
	if err == nil {
		var approvedAnn models.Announcement
		for _, a := range anns {
			if a.ID == id {
				approvedAnn = a
				break
			}
		}
		// Notify clients
		msg := models.WSMessage{
			Type:    "announcement",
			Payload: approvedAnn,
		}
		_ = s.pubSubRepo.Publish(ctx, "announcements", msg)
	}

	return nil
}

func (s *opsService) ListAnnouncements(ctx context.Context, role string) ([]models.Announcement, error) {
	return s.dbRepo.ListAnnouncements(ctx, role)
}

// Medical requests
func (s *opsService) ListMedicalRequests(ctx context.Context) ([]models.MedicalRequest, error) {
	return s.dbRepo.ListMedicalRequests(ctx)
}

func (s *opsService) CreateMedicalRequest(ctx context.Context, requester, location, description string) (*models.MedicalRequest, error) {
	req := &models.MedicalRequest{
		ID:          fmt.Sprintf("med_%d", time.Now().UnixNano()),
		Requester:   requester,
		Location:    location,
		Description: description,
		Status:      "pending",
		AssignedTo:  "",
		RequestedAt: time.Now(),
	}

	err := s.dbRepo.CreateMedicalRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Notify medical staff
	msg := models.WSMessage{
		Type:    "medical_request",
		Payload: req,
	}
	_ = s.pubSubRepo.Publish(ctx, "medical", msg)

	return req, nil
}

func (s *opsService) AssignMedicalRequest(ctx context.Context, id string, assignedTo string) error {
	err := s.dbRepo.UpdateMedicalRequestStatus(ctx, id, "assigned", assignedTo)
	if err != nil {
		return err
	}

	requests, err := s.dbRepo.ListMedicalRequests(ctx)
	if err == nil {
		var updatedReq models.MedicalRequest
		for _, r := range requests {
			if r.ID == id {
				updatedReq = r
				break
			}
		}
		msg := models.WSMessage{
			Type:    "medical_update",
			Payload: updatedReq,
		}
		_ = s.pubSubRepo.Publish(ctx, "medical", msg)
	}

	return nil
}

func (s *opsService) ResolveMedicalRequest(ctx context.Context, id string) error {
	err := s.dbRepo.UpdateMedicalRequestStatus(ctx, id, "resolved", "")
	if err != nil {
		return err
	}

	requests, err := s.dbRepo.ListMedicalRequests(ctx)
	if err == nil {
		var updatedReq models.MedicalRequest
		for _, r := range requests {
			if r.ID == id {
				updatedReq = r
				break
			}
		}
		msg := models.WSMessage{
			Type:    "medical_update",
			Payload: updatedReq,
		}
		_ = s.pubSubRepo.Publish(ctx, "medical", msg)
	}

	return nil
}

func (s *opsService) ListAuditLogs(ctx context.Context) ([]models.AuditLog, error) {
	return s.dbRepo.ListAuditLogs(ctx)
}

func (s *opsService) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	return s.dbRepo.CreateAuditLog(ctx, log)
}
