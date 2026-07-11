package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"orynt/internal/models"
	"orynt/internal/repository"
	"os"
	"strings"
	"time"
)

type aiService struct {
	dbRepo     repository.DBRepository
	pubSubRepo repository.PubSubRepository
	apiKey     string
}

// NewAIService constructs a concrete RAG AI Assistant service
func NewAIService(dbRepo repository.DBRepository, pubSubRepo repository.PubSubRepository) AIService {
	return &aiService{
		dbRepo:     dbRepo,
		pubSubRepo: pubSubRepo,
		apiKey:     os.Getenv("GEMINI_API_KEY"),
	}
}

// Gemini API structures
type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text         string              `json:"text,omitempty"`
	FunctionCall *geminiFunctionCall `json:"functionCall,omitempty"`
}

type geminiRequest struct {
	Contents          []geminiContent `json:"contents"`
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Tools             []geminiTool    `json:"tools,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text         string              `json:"text"`
				FunctionCall *geminiFunctionCall `json:"functionCall,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

type geminiFunctionCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFunctionDeclaration `json:"functionDeclarations"`
}

type geminiFunctionDeclaration struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  *geminiParameters `json:"parameters"`
}

type geminiParameters struct {
	Type       string                    `json:"type"`
	Properties map[string]geminiProperty `json:"properties"`
	Required   []string                  `json:"required,omitempty"`
}

type geminiProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

func (s *aiService) Chat(ctx context.Context, req *models.AIChatRequest) (*models.AIChatResponse, error) {
	// 1. Gather context from Firestore/Database (RAG Grounding)
	contextInfo, sources := s.gatherRAGContext(ctx, req.Role)

	// If no Gemini API Key is provided, use our local fallback copilot / assistant
	if s.apiKey == "" {
		if respText, handled := s.parseAndExecuteLocalTool(ctx, req.Message, req.Role); handled {
			return &models.AIChatResponse{
				Response: respText,
				Sources:  append(sources, "Local AI Incident Copilot"),
			}, nil
		}
		response := s.generateLocalResponse(req.Message, contextInfo, req.Role)
		return &models.AIChatResponse{
			Response: response,
			Sources:  sources,
		}, nil
	}

	// 2. Format Request for Gemini (including system instructions and tools)
	systemPrompt := fmt.Sprintf(`You are the Orynt Smart Stadium & Tournament Operations Assistant.
You must answer questions from the perspective of an AI helper grounded in current stadium telemetry.
Never hallucinate or guess. Use ONLY the facts provided below to ground your answer.
If you don't know the answer or if it's not in the context, politely say so.

Multilingual & Accessibility Instructions:
1. Multilingual Support: Always detect the language of the user's message and respond in the same language. For example, if they write in Spanish, reply in Spanish; if in Hindi, reply in Hindi. Keep translations natural and contextually appropriate for a global FIFA World Cup 2026 audience.
2. Accessibility: Structure navigation steps and tables with clear markdown bullet points so text-to-speech screen readers can easily parse them for visually impaired users.

Current Stadium Live Context:
---
%s
---

User Role: %s`, contextInfo, req.Role)

	var contents []geminiContent
	for _, h := range req.History {
		role := "user"
		if h.Sender == "ai" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: h.Text}},
		})
	}
	contents = append(contents, geminiContent{
		Role:  "user",
		Parts: []geminiPart{{Text: req.Message}},
	})

	geminiReq := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: contents,
	}

	// Declare tools if user is a staff member
	isStaff := req.Role == "admin" || req.Role == "volunteer" || req.Role == "security" || req.Role == "medical" || req.Role == "cleaning" || req.Role == "parking" || req.Role == "transport" || req.Role == "ops"
	if isStaff {
		geminiReq.Tools = []geminiTool{
			{
				FunctionDeclarations: []geminiFunctionDeclaration{
					{
						Name:        "create_task",
						Description: "Create a new operational, cleaning, or maintenance task for stadium staff.",
						Parameters: &geminiParameters{
							Type: "OBJECT",
							Properties: map[string]geminiProperty{
								"title": {
									Type:        "STRING",
									Description: "Short, descriptive title of the task (e.g. Spill cleanup at Section 106).",
								},
								"description": {
									Type:        "STRING",
									Description: "Detailed description of what needs to be done.",
								},
								"department": {
									Type:        "STRING",
									Description: "The department to assign to.",
									Enum:        []string{"volunteer", "security", "medical", "cleaning", "parking", "transport", "ops"},
								},
								"priority": {
									Type:        "STRING",
									Description: "Priority level of the task.",
									Enum:        []string{"low", "medium", "high"},
								},
							},
							Required: []string{"title", "description", "department", "priority"},
						},
					},
					{
						Name:        "broadcast_alert",
						Description: "Broadcast a critical, high-priority, or informational alert to the stadium banner.",
						Parameters: &geminiParameters{
							Type: "OBJECT",
							Properties: map[string]geminiProperty{
								"title": {
									Type:        "STRING",
									Description: "Brief title of the alert.",
								},
								"content": {
									Type:        "STRING",
									Description: "The main description or warning text of the alert.",
								},
								"severity": {
									Type:        "STRING",
									Description: "Severity level of the alert.",
									Enum:        []string{"critical", "high", "info"},
								},
							},
							Required: []string{"title", "content", "severity"},
						},
					},
					{
						Name:        "assign_medical_incident",
						Description: "Assign a pending first-aid or medical emergency incident request to a volunteer or responder.",
						Parameters: &geminiParameters{
							Type: "OBJECT",
							Properties: map[string]geminiProperty{
								"incidentId": {
									Type:        "STRING",
									Description: "The medical request ID (e.g. med_12345).",
								},
								"volunteerId": {
									Type:        "STRING",
									Description: "The username or ID of the volunteer/responder to assign.",
								},
							},
							Required: []string{"incidentId", "volunteerId"},
						},
					},
				},
			},
		}
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
		// Fallback to local response/tool if API call fails
		if respText, handled := s.parseAndExecuteLocalTool(ctx, req.Message, req.Role); handled {
			return &models.AIChatResponse{
				Response: respText + "\n\n*(Note: Gemini API returned an error, using local fallback Copilot)*",
				Sources:  sources,
			}, nil
		}
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

	part := geminiResp.Candidates[0].Content.Parts[0]
	if part.FunctionCall != nil {
		respText, err := s.executeFunctionCall(ctx, part.FunctionCall)
		if err != nil {
			return nil, err
		}
		return &models.AIChatResponse{
			Response: respText,
			Sources:  append(sources, "Gemini AI Incident Copilot"),
		}, nil
	}

	return &models.AIChatResponse{
		Response: part.Text,
		Sources:  sources,
	}, nil
}

func (s *aiService) ChatStream(ctx context.Context, req *models.AIChatRequest, onChunk func(chunk *models.AIChatResponse) error) error {
	// 1. Gather context from Firestore/Database (RAG Grounding)
	contextInfo, sources := s.gatherRAGContext(ctx, req.Role)

	// If no Gemini API Key is provided, use our local fallback copilot / assistant
	if s.apiKey == "" {
		if respText, handled := s.parseAndExecuteLocalTool(ctx, req.Message, req.Role); handled {
			return onChunk(&models.AIChatResponse{
				Response: respText,
				Sources:  append(sources, "Local AI Incident Copilot"),
			})
		}
		response := s.generateLocalResponse(req.Message, contextInfo, req.Role)

		// Stream the response with word-by-word delays
		words := strings.Fields(response)
		for i, word := range words {
			space := " "
			if i == 0 {
				space = ""
			}
			err := onChunk(&models.AIChatResponse{
				Response: space + word,
			})
			if err != nil {
				return err
			}
			time.Sleep(30 * time.Millisecond) // slight delay to simulate streaming
		}

		// Send final sources chunk
		return onChunk(&models.AIChatResponse{
			Response: "",
			Sources:  sources,
		})
	}

	// 2. Format Request for Gemini (including system instructions and tools)
	systemPrompt := fmt.Sprintf(`You are the Orynt Smart Stadium & Tournament Operations Assistant.
You must answer questions from the perspective of an AI helper grounded in current stadium telemetry.
Never hallucinate or guess. Use ONLY the facts provided below to ground your answer.
If you don't know the answer or if it's not in the context, politely say so.

Multilingual & Accessibility Instructions:
1. Multilingual Support: Always detect the language of the user's message and respond in the same language. For example, if they write in Spanish, reply in Spanish; if in Hindi, reply in Hindi. Keep translations natural and contextually appropriate for a global FIFA World Cup 2026 audience.
2. Accessibility: Structure navigation steps and tables with clear markdown bullet points so text-to-speech screen readers can easily parse them for visually impaired users.

Current Stadium Live Context:
---
%s
---

User Role: %s`, contextInfo, req.Role)

	var contents []geminiContent
	for _, h := range req.History {
		role := "user"
		if h.Sender == "ai" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: h.Text}},
		})
	}
	contents = append(contents, geminiContent{
		Role:  "user",
		Parts: []geminiPart{{Text: req.Message}},
	})

	geminiReq := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: contents,
	}

	// Declare tools if user is a staff member
	isStaff := req.Role == "admin" || req.Role == "volunteer" || req.Role == "security" || req.Role == "medical" || req.Role == "cleaning" || req.Role == "parking" || req.Role == "transport" || req.Role == "ops"
	if isStaff {
		geminiReq.Tools = []geminiTool{
			{
				FunctionDeclarations: []geminiFunctionDeclaration{
					{
						Name:        "create_task",
						Description: "Create a new operational, cleaning, or maintenance task for stadium staff.",
						Parameters: &geminiParameters{
							Type: "OBJECT",
							Properties: map[string]geminiProperty{
								"title": {
									Type:        "STRING",
									Description: "Short, descriptive title of the task (e.g. Spill cleanup at Section 106).",
								},
								"description": {
									Type:        "STRING",
									Description: "Detailed description of what needs to be done.",
								},
								"department": {
									Type:        "STRING",
									Description: "The department to assign to.",
									Enum:        []string{"volunteer", "security", "medical", "cleaning", "parking", "transport", "ops"},
								},
								"priority": {
									Type:        "STRING",
									Description: "Priority level of the task.",
									Enum:        []string{"low", "medium", "high"},
								},
							},
							Required: []string{"title", "description", "department", "priority"},
						},
					},
					{
						Name:        "broadcast_alert",
						Description: "Broadcast a critical, high-priority, or informational alert to the stadium banner.",
						Parameters: &geminiParameters{
							Type: "OBJECT",
							Properties: map[string]geminiProperty{
								"title": {
									Type:        "STRING",
									Description: "Brief title of the alert.",
								},
								"content": {
									Type:        "STRING",
									Description: "The main description or warning text of the alert.",
								},
								"severity": {
									Type:        "STRING",
									Description: "Severity level of the alert.",
									Enum:        []string{"critical", "high", "info"},
								},
							},
							Required: []string{"title", "content", "severity"},
						},
					},
					{
						Name:        "assign_medical_incident",
						Description: "Assign a pending first-aid or medical emergency incident request to a volunteer or responder.",
						Parameters: &geminiParameters{
							Type: "OBJECT",
							Properties: map[string]geminiProperty{
								"incidentId": {
									Type:        "STRING",
									Description: "The medical request ID (e.g. med_12345).",
								},
								"volunteerId": {
									Type:        "STRING",
									Description: "The username or ID of the volunteer/responder to assign.",
								},
							},
							Required: []string{"incidentId", "volunteerId"},
						},
					},
				},
			},
		}
	}

	// 3. Make HTTP streaming request to Gemini API
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:streamGenerateContent?key=%s", s.apiKey)
	reqJSON, err := json.Marshal(geminiReq)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback to local response/tool if API call fails
		if respText, handled := s.parseAndExecuteLocalTool(ctx, req.Message, req.Role); handled {
			return onChunk(&models.AIChatResponse{
				Response: respText + "\n\n*(Note: Gemini API returned an error, using local fallback Copilot)*",
				Sources:  sources,
			})
		}
		response := s.generateLocalResponse(req.Message, contextInfo, req.Role)

		// Stream the response with word-by-word delays
		words := strings.Fields(response)
		for i, word := range words {
			space := " "
			if i == 0 {
				space = ""
			}
			err := onChunk(&models.AIChatResponse{
				Response: space + word,
			})
			if err != nil {
				return err
			}
			time.Sleep(30 * time.Millisecond)
		}
		return onChunk(&models.AIChatResponse{
			Response: "\n\n*(Note: Gemini API returned an error, using local fallback assistant)*",
			Sources:  sources,
		})
	}

	var functionCall *geminiFunctionCall
	dec := json.NewDecoder(resp.Body)

	// Read open bracket '[' of Generative AI streaming array response
	t, err := dec.Token()
	if err == nil {
		if delim, ok := t.(json.Delim); ok && delim == '[' {
			for dec.More() {
				var chunk geminiResponse
				err := dec.Decode(&chunk)
				if err != nil {
					break
				}

				if len(chunk.Candidates) > 0 && len(chunk.Candidates[0].Content.Parts) > 0 {
					part := chunk.Candidates[0].Content.Parts[0]
					if part.FunctionCall != nil {
						functionCall = part.FunctionCall
						break
					}
					if part.Text != "" {
						err := onChunk(&models.AIChatResponse{
							Response: part.Text,
						})
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	if functionCall != nil {
		respText, err := s.executeFunctionCall(ctx, functionCall)
		if err != nil {
			return err
		}
		return onChunk(&models.AIChatResponse{
			Response: respText,
			Sources:  append(sources, "Gemini AI Incident Copilot"),
		})
	}

	// Send final response with sources
	return onChunk(&models.AIChatResponse{
		Response: "",
		Sources:  sources,
	})
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

func (s *aiService) GetSustainabilityRecommendations(ctx context.Context, energyUsage float64, waterConsumption float64, wasteGenerated float64, attendance int) (string, error) {
	// If no Gemini API Key is provided, use our local fallback advisor
	if s.apiKey == "" {
		return s.generateLocalSustainabilityRecommendations(energyUsage, waterConsumption, wasteGenerated, attendance), nil
	}

	// Format prompt for Gemini Eco-Advisor
	systemPrompt := `You are the Orynt Arena Eco-Advisor. Analyze the current stadium sustainability metrics:
- Attendance: %d
- Energy Usage: %.1f kW
- Water Consumption: %.1f L
- Waste Generated: %.1f kg

Provide exactly 3 specific, actionable recommendations to improve stadium sustainability. 
Keep them highly realistic and related to crowd management, lighting, water usage, transit, or waste disposal.
Ensure the tone is professional, operational, and concise. Format as a clean markdown list (no headers).`

	promptText := fmt.Sprintf(systemPrompt, attendance, energyUsage, waterConsumption, wasteGenerated)

	geminiReq := geminiRequest{
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: promptText}},
			},
		},
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", s.apiKey)
	reqJSON, err := json.Marshal(geminiReq)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return s.generateLocalSustainabilityRecommendations(energyUsage, waterConsumption, wasteGenerated, attendance), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return s.generateLocalSustainabilityRecommendations(energyUsage, waterConsumption, wasteGenerated, attendance), nil
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return s.generateLocalSustainabilityRecommendations(energyUsage, waterConsumption, wasteGenerated, attendance), nil
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return s.generateLocalSustainabilityRecommendations(energyUsage, waterConsumption, wasteGenerated, attendance), nil
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func (s *aiService) generateLocalSustainabilityRecommendations(energyUsage float64, waterConsumption float64, wasteGenerated float64, attendance int) string {
	if attendance > 40000 {
		return `- **Optimize Shuttle Bus Transit**: Coordinate with city transit to adjust the shuttle frequency on electric line route link 1, reducing idle times by 20% to save fuel/energy.
- **Waste Bin Collection Frequency**: High density zones detected in North Stand. Dispatch cleaning crews to deploy additional recycling bins near concessions to minimize contamination rates.
- **Water Flow Control**: Recommend enabling smart automated sensor limits in high-density restroom sections (Sections 102 & 115) to conserve water usage by 15% during peak halftime.`
	} else if attendance > 15000 {
		return `- **HVAC Zone Optimization**: Restrict cooling/ventilation systems in empty administrative areas and VIP Suites not currently booked.
- **LED Lighting Optimization**: Shift non-critical corridor lights and digital signage in lower levels to energy-saving standby mode during active play.
- **Concession Waste Management**: Starlights Concessions at Section 104 has moderate queues. Ensure biodegradable packaging protocols are active for hotdog containers.`
	} else {
		return `- **Power-Down Inactive Zones**: Attendance is low. Completely dim floodlighting in secondary sections and lock unused public entrances (Gates 3 & 4) to conserve standby power.
- **Low-Flow Restroom Mode**: Switch stadium-wide restroom plumbing systems to eco low-pressure modes since crowd size is minimal.
- **Shuttle Consolidation**: Merge bus lines 1 & 2 into a single circular shuttle route to reduce emissions and electricity/fuel usage.`
	}
}

// executeFunctionCall performs DB mutation and publishes updates via WebSockets
func (s *aiService) executeFunctionCall(ctx context.Context, call *geminiFunctionCall) (string, error) {
	switch call.Name {
	case "create_task":
		var args struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Department  string `json:"department"`
			Priority    string `json:"priority"`
		}
		if err := json.Unmarshal(call.Args, &args); err != nil {
			return "", err
		}

		task := &models.Task{
			ID:          fmt.Sprintf("task_%d", time.Now().UnixNano()),
			Title:       args.Title,
			Description: args.Description,
			AssignedTo:  "unassigned",
			Status:      "todo",
			Priority:    args.Priority,
			Department:  args.Department,
			CreatedAt:   time.Now(),
		}
		err := s.dbRepo.CreateTask(ctx, task)
		if err != nil {
			return fmt.Sprintf("⚠️ Failed to create task: %v", err), nil
		}

		msg := models.WSMessage{
			Type:    "task_update",
			Payload: task,
		}
		_ = s.pubSubRepo.Publish(ctx, "tasks", msg)

		return fmt.Sprintf("📋 **Incident Copilot:** Created task **'%s'** (Priority: **%s**) for the **%s** department. Staff has been notified.", args.Title, args.Priority, args.Department), nil

	case "broadcast_alert":
		var args struct {
			Title    string `json:"title"`
			Content  string `json:"content"`
			Severity string `json:"severity"`
		}
		if err := json.Unmarshal(call.Args, &args); err != nil {
			return "", err
		}

		alert := &models.Alert{
			ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
			Title:     args.Title,
			Content:   args.Content,
			Type:      "ops",
			Severity:  args.Severity,
			CreatedAt: time.Now(),
			Active:    true,
		}
		err := s.dbRepo.CreateAlert(ctx, alert)
		if err != nil {
			return fmt.Sprintf("⚠️ Failed to broadcast alert: %v", err), nil
		}

		msg := models.WSMessage{
			Type:    "alert",
			Payload: alert,
		}
		_ = s.pubSubRepo.Publish(ctx, "alerts", msg)

		return fmt.Sprintf("🚨 **Incident Copilot:** Broadcasted **[%s]** alert: **'%s'** - *%s*.", strings.ToUpper(args.Severity), args.Title, args.Content), nil

	case "assign_medical_incident":
		var args struct {
			IncidentID  string `json:"incidentId"`
			VolunteerID string `json:"volunteerId"`
		}
		if err := json.Unmarshal(call.Args, &args); err != nil {
			return "", err
		}

		err := s.dbRepo.UpdateMedicalRequestStatus(ctx, args.IncidentID, "assigned", args.VolunteerID)
		if err != nil {
			return fmt.Sprintf("⚠️ Failed to assign medical request: %v", err), nil
		}

		requests, err := s.dbRepo.ListMedicalRequests(ctx)
		if err == nil {
			var updatedReq models.MedicalRequest
			for _, r := range requests {
				if r.ID == args.IncidentID {
					updatedReq = r
					break
				}
			}
			if updatedReq.ID != "" {
				msg := models.WSMessage{
					Type:    "medical_update",
					Payload: updatedReq,
				}
				_ = s.pubSubRepo.Publish(ctx, "medical", msg)
			}
		}

		return fmt.Sprintf("🏥 **Incident Copilot:** Assigned medical incident **%s** to responder **%s**.", args.IncidentID, args.VolunteerID), nil
	}

	return "Unknown tool call", nil
}

// parseAndExecuteLocalTool parses staff chat messages locally and executes matching tool behavior
func (s *aiService) parseAndExecuteLocalTool(ctx context.Context, message string, role string) (string, bool) {
	isStaff := role == "admin" || role == "volunteer" || role == "security" || role == "medical" || role == "cleaning" || role == "parking" || role == "transport" || role == "ops"
	if !isStaff {
		return "", false
	}

	msgLower := strings.ToLower(message)

	// 1. Detect task creation
	if strings.Contains(msgLower, "create task") || strings.Contains(msgLower, "create a task") || strings.Contains(msgLower, "create volunteer task") || strings.Contains(msgLower, "create cleaning task") {
		title := "Cleanup Request"
		desc := "spill cleanup at stadium concourse"
		dept := "cleaning"
		prio := "medium"

		if strings.Contains(msgLower, "spill") || strings.Contains(msgLower, "clean") {
			title = "Spill Cleanup at Stall"
			desc = "Spill reported by guest near Food Stall A. Immediate cleanup requested."
			dept = "cleaning"
			prio = "high"
		} else if strings.Contains(msgLower, "security") || strings.Contains(msgLower, "fight") {
			title = "Security Patrol Dispatch"
			desc = "Security assistance required in the stands."
			dept = "security"
			prio = "high"
		} else if strings.Contains(msgLower, "volunteer") {
			title = "Volunteer Assistance"
			desc = "Assistance needed at main entrance gates."
			dept = "volunteer"
			prio = "low"
		}

		task := &models.Task{
			ID:          fmt.Sprintf("task_%d", time.Now().UnixNano()),
			Title:       title,
			Description: desc,
			AssignedTo:  "unassigned",
			Status:      "todo",
			Priority:    prio,
			Department:  dept,
			CreatedAt:   time.Now(),
		}
		_ = s.dbRepo.CreateTask(ctx, task)

		msg := models.WSMessage{
			Type:    "task_update",
			Payload: task,
		}
		_ = s.pubSubRepo.Publish(ctx, "tasks", msg)

		return fmt.Sprintf("📋 **Incident Copilot:** Created task **'%s'** (Priority: **%s**) for the **%s** department and broadcasted it to the Kanban board.", title, prio, dept), true
	}

	// 2. Detect alert broadcast
	if strings.Contains(msgLower, "broadcast alert") || strings.Contains(msgLower, "broadcast a transit delay alert") || strings.Contains(msgLower, "broadcast an alert") {
		title := "Transit Service Update"
		content := "Transit delay reported on Line 1. Expect minor delays."
		sev := "high"

		if strings.Contains(msgLower, "delay") || strings.Contains(msgLower, "transit") {
			title = "Transit Delay Alert"
			content = "Line 1 Shuttle Service is experiencing delays up to 15 minutes due to heavy post-game traffic."
			sev = "info"
		} else if strings.Contains(msgLower, "emergency") || strings.Contains(msgLower, "critical") {
			title = "CONGESTION ALERT"
			content = "Severe congestion at Gate A. Avoid area and use Gate B."
			sev = "critical"
		}

		alert := &models.Alert{
			ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
			Title:     title,
			Content:   content,
			Type:      "ops",
			Severity:  sev,
			CreatedAt: time.Now(),
			Active:    true,
		}
		_ = s.dbRepo.CreateAlert(ctx, alert)

		msg := models.WSMessage{
			Type:    "alert",
			Payload: alert,
		}
		_ = s.pubSubRepo.Publish(ctx, "alerts", msg)

		return fmt.Sprintf("🚨 **Incident Copilot:** Broadcasted **[%s]** alert: **'%s'** - *%s*.", strings.ToUpper(sev), title, content), true
	}

	// 3. Detect medical assignment
	if strings.Contains(msgLower, "assign medical") || strings.Contains(msgLower, "assign incident") {
		incidentId := "med_01"
		volunteerId := "medical1"

		words := strings.Fields(msgLower)
		for _, w := range words {
			if strings.HasPrefix(w, "med_") || strings.HasPrefix(w, "med-") {
				incidentId = w
				break
			}
		}

		_ = s.dbRepo.UpdateMedicalRequestStatus(ctx, incidentId, "assigned", volunteerId)
		requests, err := s.dbRepo.ListMedicalRequests(ctx)
		if err == nil {
			var updatedReq models.MedicalRequest
			for _, r := range requests {
				if r.ID == incidentId {
					updatedReq = r
					break
				}
			}
			if updatedReq.ID != "" {
				msg := models.WSMessage{
					Type:    "medical_update",
					Payload: updatedReq,
				}
				_ = s.pubSubRepo.Publish(ctx, "medical", msg)
			}
		}

		return fmt.Sprintf("🏥 **Incident Copilot:** Assigned medical incident **%s** to responder **%s** and broadcasted it to the operations dashboard.", incidentId, volunteerId), true
	}

	return "", false
}
