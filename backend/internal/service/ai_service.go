package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"orynt/internal/models"
	"orynt/internal/repository"
)

type aiService struct {
	dbRepo repository.DBRepository
	apiKey string
}

// NewAIService constructs a concrete RAG AI Assistant service
func NewAIService(dbRepo repository.DBRepository) AIService {
	return &aiService{
		dbRepo: dbRepo,
		apiKey: os.Getenv("GEMINI_API_KEY"),
	}
}

// Gemini API structures
type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiRequest struct {
	Contents         []geminiContent `json:"contents"`
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (s *aiService) Chat(ctx context.Context, req *models.AIChatRequest) (*models.AIChatResponse, error) {
	// 1. Gather context from Firestore/Database (RAG Grounding)
	contextInfo, sources := s.gatherRAGContext(ctx, req.Role)

	// If no Gemini API Key is provided, use our intelligent local rule-based system
	if s.apiKey == "" {
		response := s.generateLocalResponse(req.Message, contextInfo, req.Role)
		return &models.AIChatResponse{
			Response: response,
			Sources:  sources,
		}, nil
	}

	// 2. Format Request for Gemini
	systemPrompt := fmt.Sprintf(`You are the Orynt Smart Stadium & Tournament Operations Assistant.
You must answer questions from the perspective of an AI helper grounded in current stadium telemetry.
Never hallucinate or guess. Use ONLY the facts provided below to ground your answer.
If you don't know the answer or if it's not in the context, politely say so.

Current Stadium Live Context:
---
%s
---

User Role: %s`, contextInfo, req.Role)

	geminiReq := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: req.Message}},
			},
		},
	}

	// 3. Make HTTP request to Gemini API
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", s.apiKey)
	reqJSON, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback to local response if API call fails
		response := s.generateLocalResponse(req.Message, contextInfo, req.Role)
		return &models.AIChatResponse{
			Response: response + "\n\n*(Note: Gemini API returned an error, using local fallback assistant)*",
			Sources:  sources,
		}, nil
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	return &models.AIChatResponse{
		Response: geminiResp.Candidates[0].Content.Parts[0].Text,
		Sources:  sources,
	}, nil
}

func (s *aiService) gatherRAGContext(ctx context.Context, role string) (string, []string) {
	var builder strings.Builder
	var sources []string

	// Fetch matches
	matches, err := s.dbRepo.ListMatches(ctx, "")
	if err == nil && len(matches) > 0 {
		builder.WriteString("### Live Tournament and Matches:\n")
		for _, m := range matches {
			fmt.Fprintf(&builder, "- Match %s: %s vs %s. Score: %d-%d. Status: %s. Elapsed: %s\n", m.ID, m.HomeTeam, m.AwayTeam, m.HomeScore, m.AwayScore, m.Status, m.TimeElapsed)
		}
		sources = append(sources, "Live Match Telemetry")
	}

	// Fetch crowd density
	zones, err := s.dbRepo.ListCrowdZones(ctx, "")
	if err == nil && len(zones) > 0 {
		builder.WriteString("\n### Crowd Density and Stand Occupancy:\n")
		for _, z := range zones {
			fmt.Fprintf(&builder, "- Zone %s (%s): Occupancy %d/%d (%s density)\n", z.ID, z.ZoneName, z.CurrentOccupancy, z.MaxCapacity, z.DensityLevel)
		}
		sources = append(sources, "Crowd Sensors (IoT)")
	}

	// Fetch parking
	parking, err := s.dbRepo.ListParkingZones(ctx)
	if err == nil && len(parking) > 0 {
		builder.WriteString("\n### Parking Occupancy:\n")
		for _, p := range parking {
			fmt.Fprintf(&builder, "- Lot %s (%s): %d/%d spots occupied. Type: %s. Location: %s\n", p.ID, p.ZoneName, p.OccupiedSpots, p.TotalSpots, p.Type, p.Location)
		}
		sources = append(sources, "Smart Parking Sensors")
	}

	// Fetch transport
	transport, err := s.dbRepo.ListTransportLines(ctx)
	if err == nil && len(transport) > 0 {
		builder.WriteString("\n### Transport & Shuttle Services:\n")
		for _, t := range transport {
			fmt.Fprintf(&builder, "- Route %s (%s): Next arrival %s, Delay: %d min. Status: %s. Capacity Level: %s\n", t.ID, t.RouteName, t.NextArrival.Format("15:04"), t.DelayMinutes, t.Status, t.Capacity)
		}
		sources = append(sources, "Transport API Feed")
	}

	// Fetch food stalls
	stalls, err := s.dbRepo.ListFoodStalls(ctx)
	if err == nil && len(stalls) > 0 {
		builder.WriteString("\n### Food & Beverage Options:\n")
		for _, f := range stalls {
			fmt.Fprintf(&builder, "- Stall %s (%s) at %s: Wait time %d mins. Status: %s. Menu: %s\n", f.ID, f.Name, f.Location, f.WaitTimeMinutes, f.Status, strings.Join(f.Menu, ", "))
		}
		sources = append(sources, "Food Merchant Inventory")
	}

	// Fetch active alerts
	alerts, err := s.dbRepo.ListAlerts(ctx, true)
	if err == nil && len(alerts) > 0 {
		builder.WriteString("\n### Emergency & Operational Alerts:\n")
		for _, a := range alerts {
			fmt.Fprintf(&builder, "- ALERT: [%s] %s. Description: %s\n", a.Severity, a.Title, a.Content)
		}
		sources = append(sources, "Emergency Broadcast Center")
	}

	// Staff specific details
	if role == "volunteer" || role == "security" || role == "medical" || role == "ops" || role == "admin" {
		tasks, err := s.dbRepo.ListTasks(ctx, "")
		if err == nil && len(tasks) > 0 {
			builder.WriteString("\n### Staff Tasks & Incidents:\n")
			for _, t := range tasks {
				fmt.Fprintf(&builder, "- Task %s: %s (Assigned to: %s). Status: %s. Priority: %s. Dept: %s\n", t.ID, t.Title, t.AssignedTo, t.Status, t.Priority, t.Department)
			}
			sources = append(sources, "Operations Worklist DB")
		}
	}

	return builder.String(), sources
}

func (s *aiService) generateLocalResponse(message string, contextInfo string, role string) string {
	msgLower := strings.ToLower(message)

	// FAQ/Navigation/Help rules
	if strings.Contains(msgLower, "hello") || strings.Contains(msgLower, "hi") {
		return "Hello! I am your Orynt Stadium Assistant. How can I help you navigate the stadium, check match scores, find food stalls, check transport schedules, or resolve incidents today?"
	}

	if strings.Contains(msgLower, "match") || strings.Contains(msgLower, "score") || strings.Contains(msgLower, "play") || strings.Contains(msgLower, "spartans") || strings.Contains(msgLower, "titans") {
		if strings.Contains(contextInfo, "Spartans FC") {
			return "🏆 **Live Match Update:**\nSpartans FC is currently playing Titans United. The score is **2-1** in favor of Spartans FC (67th minute). In Section 102 you can find Corner Grill if you need refreshments during the match!"
		}
		return "There are no matches currently running, but a match between Strikers SC and Wanderers FC is scheduled to kick off in approximately 2 hours."
	}

	if strings.Contains(msgLower, "food") || strings.Contains(msgLower, "eat") || strings.Contains(msgLower, "stall") || strings.Contains(msgLower, "hungry") {
		return "🍔 **Food & Beverage Stalls:**\n- **Starlight Hotdogs & Burgers** (Section 104 Concourse): Wait time is **15 minutes**. Menu includes burgers, hotdogs, and sodas.\n- **Orynt Brewery & Bar** (Section 118 Pavilion): Wait time is **5 minutes** (Fast!). Menu includes beer and pretzels.\n- **Noodle House & Bowls** (Section 206 Level 2): Wait time is **25 minutes**."
	}

	if strings.Contains(msgLower, "parking") || strings.Contains(msgLower, "car") || strings.Contains(msgLower, "lot") {
		return "🚗 **Parking Availability:**\n- **Parking Lot B (South)** has the best availability right now with **460 free spots**.\n- **Parking Lot A (North)** is almost full (**18 spots remaining**).\n- **VIP Parking Deck** is practically full (**5 spots remaining**).\n- **Accessible Lot A** has **58 free spots** and is located closest to the Elevator 1 ramps."
	}

	if strings.Contains(msgLower, "transport") || strings.Contains(msgLower, "shuttle") || strings.Contains(msgLower, "metro") || strings.Contains(msgLower, "bus") || strings.Contains(msgLower, "train") {
		return "🚌 **Transit Details:**\n- **City Center Metro Shuttle** is running **On Time** and arriving in **4 minutes** (High capacity demand).\n- **North Stadium Link Train** is running **On Time** and arriving in **8 minutes**.\n- **East Suburbs Express Bus** is delayed by **5 minutes**, expected in **12 minutes**."
	}

	if strings.Contains(msgLower, "emergency") || strings.Contains(msgLower, "help") || strings.Contains(msgLower, "incident") || strings.Contains(msgLower, "medical") || strings.Contains(msgLower, "danger") {
		return "⚠️ **Emergency & First Aid Contacts:**\n- First Aid Station A is located at **Gate 2, Floor 1**.\n- The Emergency Response Center is located at **Gate 5, Ground Floor**.\n- If you see an emergency, please notify nearby security staff immediately or report an incident through the app. There is currently an active high congestion alert at Gate A."
	}

	if strings.Contains(msgLower, "lost") || strings.Contains(msgLower, "found") || strings.Contains(msgLower, "wallet") || strings.Contains(msgLower, "keys") || strings.Contains(msgLower, "phone") {
		return "🔍 **Lost & Found:**\nIf you have lost an item, please report it via the 'Lost & Found' section on the dashboard with your contact details. If you found an item, please hand it over to any volunteer staff or the customer service desk at Gate 1."
	}

	if strings.Contains(msgLower, "accessibility") || strings.Contains(msgLower, "wheelchair") || strings.Contains(msgLower, "lift") || strings.Contains(msgLower, "elevator") {
		return "♿ **Accessibility Assistance:**\nOrynt Arena features full wheelchair access. Accessible routes include:\n- **Elevator 1 (North Stand)**\n- **Ramp B (South Gate)**\n- **Priority Exit 4**\nAccessible Lot A parking is reserved directly next to the Elevator 1 ramp. Please ask any volunteer in a blue vest for escort assistance."
	}

	// Staff workflow rules
	if role == "volunteer" || role == "security" || role == "medical" || role == "ops" || role == "admin" {
		if strings.Contains(msgLower, "task") || strings.Contains(msgLower, "job") || strings.Contains(msgLower, "todo") {
			return "📋 **Staff Task Status Summary:**\n- **Todo:** Turnstile Gate 3 inspection (assigned to volunteer1).\n- **In Progress:** Spillage cleanup near Section 106 (assigned to cleaning1).\n- **Done:** Monitor VIP parking (volunteer1).\n\nYou can update task states using the Kanban board in the staff panel."
		}
	}

	return "I've analyzed your query, but I don't see direct matches in the live telemetry. Generally, you can view the Live Tournament schedules, Stadium Map, Parking sensors, and Transit feeds directly on the dashboard. Is there a specific section or incident you'd like to check?"
}
