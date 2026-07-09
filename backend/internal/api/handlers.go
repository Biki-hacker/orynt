package api

import (
	"context"
	"net/http"
	"orynt/internal/models"
	"orynt/internal/service"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	authService  service.AuthService
	tournService service.TournamentService
	stadService  service.StadiumService
	opsService   service.OperationsService
	aiService    service.AIService
}

func NewAPIHandler(auth service.AuthService, tourn service.TournamentService, stad service.StadiumService, ops service.OperationsService, ai service.AIService) *APIHandler {
	return &APIHandler{
		authService:  auth,
		tournService: tourn,
		stadService:  stad,
		opsService:   ops,
		aiService:    ai,
	}
}

// ----------------- AUTH HANDLERS -----------------

func (h *APIHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fields"})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username   string `json:"username" binding:"required"`
		Password   string `json:"password" binding:"required"`
		Role       string `json:"role" binding:"required"`
		Department string `json:"department"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.CreateUser(c.Request.Context(), req.Username, req.Password, req.Role, req.Department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logAdminAction(c, "create_user", "users", user.ID)
	c.JSON(http.StatusCreated, user)
}

func (h *APIHandler) GetUsers(c *gin.Context) {
	users, err := h.authService.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// ----------------- TOURNAMENT HANDLERS -----------------

func (h *APIHandler) CreateTournament(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := h.tournService.CreateTournament(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logAdminAction(c, "create_tournament", "tournaments", t.ID)
	c.JSON(http.StatusCreated, t)
}

func (h *APIHandler) ListTournaments(c *gin.Context) {
	tourns, err := h.tournService.ListTournaments(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tourns)
}

func (h *APIHandler) CreateMatch(c *gin.Context) {
	var req struct {
		TournamentID  string `json:"tournamentId" binding:"required"`
		HomeTeam      string `json:"homeTeam" binding:"required"`
		AwayTeam      string `json:"awayTeam" binding:"required"`
		ScheduledTime string `json:"scheduledTime" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.tournService.CreateMatch(c.Request.Context(), req.TournamentID, req.HomeTeam, req.AwayTeam, req.ScheduledTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logAdminAction(c, "create_match", "matches", m.ID)
	c.JSON(http.StatusCreated, m)
}

func (h *APIHandler) ListMatches(c *gin.Context) {
	tID := c.Query("tournamentId")
	matches, err := h.tournService.ListMatches(c.Request.Context(), tID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, matches)
}

func (h *APIHandler) UpdateMatchScore(c *gin.Context) {
	mID := c.Param("id")
	var req struct {
		HomeScore   int    `json:"homeScore"`
		AwayScore   int    `json:"awayScore"`
		TimeElapsed string `json:"timeElapsed" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.tournService.UpdateMatchScore(c.Request.Context(), mID, req.HomeScore, req.AwayScore, req.TimeElapsed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logAdminAction(c, "update_match_score", "matches", m.ID)
	c.JSON(http.StatusOK, m)
}

func (h *APIHandler) AddMatchEvent(c *gin.Context) {
	mID := c.Param("id")
	var req struct {
		Time      string `json:"time" binding:"required"`
		EventType string `json:"type" binding:"required"` // goal, card, substitutions
		Detail    string `json:"detail" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.tournService.AddMatchEvent(c.Request.Context(), mID, req.Time, req.EventType, req.Detail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logAdminAction(c, "add_match_event", "matches", m.ID)
	c.JSON(http.StatusOK, m)
}

// ----------------- STADIUM TELEMETRY HANDLERS -----------------

func (h *APIHandler) GetStadium(c *gin.Context) {
	stad, err := h.stadService.GetStadiumDetails(c.Request.Context(), "stadium_01")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stad)
}

func (h *APIHandler) GetCrowdZones(c *gin.Context) {
	zones, err := h.stadService.GetCrowdZones(c.Request.Context(), "stadium_01")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, zones)
}

func (h *APIHandler) GetExitRoutes(c *gin.Context) {
	routes, err := h.stadService.GetExitRoutes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, routes)
}

func (h *APIHandler) UpdateCrowdZone(c *gin.Context) {
	zID := c.Param("id")
	var req struct {
		Occupancy int `json:"occupancy" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.stadService.UpdateZoneOccupancy(c.Request.Context(), zID, req.Occupancy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Zone updated"})
}

func (h *APIHandler) GetParking(c *gin.Context) {
	lots, err := h.stadService.GetParkingZones(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lots)
}

func (h *APIHandler) UpdateParking(c *gin.Context) {
	lotID := c.Param("id")
	var req struct {
		Occupied int `json:"occupied" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.stadService.UpdateParking(c.Request.Context(), lotID, req.Occupied)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Parking updated"})
}

func (h *APIHandler) GetTransport(c *gin.Context) {
	lines, err := h.stadService.GetTransportLines(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lines)
}

func (h *APIHandler) UpdateTransport(c *gin.Context) {
	routeID := c.Param("id")
	var req struct {
		Delay  int    `json:"delay"`
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.stadService.UpdateTransport(c.Request.Context(), routeID, req.Delay, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transport updated"})
}

func (h *APIHandler) BroadcastAlert(c *gin.Context) {
	var req struct {
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Type     string `json:"type" binding:"required"` // emergency, weather, crowd, transport
		Severity string `json:"severity" binding:"required"` // critical, high, info
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alert, err := h.stadService.BroadcastAlert(c.Request.Context(), req.Title, req.Content, req.Type, req.Severity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, alert)
}

func (h *APIHandler) GetAlerts(c *gin.Context) {
	activeStr := c.Query("active")
	activeOnly := activeStr == "true" || activeStr == "" // default to true if empty

	alerts, err := h.stadService.GetAlerts(c.Request.Context(), activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

func (h *APIHandler) ResolveAlert(c *gin.Context) {
	alertID := c.Param("id")
	err := h.stadService.ResolveAlert(c.Request.Context(), alertID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logAdminAction(c, "resolve_alert", "alerts", alertID)
	c.JSON(http.StatusOK, gin.H{"message": "Alert resolved"})
}

// ----------------- OPERATIONS & TASKS HANDLERS -----------------

func (h *APIHandler) GetTasks(c *gin.Context) {
	dept := c.Query("department")
	tasks, err := h.opsService.ListTasks(c.Request.Context(), dept)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *APIHandler) CreateTask(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		AssignedTo  string `json:"assignedTo" binding:"required"`
		Priority    string `json:"priority" binding:"required"`
		Department  string `json:"department" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.opsService.CreateTask(c.Request.Context(), req.Title, req.Description, req.AssignedTo, req.Priority, req.Department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *APIHandler) UpdateTask(c *gin.Context) {
	tID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.opsService.UpdateTaskStatus(c.Request.Context(), tID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated"})
}

func (h *APIHandler) ReportLostItem(c *gin.Context) {
	var req struct {
		ItemName     string `json:"itemName" binding:"required"`
		Description  string `json:"description" binding:"required"`
		Category     string `json:"category" binding:"required"`
		ContactName  string `json:"contactName" binding:"required"`
		ContactPhone string `json:"contactPhone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.opsService.ReportLostFoundItem(c.Request.Context(), req.ItemName, req.Description, req.Category, req.ContactName, req.ContactPhone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *APIHandler) ListLostItems(c *gin.Context) {
	items, err := h.opsService.ListLostFound(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *APIHandler) ClaimLostItem(c *gin.Context) {
	itemID := c.Param("id")
	err := h.opsService.ClaimLostFoundItem(c.Request.Context(), itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Item marked as claimed"})
}

func (h *APIHandler) GetFoodStalls(c *gin.Context) {
	stalls, err := h.opsService.ListFoodStalls(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stalls)
}

func (h *APIHandler) UpdateFoodWaitTime(c *gin.Context) {
	fID := c.Param("id")
	var req struct {
		WaitTime int `json:"waitTime" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.opsService.UpdateFoodStallWaitTime(c.Request.Context(), fID, req.WaitTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Wait time updated"})
}

func (h *APIHandler) GetMedicalRequests(c *gin.Context) {
	reqs, err := h.opsService.ListMedicalRequests(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reqs)
}

func (h *APIHandler) CreateMedicalRequest(c *gin.Context) {
	var req struct {
		Requester   string `json:"requester" binding:"required"`
		Location    string `json:"location" binding:"required"`
		Description string `json:"description" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mReq, err := h.opsService.CreateMedicalRequest(c.Request.Context(), req.Requester, req.Location, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, mReq)
}

func (h *APIHandler) AssignMedicalRequest(c *gin.Context) {
	mID := c.Param("id")
	var req struct {
		AssignedTo string `json:"assignedTo" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.opsService.AssignMedicalRequest(c.Request.Context(), mID, req.AssignedTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Medical request assigned"})
}

func (h *APIHandler) ResolveMedicalRequest(c *gin.Context) {
	mID := c.Param("id")
	err := h.opsService.ResolveMedicalRequest(c.Request.Context(), mID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Medical request resolved"})
}

func (h *APIHandler) CreateAnnouncement(c *gin.Context) {
	var req struct {
		Title          string `json:"title" binding:"required"`
		Content        string `json:"content" binding:"required"`
		TargetAudience string `json:"targetAudience" binding:"required"` // public, staff, all
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ann, err := h.opsService.CreateAnnouncement(c.Request.Context(), req.Title, req.Content, req.TargetAudience)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ann)
}

func (h *APIHandler) ApproveAnnouncement(c *gin.Context) {
	annID := c.Param("id")
	userObj, exists := c.Get("currentUser")
	adminUsername := "admin"
	if exists {
		if user, ok := userObj.(*models.User); ok {
			adminUsername = user.Username
		}
	}

	err := h.opsService.ApproveAnnouncement(c.Request.Context(), annID, adminUsername)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logAdminAction(c, "approve_announcement", "announcements", annID)
	c.JSON(http.StatusOK, gin.H{"message": "Announcement approved"})
}

func (h *APIHandler) GetAnnouncements(c *gin.Context) {
	role := c.Query("role")
	if role == "" {
		role = "public"
	}

	anns, err := h.opsService.ListAnnouncements(c.Request.Context(), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, anns)
}

// ----------------- AI HANDLERS -----------------

func (h *APIHandler) AIChat(c *gin.Context) {
	var req models.AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}

	// Double check user role context, parsing token manually if auth middleware was skipped (public route)
	userObj, exists := c.Get("currentUser")
	if !exists {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr := parts[1]
				user, err := h.authService.ValidateToken(tokenStr)
				if err == nil {
					userObj = user
					exists = true
					c.Set("currentUser", user)
				}
			}
		}
	}

	if exists {
		if user, ok := userObj.(*models.User); ok {
			req.Role = user.Role
		}
	} else if req.Role == "" {
		req.Role = "public"
	}

	resp, err := h.aiService.Chat(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ----------------- AUDIT LOGS -----------------

func (h *APIHandler) GetAuditLogs(c *gin.Context) {
	logs, err := h.opsService.ListAuditLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (h *APIHandler) GetMetrics(c *gin.Context) {
	// Analytics telemetry endpoint (aggregating in-memory or database telemetry)
	zones, _ := h.stadService.GetCrowdZones(c.Request.Context(), "stadium_01")
	parking, _ := h.stadService.GetParkingZones(c.Request.Context())
	tasks, _ := h.opsService.ListTasks(c.Request.Context(), "")
	lostItems, _ := h.opsService.ListLostFound(c.Request.Context())
	medical, _ := h.opsService.ListMedicalRequests(c.Request.Context())

	// Calculate statistics
	peakCrowd := 0
	currentTotalCrowd := 0
	for _, z := range zones {
		currentTotalCrowd += z.CurrentOccupancy
		if z.CurrentOccupancy > peakCrowd {
			peakCrowd = z.CurrentOccupancy
		}
	}

	occupiedParking := 0
	totalParking := 0
	for _, p := range parking {
		occupiedParking += p.OccupiedSpots
		totalParking += p.TotalSpots
	}

	completedTasks := 0
	for _, t := range tasks {
		if t.Status == "done" {
			completedTasks++
		}
	}

	medicalPending := 0
	for _, m := range medical {
		if m.Status != "resolved" {
			medicalPending++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"attendance":           currentTotalCrowd,
		"peakCrowd":            peakCrowd,
		"parkingOccupied":      occupiedParking,
		"parkingTotal":         totalParking,
		"queueLengthAverage":   12, // mock metric
		"shuttleDelayAverage":  3,  // mock metric
		"wasteGeneratedKg":     420 + currentTotalCrowd/120,
		"energyUsageKw":        1800 + currentTotalCrowd/50,
		"waterConsumptionL":    12000 + currentTotalCrowd/4,
		"completedTasks":       completedTasks,
		"totalTasks":           len(tasks),
		"unclaimedLostItems":   len(lostItems),
		"medicalPendingAlerts": medicalPending,
		"timestamp":            time.Now(),
	})
}

// ----------------- HELPERS -----------------

func (h *APIHandler) logAdminAction(c *gin.Context, action string, resource string, details string) {
	userObj, exists := c.Get("currentUser")
	var uID, uName string
	if exists {
		if u, ok := userObj.(*models.User); ok {
			uID = u.ID
			uName = u.Username
		}
	} else {
		uID = "system"
		uName = "System Manager"
	}

	log := &models.AuditLog{
		ID:        "log_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		UserID:    uID,
		Username:  uName,
		Action:    action,
		Resource:  resource,
		Details:   details,
		Timestamp: time.Now(),
	}

	_ = h.opsService.CreateAuditLog(context.Background(), log)
}

func (h *APIHandler) GetSustainabilityMetrics(c *gin.Context) {
	zones, _ := h.stadService.GetCrowdZones(c.Request.Context(), "stadium_01")

	currentTotalCrowd := 0
	for _, z := range zones {
		currentTotalCrowd += z.CurrentOccupancy
	}

	// Compute current metrics
	energy := float64(1800 + currentTotalCrowd/50)
	water := float64(12000 + currentTotalCrowd/4)
	waste := float64(420 + currentTotalCrowd/120)

	// Generate 6 hours of historical data points representing match progress
	historicalPoints := []gin.H{}
	now := time.Now()
	attendanceTrend := []int{
		int(float64(currentTotalCrowd) * 0.15), // 5 hrs ago (gates open)
		int(float64(currentTotalCrowd) * 0.45), // 4 hrs ago
		int(float64(currentTotalCrowd) * 0.85), // 3 hrs ago (warmups)
		int(float64(currentTotalCrowd) * 0.98), // 2 hrs ago (kickoff)
		int(float64(currentTotalCrowd) * 1.00), // 1 hr ago (halftime)
		currentTotalCrowd,                      // Current
	}

	for i, att := range attendanceTrend {
		histTime := now.Add(time.Duration(-5+i) * time.Hour).Format("15:04")
		histEnergy := 1000.0 + float64(att)/45.0 + float64(i*80)
		histWater := 5000.0 + float64(att)/3.5 + float64(i*200)
		histWaste := 100.0 + float64(att)/100.0 + float64(i*15)

		historicalPoints = append(historicalPoints, gin.H{
			"time":              histTime,
			"attendance":        att,
			"energyUsageKw":     int(histEnergy),
			"waterConsumptionL": int(histWater),
			"wasteGeneratedKg":  int(histWaste),
		})
	}

	// Call Gemini Eco-Advisor for recommendations
	recs, err := h.aiService.GetSustainabilityRecommendations(c.Request.Context(), energy, water, waste, currentTotalCrowd)
	if err != nil {
		recs = "No sustainability recommendations available at this time."
	}

	c.JSON(http.StatusOK, gin.H{
		"current": gin.H{
			"attendance":        currentTotalCrowd,
			"energyUsageKw":     energy,
			"waterConsumptionL": water,
			"wasteGeneratedKg":  waste,
		},
		"historical":      historicalPoints,
		"recommendations": recs,
	})
}
