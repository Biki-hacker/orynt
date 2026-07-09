package api

import (
	"orynt/internal/repository"

	"github.com/gin-gonic/gin"
)

// SetupRouter registers middleware and routes
func SetupRouter(handler *APIHandler, wsHub *WSHub, pubSub repository.PubSubRepository) *gin.Engine {
	r := gin.New()

	// Global Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())
	r.Use(RateLimitMiddleware(pubSub))

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP", "time": "2026-07-07T21:40:00Z"})
	})

	// WebSockets Upgrader
	r.GET("/api/ws", WebSocketHandler(wsHub))

	api := r.Group("/api")
	{
		// ----------------- PUBLIC ENDPOINTS -----------------
		api.POST("/auth/login", handler.Login)

		// Public Telemetry & Info
		api.GET("/stadium", handler.GetStadium)
		api.GET("/stadium/zones", handler.GetCrowdZones)
		api.GET("/stadium/exit-routing", handler.GetExitRoutes)
		api.GET("/parking", handler.GetParking)
		api.GET("/transport", handler.GetTransport)
		api.GET("/food-stalls", handler.GetFoodStalls)
		api.GET("/alerts", handler.GetAlerts)
		api.GET("/announcements", handler.GetAnnouncements)
		api.GET("/tournaments", handler.ListTournaments)
		api.GET("/matches", handler.ListMatches)

		// Public Lost & Found Submission
		api.POST("/lost-found", handler.ReportLostItem)
		api.GET("/lost-found", handler.ListLostItems)

		// Public & Anonymous Medical help
		api.POST("/medical", handler.CreateMedicalRequest)

		// RAG AI Chat (grounded queries)
		api.POST("/ai/chat", handler.AIChat)

		// Sustainability Dashboard Metrics & Recommendations
		api.GET("/sustainability", handler.GetSustainabilityMetrics)

		// ----------------- SECURE STAFF ENDPOINTS -----------------
		staff := api.Group("")
		staff.Use(AuthMiddleware(handler.authService))
		staff.Use(RBACMiddleware("volunteer", "security", "medical", "cleaning", "ops"))
		{
			// Task boards (Kanban)
			staff.GET("/tasks", handler.GetTasks)
			staff.PUT("/tasks/:id", handler.UpdateTask)
			staff.POST("/tasks", handler.CreateTask) // Staff or Admin can create

			// Stalls
			staff.PUT("/food-stalls/:id", handler.UpdateFoodWaitTime)

			// Alerts incident broadcast
			staff.POST("/alerts", handler.BroadcastAlert)

			// Medical assignments
			staff.GET("/medical", handler.GetMedicalRequests)
			staff.PUT("/medical/:id/assign", handler.AssignMedicalRequest)
			staff.PUT("/medical/:id/resolve", handler.ResolveMedicalRequest)

			// Lost claim
			staff.PUT("/lost-found/:id/claim", handler.ClaimLostItem)

			// Staff announcements queue
			staff.POST("/announcements", handler.CreateAnnouncement)
		}

		// ----------------- SECURE ADMIN ENDPOINTS -----------------
		admin := api.Group("")
		admin.Use(AuthMiddleware(handler.authService))
		admin.Use(RBACMiddleware("admin"))
		{
			// Admin user creation & listing
			admin.POST("/users", handler.CreateUser)
			admin.GET("/users", handler.GetUsers)

			// Tournament & Match operations
			admin.POST("/tournaments", handler.CreateTournament)
			admin.POST("/matches", handler.CreateMatch)
			admin.PUT("/matches/:id", handler.UpdateMatchScore)
			admin.POST("/matches/:id/events", handler.AddMatchEvent)

			// Telemetry overriding for mock simulator
			admin.PUT("/stadium/zones/:id", handler.UpdateCrowdZone)
			admin.PUT("/parking/:id", handler.UpdateParking)
			admin.PUT("/transport/:id", handler.UpdateTransport)
			admin.PUT("/alerts/:id/resolve", handler.ResolveAlert)

			// Announcement moderation
			admin.PUT("/announcements/:id/approve", handler.ApproveAnnouncement)

			// Audit Logs
			admin.GET("/audit-logs", handler.GetAuditLogs)
			admin.GET("/metrics", handler.GetMetrics)
		}
	}

	return r
}
