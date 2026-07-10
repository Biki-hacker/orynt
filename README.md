# Orynt — Smart Stadium & Tournament Operations Platform

> An AI-powered, real-time stadium management ecosystem built for the **Smart Stadiums & Tournament Operations** hackathon challenge. Orynt combines a Golang backend, a React/TypeScript frontend, Google Gemini–driven RAG AI, WebSocket telemetry, and Firebase/Redis infrastructure into a single coherent platform that serves public visitors, on-ground staff, and venue administrators simultaneously.

---

## Table of Contents

- [Chosen Vertical](#chosen-vertical)
- [Approach & Logic](#approach--logic)
- [How It Works](#how-it-works)
- [Architecture Overview](#architecture-overview)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Core Features](#core-features)
  - [Generative AI Integrations (Google Gemini)](#generative-ai-integrations-google-gemini)
  - [Dynamic Wayfinding & Smart Exit Routing](#dynamic-wayfinding--smart-exit-routing)
  - [Real-Time Telemetry Pipeline](#real-time-telemetry-pipeline)
  - [Role-Based Access & Security](#role-based-access--security)
  - [Public-Facing Fan Experience](#public-facing-fan-experience)
  - [Staff Operations Console](#staff-operations-console)
  - [Admin Command Center](#admin-command-center)
- [Data Model](#data-model)
- [API Surface](#api-surface)
- [Security Implementation](#security-implementation)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Backend Setup](#backend-setup)
  - [Frontend Setup](#frontend-setup)
  - [Environment Variables](#environment-variables)
- [Deployment](#deployment)
- [Assumptions Made](#assumptions-made)
- [Known Limitations](#known-limitations)
- [Future Scalability & Practical Upgrades](#future-scalability--practical-upgrades)

---

## Chosen Vertical

**Smart Stadiums & Tournament Operations** — specifically targeting the intersection of crowd management, real-time navigation, multilingual AI assistance, operational coordination, accessibility, and sustainability monitoring within a live stadium environment.

The platform operates across three distinct user surfaces:

| Surface | Users | Auth Requirement |
|---------|-------|-----------------|
| **Public Dashboard** | Fans, Visitors, General Public | None |
| **Staff Console** | Volunteers, Security, Medical, Cleaning, Ops | JWT (role-scoped) |
| **Admin Command Center** | Stadium Administrators, Event Organizers | JWT (admin-only) |

---

## Approach & Logic

The guiding idea behind Orynt is that a modern stadium is not a static venue — it is a living system with constantly shifting crowd densities, transport delays, food stall queues, medical incidents, and operational tasks. Managing that system requires three things that most stadium software lacks:

1. **A single source of truth.** Crowd sensors, parking feeds, transport APIs, incident reports, and match telemetry all flow into one normalized data layer (Firestore or the in-memory fallback), which every surface reads from in real time.

2. **An AI assistant grounded in that truth.** Rather than a generic chatbot, Orynt's AI assistant runs a Retrieval-Augmented Generation (RAG) pipeline: before answering any question, it fetches the current state of match scores, crowd zones, parking lots, transport lines, food stalls, alerts, and tasks from the database, then injects that context into the Gemini API system prompt. The model is instructed to never hallucinate — it may only respond using the grounded telemetry. This means a visitor asking "Where should I park?" at 7:42 PM gets an answer that reflects the actual occupancy at 7:42 PM, not a canned FAQ response.

3. **Role-aware decision routing.** The same AI endpoint behaves differently for a public visitor (who sees navigation, food, and accessibility info) and a staff member (who additionally sees task assignments, incident queues, and operational recommendations). The backend determines the user's role from their JWT claims (or defaults to `public` for unauthenticated requests) and adjusts the RAG context window accordingly.

The architecture follows **Clean Architecture** principles throughout: repository interfaces abstract storage, service interfaces abstract business logic, and the API handler layer is purely a delivery mechanism. This separation ensures that swapping Firestore for PostgreSQL or Redis for RabbitMQ is a single-file change at the repository boundary.

---

## How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                         FRONTEND (React + Vite)                 │
│                                                                 │
│   Public Dashboard ──── Staff Console ──── Admin Center         │
│         │                     │                  │              │
│    [ REST + WS ]         [ REST + WS ]     [ REST + WS ]       │
└────────┬────────────────────┬─────────────────┬─────────────────┘
         │                    │                 │
         ▼                    ▼                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    BACKEND (Go + Gin Framework)                  │
│                                                                 │
│   ┌──────────┐  ┌──────────────┐  ┌───────────────────────┐    │
│   │ API Layer│  │ Auth + RBAC  │  │  WebSocket Hub        │    │
│   │ Handlers │  │  Middleware   │  │  (Gorilla WebSocket)  │    │
│   └────┬─────┘  └──────┬───────┘  └──────────┬────────────┘    │
│        │               │                      │                 │
│   ┌────▼───────────────▼──────────────────────▼────────────┐   │
│   │                  SERVICE LAYER                          │   │
│   │  AuthService │ TournamentService │ StadiumService       │   │
│   │  OperationsService │ AIService (Gemini RAG)             │   │
│   └────────────────────────┬───────────────────────────────┘   │
│                            │                                    │
│   ┌────────────────────────▼───────────────────────────────┐   │
│   │               REPOSITORY LAYER                          │   │
│   │  DBRepository (Firestore ←→ In-Memory Fallback)        │   │
│   │  PubSubRepository (Redis ←→ In-Memory Pub/Sub)         │   │
│   └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
         │                               │
         ▼                               ▼
  ┌──────────────┐              ┌────────────────┐
  │  Cloud       │              │  Upstash Redis  │
  │  Firestore   │              │  (Pub/Sub +     │
  │  (nam5)      │              │   Rate Limits)  │
  └──────────────┘              └────────────────┘
                    │
                    ▼
         ┌─────────────────┐
         │  Google Gemini   │
         │  (2.5 Flash)     │
         │  via REST API    │
         └─────────────────┘
```

**Request lifecycle (example: visitor asks AI about parking):**

1. Frontend sends `POST /api/ai/chat` with `{ message: "Where should I park?" }`.
2. Gin router hits `AIChat` handler → no JWT required for public chat.
3. `AIService.Chat()` calls `gatherRAGContext()`, which fetches all parking zones, crowd zones, transport lines, food stalls, active alerts, and match data from the `DBRepository`.
4. The fetched data is serialized into a structured markdown context string and injected into Gemini's system instruction alongside a strict "never hallucinate" directive.
5. Gemini returns a grounded answer: *"Parking Lot B (South) has the best availability right now with 460 free spots."*
6. The response includes source attribution tags (e.g., "Smart Parking Sensors") so the user knows the provenance of the information.

If the Gemini API key is not configured or the API call fails, the service transparently falls back to a local rule-based response engine that still uses the same grounded data — ensuring the assistant remains functional offline.

---

## Architecture Overview

### Backend — Clean Architecture in Go

The backend is structured into four distinct layers inside `internal/`:

| Layer | Directory | Responsibility |
|-------|-----------|---------------|
| **Delivery** | `internal/api/` | HTTP handlers, route registration, WebSocket hub, middleware (CORS, Auth, RBAC, Rate Limiting) |
| **Service** | `internal/service/` | Business logic — pure functions that orchestrate repositories and emit Pub/Sub events |
| **Repository** | `internal/repository/` | Data access abstraction — dual-mode (Firestore + thread-safe in-memory) with graceful degradation |
| **Models** | `internal/models/` | Strongly-typed Go structs with both JSON and Firestore struct tags |
| **Config** | `internal/config/` | Environment parsing via `godotenv`, typed config struct with sensible defaults |

Dependencies flow inward: handlers depend on services (interfaces), services depend on repositories (interfaces), and nothing depends on concrete implementations at compile time. The `main.go` wiring function is the only place where concrete types are instantiated.

### Frontend — Feature-Based React Architecture

| Directory | Purpose |
|-----------|---------|
| `src/pages/Public/` | Fan-facing views: Home (live tournament), Match Details, Venue Map, Transport & Parking, Food & FAQ |
| `src/pages/Auth/` | JWT login screen |
| `src/pages/Staff/` | Staff Kanban task board, medical dispatch, incident management |
| `src/pages/Admin/` | Admin analytics, user management, tournament CRUD, telemetry overrides, audit logs |
| `src/components/Chat/` | Floating AI assistant (RAG-grounded chatbot with source attribution) |
| `src/components/Map/` | Interactive SVG stadium map with clickable zones showing crowd density, facilities, and navigation |
| `src/components/ui/` | Reusable design system: Badge, Card, Modal, Skeleton, Spinner, Select, CustomSelect |
| `src/layouts/` | `PublicLayout` (header + nav + emergency banner + chatbot) and `DashboardLayout` (sidebar + RBAC-gated routing) |
| `src/services/` | HTTP API client (generic typed `request<T>` with automatic Bearer injection) and WebSocket service (auto-reconnect with exponential backoff) |
| `src/store/` | Zustand stores: `useAuthStore` (JWT + localStorage persistence) and `useNotificationStore` (real-time toast queue) |
| `src/types/` | Shared TypeScript interfaces mirroring backend models |

---

## Tech Stack

### Backend

| Technology | Version | Purpose |
|-----------|---------|---------|
| **Go** | 1.25 | Core language — chosen for its concurrency model (goroutines for WS hub, Pub/Sub listeners) and single-binary deployment |
| **Gin** | 1.9.1 | HTTP framework — middleware chaining, request validation via `binding` tags, structured JSON responses |
| **Gorilla WebSocket** | 1.5.1 | Full-duplex real-time communication — stadium telemetry broadcast to all connected clients |
| **golang-jwt/jwt v5** | 5.2.0 | JWT access + refresh token generation and validation (HS256 signing) |
| **golang.org/x/crypto** | 0.53.0 | bcrypt password hashing at default cost (10 rounds) |
| **Cloud Firestore SDK** | 1.22.0 | Document database — normalized collections with server-side queries and Firestore struct tags |
| **go-redis v9** | 9.21.0 | Redis Pub/Sub for cross-instance event broadcasting + sliding-window rate limiting via `INCR`/`EXPIRE` pipeline |
| **godotenv** | 1.5.1 | `.env` file loading with OS environment variable fallback |
| **Google Gemini API** | 2.5 Flash | Generative AI endpoint — system instruction–based RAG with structured context injection |
| **OpenTelemetry** | 1.43.0 | Instrumentation hooks (gRPC + HTTP) for future observability integration |

### Frontend

| Technology | Version | Purpose |
|-----------|---------|---------|
| **React** | 18.2 | Component library — hooks-based architecture with strict mode enabled |
| **TypeScript** | 5.2 | Type safety — strict mode with `noUnusedLocals`, `noImplicitReturns`, `noFallthroughCasesInSwitch` |
| **Vite** | 5.1.6 | Build tool — HMR dev server on port 3000 with automatic `/api` proxy to Go backend, sub-second builds |
| **TailwindCSS** | 3.4.1 | Utility-first styling — custom brand palette (blue/indigo), Inter font family, enterprise design tokens |
| **Zustand** | 4.5.2 | State management — two lightweight stores (auth + notifications) without boilerplate |
| **TanStack React Query** | 5.28.4 | Server state caching — automatic refetch, retry logic, and query invalidation |
| **React Router** | 6.22.3 | Client-side routing — nested layouts with `<Outlet>`, dynamic match detail routes |
| **Framer Motion** | 11.0.8 | Animations — page transitions, card hover effects, chatbot slide-in |
| **Recharts** | 2.12.3 | Charting — area, bar, and line charts for admin analytics dashboard |
| **React Hook Form + Zod** | 7.51.1 / 3.22.4 | Form handling — schema-validated login, task creation, incident reporting |
| **Lucide React** | 0.359.0 | Iconography — consistent stroke-based icons across all surfaces |
| **clsx + tailwind-merge** | 2.1.0 / 2.2.2 | Conditional classname composition without conflicts |

### Infrastructure

| Service | Purpose |
|---------|---------|
| **Firebase / Cloud Firestore** | Primary document database (project: `orynt-stadium-2026`, region: `nam5`) |
| **Upstash Redis** | Managed Redis for Pub/Sub event broadcasting and distributed rate limiting |
| **Vercel** | Frontend hosting — automatic deployments from `frontend/dist` |
| **Render** | Backend hosting — Go binary deployment with health check at `/health` |

---

## Project Structure

```
Orynt/
├── backend/
│   ├── main.go                          # Application entry — DI wiring, server bootstrap
│   ├── go.mod / go.sum                  # Go module dependencies
│   ├── .env.example                     # Environment variable template
│   └── internal/
│       ├── api/
│       │   ├── handlers.go              # 30+ REST endpoint handlers (682 lines)
│       │   ├── routes.go                # Route registration — Public / Staff / Admin groups
│       │   ├── middleware.go            # CORS, JWT Auth, RBAC, Rate Limiting middleware
│       │   └── websocket.go            # WS Hub (register/unregister/broadcast), Pub/Sub bridge
│       ├── config/
│       │   └── config.go               # Typed config with env parsing and defaults
│       ├── models/
│       │   └── models.go               # 20 domain structs with JSON + Firestore tags
│       ├── repository/
│       │   ├── repository.go           # DBRepository + PubSubRepository interfaces
│       │   ├── firestore_repo.go       # Dual-mode: Firestore client OR in-memory maps (1329 lines)
│       │   └── redis_repo.go           # Dual-mode: Redis client OR in-memory Pub/Sub broker
│       └── service/
│           ├── service.go              # AuthService, TournamentService, StadiumService, OperationsService, AIService interfaces
│           ├── ai_service.go           # RAG pipeline: context gathering → Gemini API → local fallback
│           ├── auth_service.go         # JWT generation (HS256), bcrypt verification, token validation
│           ├── tournament_service.go   # Tournament CRUD, match scoring, live event commentary
│           ├── stadium_service.go      # Crowd zones, parking, transport, alert broadcasting
│           └── ops_service.go          # Task management, lost & found, food stalls, medical dispatch, audit logs
├── frontend/
│   ├── index.html                       # SPA shell — Inter font, SEO meta, inline SVG favicon
│   ├── package.json                     # 12 dependencies, 10 devDependencies
│   ├── vite.config.ts                   # Dev server (port 3000), API proxy to :8080, WS proxy
│   ├── tailwind.config.js              # Brand palette, Inter font stack, custom neutral scale
│   ├── tsconfig.json                    # Strict TS with bundler module resolution
│   └── src/
│       ├── main.tsx                     # React 18 createRoot + WS auto-connect
│       ├── App.tsx                      # Route tree — Public, Auth, Staff, Admin + 404
│       ├── index.css                    # Tailwind directives + enterprise utility classes
│       ├── types/index.ts              # 15 shared TypeScript interfaces
│       ├── store/index.ts              # Zustand: AuthStore + NotificationStore
│       ├── services/
│       │   ├── api.ts                  # Generic typed HTTP client with Bearer token injection
│       │   └── websocket.ts           # Singleton WS service with subscribe/unsubscribe + reconnect backoff
│       ├── layouts/Layouts.tsx         # PublicLayout + DashboardLayout with toast container
│       ├── components/
│       │   ├── Chat/AIChatbot.tsx      # Floating RAG chatbot — suggested prompts, source attribution, typing indicator
│       │   ├── Map/StadiumMap.tsx      # Interactive SVG stadium map with crowd density overlays
│       │   └── ui/index.tsx            # Badge, Card, Modal, Skeleton, Spinner, Select, CustomSelect
│       └── pages/
│           ├── Public/                  # Home, MatchDetails, VenueMapPage, TransportParking, FoodFAQ
│           ├── Auth/Login.tsx          # JWT login form with Zod validation
│           ├── Staff/StaffDashboard.tsx # Kanban task board, medical dispatch, incident reporting
│           └── Admin/AdminDashboard.tsx # Analytics, user management, tournament ops, audit logs
├── firebase.json                        # Firestore config (default DB, nam5 region)
├── firestore.rules                      # Security rules
├── firestore.indexes.json              # Composite index definitions
├── .firebaserc                          # Firebase project alias (orynt-stadium-2026)
└── .gitignore                           # Secrets, binaries, node_modules, dist
```

---

## Core Features

### Generative AI Integrations (Google Gemini)

Orynt leverages the Google Gemini 2.5 Flash API to deliver three core intelligent services:

1. **Interactive AI Assistant (RAG Chatbot)** — A floating assistant grounded in live stadium context. Uses a custom Retrieval-Augmented Generation (RAG) pipeline to fetch matches, crowd zones, parking, transport, food stalls, and active alerts, compiling them into a structured system context.
2. **Multi-Turn Conversation Memory & SSE Streaming** — The chat maintains a dialog history of the last 10 exchanges for contextual continuity (e.g. asking "Where is the nearest first aid?" followed by "How do I get there?"). Responses stream word-by-word via HTTP Server-Sent Events (SSE) using Gemini's `streamGenerateContent` API, minimizing latency.
3. **Staff Incident Dispatch & Copilot** — For authenticated staff members, Gemini maps natural language instructions (e.g., *"Create a volunteer task to clean up a spill near Food Stall A"*) to structured API actions (e.g., `create_task`, `broadcast_alert`, `assign_medical_incident`) via **Gemini Function Calling**. Tool executions automatically update Firestore and broadcast notifications via WebSockets.
4. **Sustainability Eco-Advisor** — Evaluates live stadium telemetry (attendance, energy in kW, water in L, waste in kg) on the Sustainability page and generates three dynamic, actionable eco-efficiency recommendations for venue managers.
5. **Source Attribution & Fallbacks** — Every chat response displays provenance badges indicating data sources. If the Gemini API is unconfigured or unreachable, the system transparently falls back to an offline rule-based streaming engine.

### Dynamic Wayfinding & Smart Exit Routing

To prevent dangerous crowd crushing at match conclusion or during evacuations, Orynt features a dynamic exit routing engine and visual wayfinding overlay:

- **Routing Engine** — A backend algorithm (`/api/stadium/exit-routing`) that evaluates real-time crowd densities in stadium stands and pairs them dynamically with the closest, under-utilized exit gates.
- **Animated Map Overlays** — The interactive SVG Stadium Map reads the routing mapping to draw glowing, animated path indicators directing fans from congested sections toward the safest exits (e.g., routing traffic to Gate B if Gate A is bottlenecked).
- **Real-Time Synchronization** — Map overlays adapt automatically as zone densities fluctuate, keeping pathways updated in real time via WebSocket broadcasts.

### Real-Time Telemetry Pipeline

```
[State Change] → Service Layer → PubSubRepository.Publish()
                                        │
                    ┌───────────────────┼────────────────────┐
                    │                   │                    │
              [Redis Pub/Sub]    [In-Memory Broker]          │
                    │                   │                    │
                    └───────────┬───────┘                    │
                                │                            │
                    StartPubSubListener()                    │
                                │                            │
                        WSHub.broadcast ←────────────────────┘
                                │
                    ┌───────────┴───────────┐
                    │    │    │    │    │    │
                  [Client1] [Client2] ... [ClientN]
```

Nine dedicated channels carry distinct event types: `stadium`, `parking`, `transport`, `alerts`, `announcements`, `tasks`, `lost_found`, `medical`, and `matches`. The WebSocket service on the frontend subscribes by event type and dispatches payloads to registered React component callbacks. The `WebSocketService` class implements exponential backoff reconnection (1s → 30s cap) and a pub/sub pattern via `subscribe(type, callback) → unsubscribe`.

### Role-Based Access & Security

Three middleware layers protect backend routes:

| Middleware | Applied To | Behavior |
|-----------|-----------|----------|
| `CORSMiddleware` | All routes | Sets permissive CORS headers; handles `OPTIONS` preflight |
| `RateLimitMiddleware` | All routes | 120 req/min per IP — uses Redis `INCR`/`EXPIRE` pipeline when available, falls back to in-memory sliding window with mutex-protected timestamp arrays |
| `AuthMiddleware` | Staff + Admin routes | Parses `Bearer` token from `Authorization` header, validates JWT claims, injects `*models.User` into Gin context |
| `RBACMiddleware` | Staff + Admin routes | Checks `currentUser.Role` against allowed role list; admin role bypasses all constraints |

User accounts are created exclusively by administrators — there is no public registration endpoint. Passwords are hashed with bcrypt at default cost. JWTs carry `userId`, `username`, `role`, and `department` claims with 12-hour expiry. Refresh tokens are issued with 7-day expiry.

### Public-Facing Fan Experience

- **Live Tournament Hub** — Real-time match cards with scores, elapsed time, and live commentary events. Match status badges (`scheduled`, `live`, `completed`) update via WebSocket.
- **Interactive Stadium Map** — SVG-based venue map with clickable zones showing crowd density (color-coded: green/amber/red), facility locations (food, restrooms, medical, exits), and accessible routes.
- **Transport & Parking** — Live parking occupancy bars per zone (public, VIP, accessible), shuttle/bus/train schedules with delay indicators, capacity badges.
- **Food & FAQ** — Food stall directory with wait times, menus, and open/closed status. FAQ section with common visitor questions.
- **Emergency Alert Banner** — Persistent red banner for active critical alerts, auto-updated via WebSocket, dismissible.
- **Lost & Found** — Public submission form for lost items; status tracking (reported → found → claimed).
- **Medical Assistance** — Anonymous emergency help request form (no login required).

### Staff Operations Console

- **Kanban Task Board** — Tasks organized by status (todo / in_progress / done) with drag-to-update via REST `PUT`. Filterable by department. Priority badges (low/medium/high).
- **Medical Dispatch** — Incoming medical request queue with assign/resolve workflow. WebSocket-driven real-time updates.
- **Incident Broadcasting** — Staff can create alerts with type (emergency/weather/crowd/transport) and severity (critical/high/info). Alerts propagate to all connected clients instantly.
- **Announcement Queue** — Staff submit announcements targeting public, staff, or all audiences. Announcements require admin approval before public broadcast.
- **Food Stall Management** — Update wait times in real time; changes propagate to public dashboard.
- **Lost & Found Claims** — Staff can mark found items as claimed.

### Admin Command Center

- **Operations Analytics** — Live metrics dashboard: attendance, peak crowd, parking occupancy percentage, queue lengths, shuttle delays, waste generated (kg), energy usage (kW), water consumption (L), task completion rates, unclaimed lost items, pending medical alerts.
- **User Management** — Create staff accounts with role and department assignment. List all system users.
- **Tournament Management** — Create tournaments, schedule matches, update live scores, add match commentary events (goals, cards, substitutions).
- **Telemetry Overrides** — Admin can manually update crowd zone occupancy, parking counts, and transport status for simulation or manual correction.
- **Announcement Moderation** — Approve or reject staff-submitted announcements. Approved announcements broadcast to all clients.
- **Audit Logs** — Timestamped log of all admin actions (user creation, score updates, alert resolution, announcement approval) with actor identity.

---

## Data Model

The system manages 18 Firestore collections (or in-memory map equivalents):

| Collection | Key Fields | Purpose |
|-----------|-----------|---------|
| `users` | id, username, password (bcrypt), role, department | Staff & admin accounts |
| `stadiums` | id, name, capacity, accessibleRoutes, facilities (nested) | Venue configuration |
| `tournaments` | id, name, status, startDate, endDate | Tournament metadata |
| `matches` | id, tournamentId, homeTeam, awayTeam, scores, status, events[] | Live match tracking |
| `crowd_zones` | id, stadiumId, zoneName, currentOccupancy, maxCapacity, densityLevel | Crowd density sensors |
| `parking` | id, zoneName, totalSpots, occupiedSpots, type | Parking lot telemetry |
| `transport` | id, routeName, mode, nextArrival, delayMinutes, status, capacity | Transit feed |
| `alerts` | id, title, content, type, severity, active | Emergency broadcasting |
| `announcements` | id, title, content, targetAudience, approved, approvedBy | Moderated notices |
| `tasks` | id, title, description, assignedTo, status, priority, department | Staff Kanban |
| `food_stalls` | id, name, location, foodType, status, waitTimeMinutes, menu[] | F&B management |
| `medical` | id, requester, location, description, status, assignedTo | Medical dispatch |
| `lost_found` | id, itemName, description, category, status, contactName, contactPhone | Lost & Found |
| `audit_logs` | id, userId, username, action, resource, details, timestamp | Admin audit trail |

Density levels are automatically computed from occupancy ratios: ≥80% = `high`, ≥40% = `medium`, <40% = `low`.

---

## API Surface

### Public Endpoints (No Auth)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/auth/login` | Authenticate and receive JWT + refresh token |
| `GET` | `/api/stadium` | Venue details, facilities, accessible routes |
| `GET` | `/api/stadium/zones` | Live crowd density per zone |
| `GET` | `/api/stadium/exit-routing` | Live crowd exit pathways mapping (smart exit-routing) |
| `GET` | `/api/parking` | Parking lot occupancy |
| `GET` | `/api/transport` | Transit schedules and delays |
| `GET` | `/api/food-stalls` | Food stall directory with wait times |
| `GET` | `/api/alerts` | Active emergency alerts |
| `GET` | `/api/announcements` | Approved public announcements |
| `GET` | `/api/tournaments` | Tournament listings |
| `GET` | `/api/matches` | Match schedule and live scores |
| `POST` | `/api/lost-found` | Report a lost item |
| `GET` | `/api/lost-found` | Browse lost & found items |
| `POST` | `/api/medical` | Request medical assistance |
| `POST` | `/api/ai/chat` | RAG AI streaming chat query (supports SSE, history context, and function calling tools) |
| `GET` | `/api/sustainability` | Sustainability metrics dashboard telemetry (live & historical) + Gemini Eco-Advisor recommendations |
| `GET` | `/api/ws` | WebSocket upgrade for real-time events |
| `GET` | `/health` | Health check (`{ status: "UP" }`) |

### Staff Endpoints (JWT + RBAC: volunteer, security, medical, cleaning, ops)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/tasks` | List tasks (filterable by `?department=`) |
| `POST` | `/api/tasks` | Create a new task |
| `PUT` | `/api/tasks/:id` | Update task status |
| `PUT` | `/api/food-stalls/:id` | Update food stall wait time |
| `POST` | `/api/alerts` | Broadcast an alert |
| `GET` | `/api/medical` | List medical requests |
| `PUT` | `/api/medical/:id/assign` | Assign medical request to staff |
| `PUT` | `/api/medical/:id/resolve` | Resolve medical request |
| `PUT` | `/api/lost-found/:id/claim` | Mark lost item as claimed |
| `POST` | `/api/announcements` | Submit announcement for approval |

### Admin Endpoints (JWT + RBAC: admin only)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/users` | Create staff account |
| `GET` | `/api/users` | List all users |
| `POST` | `/api/tournaments` | Create tournament |
| `POST` | `/api/matches` | Schedule a match |
| `PUT` | `/api/matches/:id` | Update match score |
| `POST` | `/api/matches/:id/events` | Add match event (goal, card, etc.) |
| `PUT` | `/api/stadium/zones/:id` | Override crowd zone occupancy |
| `PUT` | `/api/parking/:id` | Override parking count |
| `PUT` | `/api/transport/:id` | Override transport status |
| `PUT` | `/api/alerts/:id/resolve` | Resolve an active alert |
| `PUT` | `/api/announcements/:id/approve` | Approve a pending announcement |
| `GET` | `/api/audit-logs` | View admin action history |
| `GET` | `/api/metrics` | Aggregated operations analytics |

---

## Security Implementation

| Measure | Implementation |
|---------|---------------|
| **Password Storage** | bcrypt hashing at default cost (10 rounds), password field omitted from JSON via `json:"-"` tag |
| **JWT Tokens** | HS256-signed access tokens (12h expiry) + refresh tokens (7d expiry), secret loaded from `JWT_SECRET` env |
| **Token Validation** | Claims parsing + database user existence check on every authenticated request |
| **RBAC** | Middleware checks user role against allowed list; admin role implicitly satisfies all constraints |
| **Rate Limiting** | Dual-mode: Redis `INCR`/`EXPIRE` pipeline for distributed limiting, in-memory sliding window (mutex-protected) as fallback — 120 req/min/IP |
| **Input Validation** | Gin's `binding:"required"` tags on all request structs; explicit error responses |
| **CORS** | Configurable origin headers with preflight handling |
| **Secret Management** | All secrets loaded from environment variables; `.env` files excluded from version control via `.gitignore`; Firebase service account keys pattern-matched in gitignore |
| **API Key Protection** | Gemini API key read from `GEMINI_API_KEY` env, never exposed to frontend |
| **No Public Registration** | Users can only be created by admin — prevents unauthorized account creation |
| **Audit Logging** | All privileged admin actions (user creation, score updates, alert resolution) recorded with actor identity and timestamp |

---

## Getting Started

### Prerequisites

- **Go** ≥ 1.25 — [go.dev/dl](https://go.dev/dl)
- **Node.js** ≥ 18 — [nodejs.org](https://nodejs.org)
- **npm** ≥ 9
- **(Optional)** Firebase project with Firestore enabled
- **(Optional)** Upstash Redis instance
- **(Optional)** Google Gemini API key

> The application runs fully offline using in-memory storage and a local AI fallback. External services are only required for production persistence and AI capabilities.

### Backend Setup

```bash
cd backend
cp .env.example .env
# Edit .env — set GEMINI_API_KEY for AI, or leave blank for local fallback

go mod download
go run main.go
# Server starts at http://localhost:8080
# WebSocket hub at ws://localhost:8080/api/ws
```

**Default seeded credentials:**

| Username | Password | Role |
|----------|----------|------|
| `admin` | `admin123` | Admin |
| `volunteer1` | `staff123` | Volunteer |
| `security1` | `staff123` | Security |
| `medical1` | `staff123` | Medical |
| `cleaning1` | `staff123` | Cleaning |
| `ops1` | `staff123` | Operations |

### Frontend Setup

```bash
cd frontend
npm install
cp .env.example .env
npm run dev
# Dev server at http://localhost:3000 with API proxy to :8080
```

### Environment Variables

#### Backend (`backend/.env`)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | HTTP server port |
| `JWT_SECRET` | Yes (prod) | `orynt-default-secret-hackathon-2026-key` | JWT signing key |
| `GEMINI_API_KEY` | No | — | Google Gemini API key for RAG AI |
| `USE_FIRESTORE` | No | `false` | Enable Cloud Firestore |
| `FIREBASE_PROJECT_ID` | If Firestore | — | GCP project ID |
| `GOOGLE_APPLICATION_CREDENTIALS` | If Firestore | — | Path to service account JSON |
| `USE_REDIS` | No | `false` | Enable Redis Pub/Sub |
| `REDIS_URL` | If Redis | `redis://localhost:6379` | Redis connection string |

#### Frontend (`frontend/.env`)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VITE_API_BASE_URL` | No | `http://localhost:8080` | Backend API base |
| `VITE_WS_HOST` | No | `localhost:8080` | WebSocket host |
| `VITE_PORT` | No | `3000` | Dev server port |

---

## Deployment

### Frontend → Vercel

```bash
cd frontend
npm run build
# Deploy `dist/` to Vercel
```

### Backend → Render

The backend compiles to a single Go binary. Render's native Go support detects `go.mod` and builds automatically. Set environment variables in the Render dashboard.

### Firebase

```bash
# Deploy Firestore rules and indexes
firebase deploy --only firestore
```

---

## Assumptions Made

1. **Single-stadium scope** — The current implementation assumes a single venue (`stadium_01`). Multi-venue support is architecturally possible (the stadium ID is already a parameter) but not exercised in the seed data.

2. **IoT data simulation** — Real stadium IoT sensor data (crowd counters, parking barrier feeds) is simulated via admin telemetry override endpoints. In production, these would be replaced by ingestion pipelines from hardware vendors.

3. **Gemini availability** — The RAG pipeline is designed to gracefully degrade. If the Gemini API is unreachable or unconfigured, the local rule-based engine provides contextually relevant responses using the same grounded data.

4. **Authentication simplicity** — Refresh token rotation and secure HTTP-only cookie transport are designed for but simplified for the hackathon demo (tokens returned in JSON body). Production would store refresh tokens in HTTP-only cookies with `SameSite=Strict`.

5. **Firestore rules** — The current Firestore rules (`allow read, write: if true`) are permissive for development. Production deployment would enforce granular access rules per collection and operation.

6. **Seed data represents a snapshot** — The pre-seeded match, crowd, parking, and transport data represents a single moment in time during a live match event. This demonstrates the system's behavior under real operational conditions.

---

## Known Limitations

- **No persistent chat history** — Conversation context is maintained in React component state. Closing the chatbot resets history. Production would persist sessions in Firestore.
- **No push notifications** — Alerts rely on WebSocket + in-app toasts. Production would integrate FCM (Firebase Cloud Messaging) for mobile push.
- **Single-region deployment** — The Firestore instance runs in `nam5`. Global deployments would use multi-region databases.
- **No automated tests** — The architecture supports testing (interfaces enable mocking), but unit and integration tests were not included in the hackathon scope.

---

## Future Scalability & Practical Upgrades

### Near-Term Enhancements

- **Session History Persistence** — Store chat sessions in Firestore rather than local React component memory, allowing users to resume conversations across device refreshes.
- **Push Notification Layer** — Integrate Firebase Cloud Messaging (FCM) for browser and mobile push notifications on critical alerts, matching the priority system already in place (critical/high/info).
- **Predictive Crowd Analytics** — Feed historical crowd zone data into a time-series model (or Gemini's function calling) to predict congestion 15–30 minutes ahead, enabling proactive gate management and exit flow optimization.

### Medium-Term Platform Growth

- **Multi-Stadium / Multi-Venue** — The `stadiumId` foreign key already exists in crowd zones. Extending the UI to a venue selector and deploying separate data partitions per stadium would enable tournament-wide coverage across multiple venues.
- **Multilingual Interface** — The Gemini system prompt can be extended with a language directive (e.g., "Respond in Hindi if the user writes in Hindi"). Combined with `react-i18next` on the frontend, this enables full multilingual support without separate translation files — the AI handles dynamic translation contextually.
- **IoT Ingestion Pipeline** — Replace admin override endpoints with dedicated ingestion APIs that accept bulk sensor readings (crowd counters, parking barriers, turnstile taps) over MQTT or gRPC, with rate-controlled writes to Firestore.
- **SSO Integration** — Replace internal JWT auth with OAuth 2.0 / OpenID Connect via Firebase Auth or Auth0, enabling staff to log in with organizational credentials (Google Workspace, Microsoft Entra ID).

### Long-Term Architectural Evolution

- **Microservice Decomposition** — The current monolith's clean interface boundaries (5 service interfaces, 2 repository interfaces) map directly to independent microservices. The `PubSubRepository` interface already abstracts inter-service communication — migrating to Kafka or Cloud Pub/Sub is a single implementation swap.
- **Horizontal Scaling** — The WebSocket hub currently runs in-process. Migrating to a Redis-backed broadcast channel (already partially implemented via `StartPubSubListener`) enables running multiple backend instances behind a load balancer with shared WS state.
- **Offline-First Mobile App** — The REST + WS API surface is already mobile-ready. A React Native or Flutter wrapper with local SQLite caching would enable stadium navigation and AI assistance even in underground concourses with poor connectivity.
- **Digital Twin** — The interactive SVG stadium map is a foundation for a full 3D digital twin (via Three.js or Mapbox GL), enabling real-time crowd flow visualization, heat mapping, and wayfinding with turn-by-turn navigation.

---

<p align="center">
  Built with precision for the Smart Stadiums & Tournament Operations.
  <br />
  <strong>Orynt</strong> — Where every seat, gate, and shuttle is one query away.
</p>
