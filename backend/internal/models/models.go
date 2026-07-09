package models

import "time"

// User represents system administrators and management staff
type User struct {
	ID         string    `json:"id" firestore:"id"`
	Username   string    `json:"username" firestore:"username"`
	Password   string    `json:"-" firestore:"password"` // bcrypt hashed, omitted from JSON
	Role       string    `json:"role" firestore:"role"` // admin, volunteer, security, medical, cleaning, parking, transport, ops
	Department string    `json:"department" firestore:"department"`
	CreatedAt  time.Time `json:"createdAt" firestore:"createdAt"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse represents authentication tokens returned to client
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         User   `json:"user"`
}

// Stadium represents a venue details
type Stadium struct {
	ID               string                 `json:"id" firestore:"id"`
	Name             string                 `json:"name" firestore:"name"`
	Capacity         int                    `json:"capacity" firestore:"capacity"`
	AccessibleRoutes []string               `json:"accessibleRoutes" firestore:"accessibleRoutes"`
	Facilities       map[string][]Facility  `json:"facilities" firestore:"facilities"` // "food", "restrooms", "medical", "exits"
}

// Facility represents icons/locations on map
type Facility struct {
	ID       string `json:"id" firestore:"id"`
	Name     string `json:"name" firestore:"name"`
	Location string `json:"location" firestore:"location"` // e.g. "Gate A", "Section 104"
	Type     string `json:"type" firestore:"type"`         // food, restroom, medical, exit
}

// Tournament details
type Tournament struct {
	ID        string    `json:"id" firestore:"id"`
	Name      string    `json:"name" firestore:"name"`
	Status    string    `json:"status" firestore:"status"` // planned, active, completed
	StartDate time.Time `json:"startDate" firestore:"startDate"`
	EndDate   time.Time `json:"endDate" firestore:"endDate"`
}

// Match details
type Match struct {
	ID           string         `json:"id" firestore:"id"`
	TournamentID string         `json:"tournamentId" firestore:"tournamentId"`
	HomeTeam     string         `json:"homeTeam" firestore:"homeTeam"`
	AwayTeam     string         `json:"awayTeam" firestore:"awayTeam"`
	HomeScore    int            `json:"homeScore" firestore:"homeScore"`
	AwayScore    int            `json:"awayScore" firestore:"awayScore"`
	Status       string         `json:"status" firestore:"status"` // scheduled, live, completed
	ScheduledAt  time.Time      `json:"scheduledAt" firestore:"scheduledAt"`
	TimeElapsed  string         `json:"timeElapsed" firestore:"timeElapsed"` // e.g., "45:00", "FT"
	Events       []MatchComment `json:"events" firestore:"events"`
}

// MatchComment represents live text commentary event
type MatchComment struct {
	Time    string `json:"time" firestore:"time"`
	Type    string `json:"type" firestore:"type"` // goal, card, substitution, info
	Detail  string `json:"detail" firestore:"detail"`
}

// CrowdZone represents real-time density in sections
type CrowdZone struct {
	ID               string    `json:"id" firestore:"id"`
	StadiumID        string    `json:"stadiumId" firestore:"stadiumId"`
	ZoneName         string    `json:"zoneName" firestore:"zoneName"` // e.g. "North Stand", "West Gate"
	CurrentOccupancy int       `json:"currentOccupancy" firestore:"currentOccupancy"`
	MaxCapacity      int       `json:"maxCapacity" firestore:"maxCapacity"`
	DensityLevel     string    `json:"densityLevel" firestore:"densityLevel"` // low, medium, high
	LastUpdated      time.Time `json:"lastUpdated" firestore:"lastUpdated"`
}

// Parking zone occupancy
type Parking struct {
	ID            string `json:"id" firestore:"id"`
	ZoneName      string `json:"zoneName" firestore:"zoneName"` // e.g. "Lot A", "VIP Deck"
	TotalSpots    int    `json:"totalSpots" firestore:"totalSpots"`
	OccupiedSpots int    `json:"occupiedSpots" firestore:"occupiedSpots"`
	Type          string `json:"type" firestore:"type"`         // public, vip, staff, accessible
	Location      string `json:"location" firestore:"location"`
}

// Transport line schedule
type Transport struct {
	ID           string    `json:"id" firestore:"id"`
	RouteName    string    `json:"routeName" firestore:"routeName"` // e.g. "Express Shuttle 1"
	Mode         string    `json:"mode" firestore:"mode"`           // bus, train, shuttle
	NextArrival  time.Time `json:"nextArrival" firestore:"nextArrival"`
	DelayMinutes int       `json:"delayMinutes" firestore:"delayMinutes"`
	Status       string    `json:"status" firestore:"status"` // on_time, delayed, suspended
	Capacity     string    `json:"capacity" firestore:"capacity"` // normal, high, full
}

// Alert is for emergency/incident broadcasting
type Alert struct {
	ID        string    `json:"id" firestore:"id"`
	Title     string    `json:"title" firestore:"title"`
	Content   string    `json:"content" firestore:"content"`
	Type      string    `json:"type" firestore:"type"`         // emergency, weather, crowd, transport
	Severity  string    `json:"severity" firestore:"severity"` // critical, high, info
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
	Active    bool      `json:"active" firestore:"active"`
}

// Announcement for public/staff notices
type Announcement struct {
	ID             string    `json:"id" firestore:"id"`
	Title          string    `json:"title" firestore:"title"`
	Content        string    `json:"content" firestore:"content"`
	TargetAudience string    `json:"targetAudience" firestore:"targetAudience"` // public, staff, all
	Approved       bool      `json:"approved" firestore:"approved"`
	ApprovedBy     string    `json:"approvedBy" firestore:"approvedBy"`
	CreatedAt      time.Time `json:"createdAt" firestore:"createdAt"`
}

// FoodStall details
type FoodStall struct {
	ID              string   `json:"id" firestore:"id"`
	Name            string   `json:"name" firestore:"name"`
	Location        string   `json:"location" firestore:"location"`
	FoodType        string   `json:"foodType" firestore:"foodType"`
	Status          string   `json:"status" firestore:"status"` // open, closed
	WaitTimeMinutes int      `json:"waitTimeMinutes" firestore:"waitTimeMinutes"`
	Menu            []string `json:"menu" firestore:"menu"`
}

// MedicalRequest for emergency help
type MedicalRequest struct {
	ID          string    `json:"id" firestore:"id"`
	Requester   string    `json:"requester" firestore:"requester"`
	Location    string    `json:"location" firestore:"location"`
	Description string    `json:"description" firestore:"description"`
	Status      string    `json:"status" firestore:"status"` // pending, assigned, resolved
	AssignedTo  string    `json:"assignedTo" firestore:"assignedTo"` // Staff user ID or name
	RequestedAt time.Time `json:"requestedAt" firestore:"requestedAt"`
}

// Volunteer tracking status
type Volunteer struct {
	ID            string    `json:"id" firestore:"id"`
	Name          string    `json:"name" firestore:"name"`
	Status        string    `json:"status" firestore:"status"` // active, offline
	CurrentTaskID string    `json:"currentTaskId" firestore:"currentTaskId"`
	CheckInTime   time.Time `json:"checkInTime" firestore:"checkInTime"`
}

// Task assigned to staff members
type Task struct {
	ID          string    `json:"id" firestore:"id"`
	Title       string    `json:"title" firestore:"title"`
	Description string    `json:"description" firestore:"description"`
	AssignedTo  string    `json:"assignedTo" firestore:"assignedTo"` // Username or user ID
	Status      string    `json:"status" firestore:"status"` // todo, in_progress, done
	Priority    string    `json:"priority" firestore:"priority"` // low, medium, high
	Department  string    `json:"department" firestore:"department"` // volunteer, security, cleaning, etc.
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
}

// LostFoundItem report
type LostFoundItem struct {
	ID           string    `json:"id" firestore:"id"`
	ItemName     string    `json:"itemName" firestore:"itemName"`
	Description  string    `json:"description" firestore:"description"`
	Category     string    `json:"category" firestore:"category"` // electronics, clothing, keys, bags, other
	Status       string    `json:"status" firestore:"status"`     // reported_lost, found, claimed
	ContactName  string    `json:"contactName" firestore:"contactName"`
	ContactPhone string    `json:"contactPhone" firestore:"contactPhone"`
	ReportedAt   time.Time `json:"reportedAt" firestore:"reportedAt"`
}

// AuditLog captures admin actions
type AuditLog struct {
	ID        string    `json:"id" firestore:"id"`
	UserID    string    `json:"userId" firestore:"userId"`
	Username  string    `json:"username" firestore:"username"`
	Action    string    `json:"action" firestore:"action"`
	Resource  string    `json:"resource" firestore:"resource"`
	Details   string    `json:"details" firestore:"details"`
	Timestamp time.Time `json:"timestamp" firestore:"timestamp"`
}

// Notification for client alerts
type Notification struct {
	ID        string    `json:"id" firestore:"id"`
	UserID    string    `json:"userId" firestore:"userId"` // empty for broadcast
	Message   string    `json:"message" firestore:"message"`
	Read      bool      `json:"read" firestore:"read"`
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
}

// AIChatHistoryItem represents a message in conversation history
type AIChatHistoryItem struct {
	Sender string `json:"sender"` // "user" or "ai"
	Text   string `json:"text"`
}

// AIChatRequest represents prompt sent to Gemini API
type AIChatRequest struct {
	Message string              `json:"message" binding:"required"`
	Role    string              `json:"role"` // user's role (public, staff, admin) for context grounding
	History []AIChatHistoryItem `json:"history"`
}

// AIChatResponse returned by Gemini agent
type AIChatResponse struct {
	Response string   `json:"response"`
	Sources  []string `json:"sources"`
}

// WSMessage represents real-time WebSockets payloads
type WSMessage struct {
	Type    string      `json:"type"`    // e.g. "match_score", "crowd_density", "alert", "task_update", "announcement"
	Payload interface{} `json:"payload"`
}

// Point represents a coordinate in 2D space
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ExitRoute represents optimal route suggestions for crowd management
type ExitRoute struct {
	ZoneID          string  `json:"zoneId"`
	ZoneName        string  `json:"zoneName"`
	RecommendedExit string  `json:"recommendedExit"`
	ExitName        string  `json:"exitName"`
	CongestionLevel string  `json:"congestionLevel"`
	PathPoints      []Point `json:"pathPoints"`
	Reason          string  `json:"reason"`
}

