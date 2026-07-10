# Orynt Backend

> Go-based API server powering the Orynt Smart Stadium & Tournament Operations Platform. Implements Clean Architecture with dual-mode persistence (Cloud Firestore + in-memory), real-time WebSocket broadcasting via Redis Pub/Sub, JWT authentication with RBAC, and a Retrieval-Augmented Generation (RAG) AI pipeline backed by Google Gemini 2.5 Flash.

---

## Table of Contents

- [Architecture](#architecture)
- [Directory Structure](#directory-structure)
- [Dependencies](#dependencies)
- [Configuration](#configuration)
- [Running Locally](#running-locally)
- [Seeded Data](#seeded-data)
- [API Reference](#api-reference)
  - [Public Endpoints](#public-endpoints)
  - [Staff Endpoints](#staff-endpoints)
  - [Admin Endpoints](#admin-endpoints)
  - [Real-Time WebSocket](#real-time-websocket)
- [Dual-Mode Persistence](#dual-mode-persistence)
- [RAG AI Pipeline](#rag-ai-pipeline)
- [Smart Exit Routing Engine](#smart-exit-routing-engine)
- [Authentication & Authorization](#authentication--authorization)
- [Rate Limiting](#rate-limiting)
- [Middleware Stack](#middleware-stack)
- [Deployment](#deployment)

---

## Architecture

The backend follows **Clean Architecture** with strict dependency inversion. No layer imports from a layer above it. The wiring happens exclusively in `main.go`.

```
main.go (composition root)
   │
   ├── config.LoadConfig()           → Reads .env + OS environment
   │
   ├── repository.NewFirestoreRepository()  → DBRepository interface
   ├── repository.NewRedisRepository()      → PubSubRepository interface
   │
   ├── service.NewAuthService(dbRepo, secret)
   ├── service.NewTournamentService(dbRepo, pubSubRepo)
   ├── service.NewStadiumService(dbRepo, pubSubRepo)
   ├── service.NewOperationsService(dbRepo, pubSubRepo)
   ├── service.NewAIService(dbRepo)
   │
   ├── api.NewWSHub()                → WebSocket connection manager
   ├── api.StartPubSubListener()     → Bridges Pub/Sub → WS broadcast
   │
   ├── api.NewAPIHandler(services...)
   └── api.SetupRouter(handler, hub, pubSub)  → Gin engine with all routes + middleware
```

**Dependency flow:**

```
Handlers (api/) ──depends on──▶ Service Interfaces (service/)
Service Implementations ──depends on──▶ Repository Interfaces (repository/)
Repository Implementations ──depends on──▶ External SDKs (Firestore, Redis)
Models (models/) ──depended on by──▶ all layers (shared domain types)
```

This means:
- Swapping Firestore for PostgreSQL requires changing only `firestore_repo.go` — services and handlers remain untouched.
- Swapping Redis for RabbitMQ requires changing only `redis_repo.go`.
- Every service can be unit-tested by injecting mock repositories.

---

## Directory Structure

```
backend/
├── main.go                              # Composition root — DI wiring, server startup
├── go.mod                               # Module: orynt, Go 1.25
├── go.sum                               # Dependency checksums
├── .env.example                         # Environment variable template
├── .env                                 # Local environment (gitignored)
└── internal/
    ├── api/
    │   ├── handlers.go                  # 30+ HTTP handler functions (682 lines)
    │   │                                  Auth, Tournament, Stadium, Operations, AI, Metrics
    │   ├── routes.go                    # Route registration with 3 permission tiers
    │   │                                  Public → Staff (JWT + RBAC) → Admin (JWT + admin)
    │   ├── middleware.go                # CORSMiddleware, AuthMiddleware, RBACMiddleware, RateLimitMiddleware
    │   └── websocket.go                # WSHub (goroutine-based), Client read/write pumps,
    │                                      PubSub→WS bridge, PublishWSUtility helper
    ├── config/
    │   └── config.go                    # Typed Config struct, godotenv loader, getEnvBool helper
    ├── models/
    │   └── models.go                    # 20 domain structs: User, Stadium, Tournament, Match,
    │                                      CrowdZone, Parking, Transport, Alert, Announcement,
    │                                      FoodStall, MedicalRequest, Volunteer, Task,
    │                                      LostFoundItem, AuditLog, Notification,
    │                                      AIChatRequest, AIChatResponse, WSMessage
    ├── repository/
    │   ├── repository.go               # DBRepository interface (30+ methods)
    │   │                                  PubSubRepository interface (Publish, Subscribe, IncrLimit)
    │   ├── firestore_repo.go           # Dual-mode implementation (1329 lines):
    │   │                                  - Firestore client with gRPC connection
    │   │                                  - Thread-safe MemoryDB fallback (sync.RWMutex)
    │   │                                  - Auto-seeding on empty database
    │   │                                  - Density level auto-computation
    │   └── redis_repo.go               # Dual-mode implementation:
    │                                      - Redis client with Pub/Sub channels
    │                                      - In-memory subscriber map fallback
    │                                      - Rate limit INCR/EXPIRE pipeline
    └── service/
        ├── service.go                   # 5 service interfaces:
        │                                  AuthService, TournamentService, StadiumService,
        │                                  OperationsService, AIService
        ├── ai_service.go               # RAG pipeline: gatherRAGContext() → Gemini API → local fallback
        ├── auth_service.go             # JWT (HS256) generation, bcrypt verification, token validation
        ├── tournament_service.go       # Tournament CRUD, match scoring, live event broadcasting
        ├── stadium_service.go          # Crowd zones, parking, transport, alert lifecycle
        └── ops_service.go              # Task Kanban, lost & found, food stalls, medical dispatch,
                                          announcements with approval workflow, audit logging
```

---

## Dependencies

### Direct Dependencies

| Package | Version | Why It's Here |
|---------|---------|--------------|
| `github.com/gin-gonic/gin` | v1.9.1 | HTTP framework. Gin's middleware chaining model maps cleanly to the CORS → RateLimit → Auth → RBAC pipeline. `binding:"required"` struct tags give us request validation without a separate library. |
| `github.com/golang-jwt/jwt/v5` | v5.2.0 | JWT token management. v5 provides `RegisteredClaims` with typed `NumericDate` for expiration handling. HS256 signing avoids RSA key management overhead for a hackathon deployment. |
| `github.com/gorilla/websocket` | v1.5.1 | WebSocket connections. Gorilla remains the most battle-tested Go WS library. The `Upgrader` with `CheckOrigin` bypass and buffered `NextWriter` pattern enable efficient fan-out to hundreds of concurrent clients. |
| `golang.org/x/crypto` | v0.53.0 | bcrypt password hashing. `bcrypt.DefaultCost` (10 rounds) balances security with login latency (~100ms per hash comparison). |

### Indirect / Infrastructure Dependencies

| Package | Version | Why It's Here |
|---------|---------|--------------|
| `cloud.google.com/go/firestore` | v1.22.0 | Firestore client SDK. Document-level `DataTo()` deserialization maps directly to Go structs via `firestore:""` tags. |
| `github.com/redis/go-redis/v9` | v9.21.0 | Redis client. The `Pipeline()` API enables atomic `INCR` + `EXPIRE` for rate limiting in a single round trip. `Subscribe().Channel()` provides a Go channel–based Pub/Sub consumer. |
| `github.com/joho/godotenv` | v1.5.1 | `.env` file loading. Fails silently when no `.env` exists (production uses OS env). |
| `google.golang.org/api` | v0.287.1 | Google API client libraries (transitive via Firestore SDK). |
| `google.golang.org/grpc` | v1.82.0 | gRPC transport for Firestore (transitive). |
| `go.opentelemetry.io/otel` | v1.43.0 | OpenTelemetry instrumentation (transitive via Firestore/gRPC SDKs). Hooks are present for future Jaeger/Datadog integration. |
| `golang.org/x/time` | v0.15.0 | Time utilities (transitive). |

---

## Configuration

Configuration is loaded by `internal/config/config.go` in this order:

1. Attempt to load `.env` file via `godotenv.Load()` — log and continue if missing.
2. Read OS environment variables with fallback defaults.
3. Return a typed `*Config` struct.

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `PORT` | string | `8080` | HTTP server listen port |
| `JWT_SECRET` | string | `orynt-default-secret-...` | HMAC-SHA256 signing key for JWT tokens |
| `GEMINI_API_KEY` | string | *(empty)* | Google Gemini API key. When empty, the AI service uses the local rule-based fallback — the chatbot still works. |
| `USE_FIRESTORE` | bool | `false` | `true` activates Cloud Firestore. `false` uses thread-safe in-memory maps with pre-seeded data. |
| `FIREBASE_PROJECT_ID` | string | — | GCP project ID (required when `USE_FIRESTORE=true`) |
| `GOOGLE_APPLICATION_CREDENTIALS` | string | — | Path to Firebase service account JSON file |
| `USE_REDIS` | bool | `false` | `true` activates Redis Pub/Sub + rate limiting. `false` uses in-memory subscriber maps and disables distributed rate limiting. |
| `REDIS_URL` | string | `redis://localhost:6379` | Redis connection string (supports Upstash TLS URLs) |
| `REDIS_ADDR` | string | `localhost:6379` | Fallback Redis address if `REDIS_URL` is not set |

**Boolean parsing:** `"true"`, `"1"`, and `"yes"` are treated as true. Everything else (including empty) is false.

---

## Running Locally

```bash
# 1. Clone and enter
cd backend

# 2. Copy environment template
cp .env.example .env

# 3. (Optional) Set your Gemini API key for AI features
#    Leave GEMINI_API_KEY empty for local rule-based chatbot fallback

# 4. Install dependencies
go mod download

# 5. Run
go run main.go
```

**Expected startup logs (in-memory mode):**

```
[CONFIG] Loaded .env file successfully
[ORYNT] Starting Smart Stadium & Tournament Operations Platform...
[ORYNT] Server running at http://localhost:8080
[ORYNT] Real-time WebSockets hub at ws://localhost:8080/api/ws
```

**Health check:**

```bash
curl http://localhost:8080/health
# → {"status":"UP","time":"2026-07-07T21:40:00Z"}
```

---

## Seeded Data

When the database is empty (in-memory mode always, Firestore on first run), the repository auto-seeds realistic operational data:

| Collection | Count | Highlights |
|-----------|-------|-----------|
| **Users** | 6 | 1 admin (`admin/admin123`), 5 staff accounts across volunteer, security, medical, cleaning, ops roles (all `staff123`) |
| **Stadiums** | 1 | "Orynt Arena" — 65,000 capacity, 3 accessible routes, facilities across food/restrooms/medical/exits |
| **Tournaments** | 2 | "National Championship Cup 2026" (active), "Pro League Season Finale" (planned) |
| **Matches** | 2 | Spartans FC vs Titans United (live, 2-1, 67th min with 5 commentary events), Strikers SC vs Wanderers FC (scheduled) |
| **Crowd Zones** | 5 | North Stand (84% — high), East Stand (35% — low), South Stand (78% — medium), West Stand (18% — low), VIP Concourse (96% — high) |
| **Parking** | 4 | Lot A (96% full), Lot B (42%), VIP Deck (97%), Accessible Lot A (28%) |
| **Transport** | 3 | Metro Shuttle (on time), Express Bus (5 min delay), Stadium Link Train (on time, full capacity) |
| **Food Stalls** | 3 | Hotdogs (15 min wait), Brewery (5 min wait), Noodle House (25 min wait) |
| **Tasks** | 3 | Turnstile inspection (todo), Spillage cleanup (in_progress), VIP parking monitor (done) |
| **Alerts** | 2 | Gate A crowd build-up (high), First-aid dispatch Section 104 (critical) |
| **Announcements** | 2 | Free post-match shuttle (public, approved), Staff briefing (staff, approved) |

---

## API Reference

### Public Endpoints

No authentication required.

| Method | Path | Handler | Request Body | Response |
|--------|------|---------|-------------|----------|
| `POST` | `/api/auth/login` | `Login` | `{ username, password }` | `{ accessToken, refreshToken, user }` |
| `GET` | `/api/stadium` | `GetStadium` | — | Stadium object with facilities map |
| `GET` | `/api/stadium/zones` | `GetCrowdZones` | — | `CrowdZone[]` |
| `GET` | `/api/stadium/exit-routing` | `GetExitRoutes` | — | `ExitRoute[]` (smart exit pathways mapping) |
| `GET` | `/api/parking` | `GetParking` | — | `Parking[]` |
| `GET` | `/api/transport` | `GetTransport` | — | `Transport[]` |
| `GET` | `/api/food-stalls` | `GetFoodStalls` | — | `FoodStall[]` |
| `GET` | `/api/alerts?active=true` | `GetAlerts` | — | `Alert[]` (active-only by default) |
| `GET` | `/api/announcements?role=public` | `GetAnnouncements` | — | `Announcement[]` |
| `GET` | `/api/tournaments` | `ListTournaments` | — | `Tournament[]` |
| `GET` | `/api/matches?tournamentId=` | `ListMatches` | — | `Match[]` (optional filter) |
| `POST` | `/api/lost-found` | `ReportLostItem` | `{ itemName, description, category, contactName, contactPhone }` | `LostFoundItem` |
| `GET` | `/api/lost-found` | `ListLostItems` | — | `LostFoundItem[]` |
| `POST` | `/api/medical` | `CreateMedicalRequest` | `{ requester, location, description }` | `MedicalRequest` |
| `POST` | `/api/ai/chat` | `AIChat` | `{ message, role, history: ChatHistoryItem[] }` | SSE stream (chunked text tokens) + final chunk containing sources list |
| `GET` | `/api/sustainability` | `GetSustainabilityMetrics` | — | `{ current: { attendance, energyUsageKw, waterConsumptionL, wasteGeneratedKg }, historical: [{ time, attendance, energyUsageKw, waterConsumptionL, wasteGeneratedKg }], recommendations: string }` |
| `GET` | `/api/ws` | `WebSocketHandler` | — | Upgrades to WebSocket |
| `GET` | `/health` | inline | — | `{ status: "UP" }` |

### Staff Endpoints

Require `Authorization: Bearer <token>` header. RBAC allows: `volunteer`, `security`, `medical`, `cleaning`, `ops` (and `admin` implicitly).

| Method | Path | Handler | Request Body | Description |
|--------|------|---------|-------------|-------------|
| `GET` | `/api/tasks?department=` | `GetTasks` | — | List tasks, optional department filter |
| `POST` | `/api/tasks` | `CreateTask` | `{ title, description, assignedTo, priority, department }` | Create task → publishes `task_update` |
| `PUT` | `/api/tasks/:id` | `UpdateTask` | `{ status }` | Update task status → publishes `task_update` |
| `PUT` | `/api/food-stalls/:id` | `UpdateFoodWaitTime` | `{ waitTime }` | Update stall wait time → publishes `food_stall` |
| `POST` | `/api/alerts` | `BroadcastAlert` | `{ title, content, type, severity }` | Create + broadcast alert → publishes `alert` |
| `GET` | `/api/medical` | `GetMedicalRequests` | — | List all medical requests |
| `PUT` | `/api/medical/:id/assign` | `AssignMedicalRequest` | `{ assignedTo }` | Assign to staff → publishes `medical_update` |
| `PUT` | `/api/medical/:id/resolve` | `ResolveMedicalRequest` | — | Mark resolved → publishes `medical_update` |
| `PUT` | `/api/lost-found/:id/claim` | `ClaimLostItem` | — | Mark claimed → publishes `lost_found_claim` |
| `POST` | `/api/announcements` | `CreateAnnouncement` | `{ title, content, targetAudience }` | Submit for admin approval → publishes `announcement_pending` |

### Admin Endpoints

Require `Authorization: Bearer <token>` with `role: "admin"`.

| Method | Path | Handler | Request Body | Description |
|--------|------|---------|-------------|-------------|
| `POST` | `/api/users` | `CreateUser` | `{ username, password, role, department }` | Create staff account (bcrypt hashed) + audit log |
| `GET` | `/api/users` | `GetUsers` | — | List all system users |
| `POST` | `/api/tournaments` | `CreateTournament` | `{ name }` | Create tournament + audit log |
| `POST` | `/api/matches` | `CreateMatch` | `{ tournamentId, homeTeam, awayTeam, scheduledTime }` | Schedule match + audit log |
| `PUT` | `/api/matches/:id` | `UpdateMatchScore` | `{ homeScore, awayScore, timeElapsed }` | Update score → publishes `match_score` + audit log |
| `POST` | `/api/matches/:id/events` | `AddMatchEvent` | `{ time, type, detail }` | Add commentary → publishes `match_event` + audit log |
| `PUT` | `/api/stadium/zones/:id` | `UpdateCrowdZone` | `{ occupancy }` | Override crowd telemetry → publishes `crowd_density` |
| `PUT` | `/api/parking/:id` | `UpdateParking` | `{ occupied }` | Override parking count → publishes `parking` |
| `PUT` | `/api/transport/:id` | `UpdateTransport` | `{ delay, status }` | Override transport → publishes `transport` |
| `PUT` | `/api/alerts/:id/resolve` | `ResolveAlert` | — | Deactivate alert → publishes `alert_resolved` + audit log |
| `PUT` | `/api/announcements/:id/approve` | `ApproveAnnouncement` | — | Approve → publishes `announcement` + audit log |
| `GET` | `/api/audit-logs` | `GetAuditLogs` | — | All admin action history |
| `GET` | `/api/metrics` | `GetMetrics` | — | Aggregated analytics (attendance, parking %, waste, energy, water, tasks, medical) |

### Real-Time WebSocket

**Endpoint:** `ws://host/api/ws`

The hub runs as a dedicated goroutine (`WSHub.Run()`) managing a `map[*Client]bool` registry. Messages flow one direction: server → client. Clients are read-pumped only to detect disconnections.

**Event channels and message types:**

| Pub/Sub Channel | WS Message `type` | Payload | Triggered By |
|----------------|-------------------|---------|-------------|
| `matches` | `match_score` | Full `Match` object | Admin updates score |
| `matches` | `match_event` | Full `Match` object | Admin adds commentary |
| `stadium` | `crowd_density` | Single `CrowdZone` | Admin overrides occupancy |
| `parking` | `parking` | Single `Parking` | Admin overrides spots |
| `transport` | `transport` | Single `Transport` | Admin overrides status |
| `alerts` | `alert` | Full `Alert` object | Staff broadcasts alert |
| `alerts` | `alert_resolved` | `{ id: string }` | Admin resolves alert |
| `announcements` | `announcement_pending` | Full `Announcement` | Staff submits for approval |
| `announcements` | `announcement` | Full `Announcement` | Admin approves |
| `tasks` | `task_update` | Full `Task` object | Staff creates or updates |
| `lost_found` | `lost_found_report` | Full `LostFoundItem` | Public reports item |
| `lost_found` | `lost_found_claim` | `{ id: string }` | Staff claims item |
| `medical` | `medical_request` | Full `MedicalRequest` | Public requests help |
| `medical` | `medical_update` | Full `MedicalRequest` | Staff assigns/resolves |

**Message envelope:**

```json
{
  "type": "match_score",
  "payload": { "id": "m_01", "homeTeam": "Spartans FC", "homeScore": 3, ... }
}
```

---

## Dual-Mode Persistence

Both the database repository and the Pub/Sub repository operate in dual mode, controlled by `USE_FIRESTORE` and `USE_REDIS` environment flags.

### DBRepository: Firestore ↔ In-Memory

```
NewFirestoreRepository(useFirestore=true, projectID, credsFile)
   │
   ├─ [true] → Connect Firestore client via gRPC
   │            On failure → log, fall back to in-memory
   │            On success → check if `users` collection is empty
   │                         If empty → seedData() into Firestore
   │
   └─ [false] → Initialize MemoryDB (map[string]*Model per collection)
                 → seedData() into memory maps
```

**MemoryDB** uses `sync.RWMutex` for concurrent read/write safety. Every method follows the pattern:

```go
if r.isMemory {
    r.mem.mu.RLock()  // or Lock() for writes
    defer r.mem.mu.RUnlock()
    // operate on r.mem.mapName
    return
}
// Firestore path
```

**Density auto-computation** (in `UpdateCrowdZoneOccupancy`):

```
occupancy / maxCapacity ≥ 0.80 → "high"
occupancy / maxCapacity ≥ 0.40 → "medium"
occupancy / maxCapacity  < 0.40 → "low"
```

### PubSubRepository: Redis ↔ In-Memory

```
NewRedisRepository(useRedis=true, redisURL)
   │
   ├─ [true] → Parse URL → Connect → Ping with 5s timeout
   │            On failure → log, fall back to in-memory
   │
   └─ [false] → In-memory broker: map[string][]func(msg string)
                  Publish() → iterate handlers, call each in a goroutine
                  Subscribe() → append handler to channel's slice
```

**Rate limiting** uses Redis `INCR`/`EXPIRE` pipeline for distributed counting. When Redis is unavailable, `IncrLimit()` returns `(0, nil)`, which signals the rate limit middleware to use its own in-memory sliding window instead.

---

## RAG AI Pipeline

`ai_service.go` implements the full RAG cycle:

### Step 1: Context Retrieval

`gatherRAGContext(ctx, role)` fetches live state from the database:

| Data Source | Always Included | Staff-Only |
|------------|----------------|-----------|
| Matches (scores, events, status) | ✓ | |
| Crowd zones (occupancy, density) | ✓ | |
| Parking (spots, availability) | ✓ | |
| Transport (delays, capacity) | ✓ | |
| Food stalls (wait times, menus) | ✓ | |
| Active alerts | ✓ | |
| Tasks & incidents | | ✓ |

Each data source appends to a `strings.Builder` in markdown format and contributes a source label to the `sources[]` array.

### Step 2: System Prompt Construction

```
You are the Orynt Smart Stadium & Tournament Operations Assistant.
Never hallucinate or guess. Use ONLY the facts provided below.

Current Stadium Live Context:
---
### Live Tournament and Matches:
- Match m_01: Spartans FC vs Titans United. Score: 2-1. Status: live.

### Parking Occupancy:
- Lot p1 (Parking Lot A): 482/500 spots occupied. Type: public.
...
---

User Role: public
```

### Step 3: Multi-Turn Conversation History
The API endpoint `/api/ai/chat` accepts an optional `history` array. The backend maps past exchanges to the correct conversational role (`user` vs `model`) and constructs a continuous dialogue context before calling the model, enabling multi-turn conversation threads.

### Step 4: Gemini Streaming API & SSE
- **Model**: `gemini-2.5-flash`
- **Endpoint**: `https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:streamGenerateContent`
- **Timeout**: 30 seconds
- **Delivery**: Backend processes chunked JSON tokens from the Gemini Stream and forwards them immediately to the client as Server-Sent Events (SSE) using Gin's `c.Stream()` and `http.Flusher`.

### Step 5: Gemini Function Calling (Incident Dispatch Copilot)
If the user's role belongs to the staff console (e.g., admin, volunteer, security, medical, cleaning, ops), the RAG pipeline registers three executable tools with the Gemini API:
1. `create_task(title, description, department, priority)`: Creates a new operational task in the Firestore repository and broadcasts the update to the Kanban board via the `tasks` WebSocket channel.
2. `broadcast_alert(title, content, severity)`: Creates and displays a stadium-wide alert banner, broadcasting the update via the `alerts` WebSocket channel.
3. `assign_medical_incident(incidentId, volunteerId)`: Assigns a pending medical emergency request to a responder, broadcasting the update via the `medical` WebSocket channel.

Gemini decides when to call these tools based on the staff conversation, and the backend handles tool execution and publishes the WS notifications automatically.

### Step 6: Sustainability Eco-Advisor
- **Endpoint**: `/api/sustainability`
- **Logic**: packages current crowd size alongside dynamically computed energy usage (kW), water flow (L), and waste metrics (kg). Passes them to Gemini with a system prompt instructing the model to return 3 highly specific eco-efficiency suggestions.

### Step 7: Fallback Engine

If `GEMINI_API_KEY` is empty or the API returns an error:
1. **Chat/Copilot Fallback**: A local word-by-word streaming generator parses user text for keywords. If tool keywords are found (e.g. "create task"), it invokes `parseAndExecuteLocalTool` to execute local mocks, otherwise it returns telemetry-grounded rule responses.
2. **Eco-Advisor Fallback**: Dynamically selects preset recommendations mapped to low, medium, or high attendance brackets.

The fallback transparently annotates its response with `*(Note: Gemini API returned an error, using local fallback assistant)*` when the API specifically fails (vs. being unconfigured).

---

## Smart Exit Routing Engine

To prevent dangerous crowd crushing when a match concludes or during evacuation, the backend features a smart routing engine that dynamically matches stadium sections to exit gates based on live density telemetry.

- **Endpoint**: `/api/stadium/exit-routing`
- **Algorithm**:
  1. Retrieves all `crowd_zones` representing stadium stands (e.g. North Stand, South Stand) and their current occupancy levels.
  2. Retrieves the configured exit gates (e.g. Gate A, Gate B) and their status.
  3. Maps each section to its closest exit gates.
  4. If a primary exit gate is heavily congested (e.g. >80% capacity or high density in surrounding zones), the routing engine calculates alternative paths to nearby under-utilized exits (e.g., routing North Stand visitors to Gate B instead of Gate A).
  5. Returns a structured mapping of sections, exit directions, coordinates, and recommendations.

The frontend reads this mapping to overlay dynamic, animated glowing paths on the interactive SVG Stadium Map, warning users away from bottlenecks and visually directing them along optimal routes.

---

## Authentication & Authorization

### Token Generation (`auth_service.go`)

**Login flow:**

1. Look up user by username via `DBRepository.GetUserByUsername()`.
2. Compare submitted password against stored bcrypt hash.
3. Generate access token: `Claims{ UserID, Username, Role, Department }` signed with HS256, 12-hour expiry.
4. Generate refresh token: `Claims{ UserID }` signed with HS256, 7-day expiry.
5. Return both tokens + sanitized user object (password omitted via `json:"-"` tag).

**Token validation:**

1. Parse JWT with claims.
2. Verify signature and expiry.
3. Fetch user from database by `claims.UserID` to confirm the account still exists.
4. Return `*models.User` for context injection.

### User Creation

Admin-only. `CreateUser()` hashes the password with `bcrypt.DefaultCost`, prefixes the ID with `u_`, and persists to the database. A duplicate username check exists at the repository level.

---

## Rate Limiting

The `RateLimitMiddleware` implements a dual-strategy rate limiter:

**Strategy 1 — Redis (distributed):**
```
key = "rate:ip:" + clientIP
INCR key → count
EXPIRE key 60s
if count > 120 → HTTP 429
```

**Strategy 2 — In-Memory (single-instance fallback):**
```go
type RateLimiter struct {
    mu           sync.Mutex
    ips          map[string][]time.Time  // sliding window of request timestamps
    limitSeconds int                     // 60
    maxRequests  int                     // 120
}
```

On each request, expired timestamps (older than 60s) are pruned. If the remaining count exceeds 120, the request is rejected with `HTTP 429 Too Many Requests`.

---

## Middleware Stack

Requests pass through middleware in this order:

```
1. gin.Logger()           → Request logging to stdout
2. gin.Recovery()         → Panic recovery with 500 response
3. CORSMiddleware()       → Sets Access-Control-* headers, handles OPTIONS
4. RateLimitMiddleware()  → 120 req/min/IP (Redis or in-memory)
5. [Staff/Admin routes only]:
   5a. AuthMiddleware()   → JWT validation, injects *User into context
   5b. RBACMiddleware()   → Role check against allowed list
```

RBAC enforces that only listed roles can access staff endpoints (`volunteer`, `security`, `medical`, `cleaning`, `ops`). The `admin` role implicitly bypasses all RBAC checks — an admin can access any endpoint.

---

## Deployment

### Build

```bash
# Compile to a single static binary
go build -o orynt .

# Run
./orynt
```

### Render (recommended)

Render auto-detects `go.mod` and builds natively. Set environment variables in the Render dashboard:

- `PORT` — Render assigns this automatically
- `JWT_SECRET` — Generate a strong random secret
- `GEMINI_API_KEY` — Your Google AI Studio key
- `USE_FIRESTORE=true` + `FIREBASE_PROJECT_ID` + `GOOGLE_APPLICATION_CREDENTIALS`
- `USE_REDIS=true` + `REDIS_URL` (Upstash TLS URL)

### Health Check

```
GET /health → { "status": "UP", "time": "..." }
```

Configure Render/load-balancer health probes to hit this endpoint on a 30-second interval.
