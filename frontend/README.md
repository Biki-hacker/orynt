# Orynt Frontend

> React 18 + TypeScript single-page application for the Orynt Smart Stadium & Tournament Operations Platform. Serves three distinct user surfaces — a public fan dashboard, a staff operations console, and an admin command center — unified by a real-time WebSocket telemetry layer, Zustand state management, and a floating RAG-grounded AI chatbot.

---

## Table of Contents

- [Architecture](#architecture)
- [Directory Structure](#directory-structure)
- [Dependencies](#dependencies)
- [Running Locally](#running-locally)
- [Build & Deployment](#build--deployment)
- [Configuration](#configuration)
- [Routing](#routing)
- [State Management](#state-management)
- [Services Layer](#services-layer)
  - [HTTP API Client](#http-api-client)
  - [WebSocket Service](#websocket-service)
- [Design System](#design-system)
  - [Color Palette](#color-palette)
  - [Typography](#typography)
  - [Component Library](#component-library)
  - [CSS Architecture](#css-architecture)
- [Pages & Layouts](#pages--layouts)
  - [Public Layout](#public-layout)
  - [Dashboard Layout](#dashboard-layout)
  - [Public Pages](#public-pages)
  - [Staff Console](#staff-console)
  - [Admin Command Center](#admin-command-cent- [Generative AI Integrations (Google Gemini)](#generative-ai-integrations-google-gemini)
- [Interactive Stadium Map](#interactive-stadium-map)
- [Real-Time Notifications](#real-time-notifications)
- [TypeScript Types](#typescript-types)
- [Accessibility](#accessibility)

---

## Architecture

The frontend is organized by feature, with shared infrastructure (services, store, types, components) at the `src/` root level and page-level code grouped by user role.

```
src/
├── main.tsx                    # React 18 createRoot + WS auto-connect
├── App.tsx                     # Route tree (BrowserRouter + QueryClientProvider)
├── index.css                   # Tailwind directives + enterprise utility classes
│
├── types/index.ts              # 15 shared TypeScript interfaces
├── store/index.ts              # Zustand: AuthStore + NotificationStore
├── services/
│   ├── api.ts                  # Generic typed fetch wrapper with Bearer injection
│   └── websocket.ts            # Singleton WS service (subscribe/reconnect/backoff)
│
├── layouts/Layouts.tsx         # PublicLayout + DashboardLayout + ToastContainer
│
├── components/
│   ├── Chat/AIChatbot.tsx      # Floating AI chatbot (RAG-grounded, source attribution)
│   ├── Map/StadiumMap.tsx      # Interactive SVG venue map with crowd overlays
│   └── ui/index.tsx            # Badge, Card, Modal, Skeleton, Spinner, Select, CustomSelect
│
└── pages/
    ├── Public/
    │   ├── Home.tsx            # Live tournament hub (match cards, crowd, alerts)
    │   ├── MatchDetails.tsx    # Individual match view with live commentary
    │   ├── VenueMapPage.tsx    # Stadium map wrapper page
    │   ├── TransportParking.tsx # Transport schedules + parking occupancy
    │   └── FoodFAQ.tsx         # Food stalls + FAQ section
    ├── Auth/Login.tsx          # JWT login form with validation
    ├── Staff/StaffDashboard.tsx # Kanban tasks, medical dispatch, incidents
    └── Admin/AdminDashboard.tsx # Analytics, user mgmt, tournament ops, audit logs
```

**Data flow pattern:**

```
[User Action] → api.post/put/get() → Backend REST API
                                           │
                                      [State Change]
                                           │
                              PubSub → WebSocket Hub → wsService.onmessage()
                                                            │
                                          ┌─────────────────┼──────────────────┐
                                          │                 │                  │
                               NotificationStore    Component Callbacks    Layout Alerts
                               (toast queue)        (via subscribe())     (banner updates)
```

---

## Directory Structure

```
frontend/
├── index.html                   # SPA shell
│                                  - Inter font from Google Fonts (preconnected)
│                                  - Inline SVG favicon (stadium trophy icon)
│                                  - SEO meta description
│                                  - Tailwind body classes (antialiased, Inter)
├── package.json                 # 12 dependencies, 10 devDependencies
├── package-lock.json            # Deterministic installs
├── vite.config.ts               # Dev server port 3000, /api proxy to :8080, WS proxy
├── tailwind.config.js           # Brand palette, Inter font stack, custom neutral scale
├── postcss.config.js            # autoprefixer + tailwindcss plugins
├── tsconfig.json                # Strict TS: noUnusedLocals, noImplicitReturns, etc.
├── .env.example                 # Environment variable template
└── src/
    ├── main.tsx                 # (14 lines) createRoot + StrictMode + wsService.connect()
    ├── App.tsx                  # (59 lines) Route tree with QueryClient config
    ├── index.css                # (56 lines) Tailwind base + scrollbar + animation + utility classes
    ├── vite-env.d.ts            # Vite client type reference
    ├── types/index.ts           # (184 lines) 15 TypeScript interfaces mirroring backend models
    ├── store/index.ts           # (71 lines) AuthStore + NotificationStore (Zustand)
    ├── services/
    │   ├── api.ts               # (59 lines) Generic request<T> with Bearer injection
    │   └── websocket.ts         # (136 lines) WebSocketService class with pub/sub + reconnect
    ├── layouts/
    │   └── Layouts.tsx          # (331 lines) PublicLayout, DashboardLayout, ToastContainer
    ├── components/
    │   ├── Chat/AIChatbot.tsx   # (222 lines) Floating chatbot with suggested prompts
    │   ├── Map/StadiumMap.tsx   # (300+ lines) SVG interactive stadium map
    │   └── ui/index.tsx         # (204 lines) 7 reusable UI primitives
    └── pages/
        ├── Public/
        │   ├── Home.tsx            # (280+ lines) Tournament hub, crowd density, alerts
        │   ├── MatchDetails.tsx    # (170+ lines) Live match view with commentary timeline
        │   ├── VenueMapPage.tsx    # (70 lines) Map wrapper with accessibility info
        │   ├── TransportParking.tsx # (170+ lines) Transport table + parking bars
        │   └── FoodFAQ.tsx         # (170+ lines) Food stalls + FAQ accordion
        ├── Auth/
        │   └── Login.tsx           # (150+ lines) Login form with error handling
        ├── Staff/
        │   └── StaffDashboard.tsx  # (530+ lines) Kanban + medical + incidents
        └── Admin/
            └── AdminDashboard.tsx  # (570+ lines) Analytics + user mgmt + tournament ops
```

---

## Dependencies

### Runtime Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `react` | 18.2.0 | Component library — hooks-based, StrictMode enabled for dev double-rendering checks |
| `react-dom` | 18.2.0 | DOM renderer with `createRoot` API |
| `react-router-dom` | 6.22.3 | Client-side routing — nested `<Route>` with `<Outlet>`, dynamic params (`:id`), layout routes |
| `@tanstack/react-query` | 5.28.4 | Server state management — automatic caching, background refetch, retry (1 attempt), disabled refetch-on-focus |
| `zustand` | 4.5.2 | Client state — two stores with no boilerplate. Auth store persists to `localStorage` for session survival across refreshes. |
| `framer-motion` | 11.0.8 | Animations — page transitions, card hover lift effects, chatbot slide-in, toast entry animations |
| `recharts` | 2.12.3 | Data visualization — area charts, bar charts, and line charts for admin analytics dashboard |
| `react-hook-form` | 7.51.1 | Form management — uncontrolled inputs with validation. Used in Login, task creation, incident reporting. |
| `zod` | 3.22.4 | Schema validation — login form field validation, structured error messages |
| `lucide-react` | 0.359.0 | Icon library — 30+ icons used: `MessageSquare`, `Send`, `Sparkles`, `ShieldAlert`, `LogOut`, `Menu`, `X`, `CheckCircle2`, `Info`, `User`, `ChevronDown`, `Check`, etc. |
| `clsx` | 2.1.0 | Conditional className joining — e.g., `clsx('base', isActive && 'active')` |
| `tailwind-merge` | 2.2.2 | Tailwind class conflict resolution — prevents `p-4` + `p-2` ambiguity in composed components |

### Dev Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `vite` | 5.1.6 | Build tool — sub-second HMR, native ESM, optimized production bundles |
| `@vitejs/plugin-react` | 4.2.1 | React Fast Refresh + JSX transform for Vite |
| `typescript` | 5.2.2 | Type checker — strict mode, bundler module resolution, `react-jsx` transform |
| `tailwindcss` | 3.4.1 | Utility-first CSS framework — JIT compilation, custom design token system |
| `postcss` | 8.4.38 | CSS pipeline — processes Tailwind directives |
| `autoprefixer` | 10.4.19 | Vendor prefix injection for cross-browser compatibility |
| `@types/react` | 18.2.66 | React type definitions |
| `@types/react-dom` | 18.2.22 | ReactDOM type definitions |
| `@typescript-eslint/eslint-plugin` | 7.2.0 | TypeScript-aware ESLint rules |
| `@typescript-eslint/parser` | 7.2.0 | TypeScript ESLint parser |

---

## Running Locally

```bash
# 1. Install dependencies
npm install

# 2. Copy environment template
cp .env.example .env

# 3. Start dev server
npm run dev
# → http://localhost:3000
```

**The Vite dev server automatically proxies:**

| Frontend Request | Backend Target |
|-----------------|---------------|
| `/api/*` | `http://localhost:8080/api/*` |
| `/api/ws` (WebSocket) | `ws://localhost:8080/api/ws` |

This proxy configuration lives in `vite.config.ts`:

```ts
server: {
  port: 3000,
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
      ws: true          // enables WebSocket proxy
    }
  }
}
```

---

## Build & Deployment

```bash
# Type-check + production build
npm run build
# Outputs to frontend/dist/

# Preview production build locally
npm run preview
```

**Deploy `dist/` to Vercel:**

Vercel auto-detects the Vite project. Set the build command to `npm run build` and the output directory to `dist`. Configure `VITE_API_BASE_URL` and `VITE_WS_HOST` in the Vercel environment variables to point at the production backend.

---

## Configuration

All environment variables are prefixed with `VITE_` so Vite exposes them to client-side code via `import.meta.env`.

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_BASE_URL` | `http://localhost:8080` | Backend API base URL (used by Vite proxy config, not directly by fetch calls since `/api` paths are relative) |
| `VITE_WS_HOST` | `localhost:8080` | WebSocket host. The `WebSocketService` reads this to construct the `ws://` or `wss://` URL. Falls back to `localhost:8080` when the dev server runs on port 3000. |
| `VITE_PORT` | `3000` | Dev server port |

---

## Routing

Defined in `App.tsx` with nested layout routes:

```
/ ........................ PublicLayout
├── / .................... Home (live tournament hub)
├── /match/:id ........... MatchDetails (live commentary)
├── /map ................. VenueMapPage (interactive SVG map)
├── /transport ........... TransportParking
├── /food-faq ............ FoodFAQ
├── /sustainability ...... Sustainability (energy, water, waste + Gemini Eco-Advisor)
└── /login ............... Login

/staff ................... DashboardLayout (JWT-gated)
└── /staff ............... StaffDashboard (Kanban + medical)

/admin ................... DashboardLayout (JWT-gated)
└── /admin ............... AdminDashboard (analytics + ops)

/* ....................... 404 Not Found fallback
```

**Route protection:** `DashboardLayout` checks `useAuthStore.getState().isAuthenticated` on mount. If `false`, it redirects to `/login` via `useNavigate()`.

---

## State Management

Two Zustand stores, each under 40 lines:

### `useAuthStore`

```typescript
interface AuthState {
  user: User | null
  accessToken: string | null
  isAuthenticated: boolean
  setAuth(user, token): void   // persists to localStorage
  clearAuth(): void            // clears localStorage + resets state
}
```

**Persistence:** On store creation, the initializer reads `orynt_user` and `orynt_token` from `localStorage`. This means a browser refresh does not log the user out — the JWT and user object survive until `clearAuth()` is called.

### `useNotificationStore`

```typescript
interface NotificationState {
  notifications: ToastMessage[]
  addNotification(notification): void   // prepends, caps at 8 items
  removeNotification(id): void
  clearAll(): void
}
```

**Toast types:** `emergency`, `weather`, `crowd`, `transport`, `info`
**Severity levels:** `critical` (red border), `high` (amber border), `info` (blue border)

---

## Services Layer

### HTTP API Client

`services/api.ts` exports a generic typed client:

```typescript
api.get<T>(path)              // GET  /api{path}
api.post<T>(path, body?)      // POST /api{path}
api.put<T>(path, body?)       // PUT  /api{path}
api.delete<T>(path)           // DELETE /api{path}
```

**Features:**
- Automatic `Bearer` token injection from `useAuthStore.getState().accessToken`
- Automatic `Content-Type: application/json` for non-FormData bodies
- Structured error parsing — extracts `error` field from JSON error responses
- `204 No Content` handling — returns empty object instead of parsing empty body
- Relative `/api` base path — works with both Vite proxy (dev) and production reverse proxy

### WebSocket Service

`services/websocket.ts` exports a singleton `wsService`:

```typescript
wsService.connect()                    // Initiate or resume connection
wsService.subscribe(type, callback)    // Returns unsubscribe function
wsService.disconnect()                 // Clean shutdown
```

**Connection management:**
- URL construction: auto-detects `wss:` for HTTPS, reads `VITE_WS_HOST` or falls back to `localhost:8080` when dev server is on port 3000
- Exponential backoff reconnection: starts at 1s, doubles on each failure, caps at 30s, resets on successful connection

**Message dispatch:**
- Parses incoming JSON `{ type, payload }`
- Dispatches to type-specific subscriber callbacks
- Additionally dispatches to wildcard (`*`) subscribers
- Auto-triggers `NotificationStore.addNotification()` for `alert`, `announcement`, and `medical_request` message types

**Lifecycle:**
- `connect()` is called in `main.tsx` on application boot
- `connect()` is also called in both `PublicLayout` and `DashboardLayout` `useEffect` hooks to ensure reconnection after navigation
- `subscribe()` returns a cleanup function, making it compatible with React's `useEffect` teardown pattern

---

## Design System

### Color Palette

Defined in `tailwind.config.js` under `theme.extend.colors`:

| Token | Hex | Usage |
|-------|-----|-------|
| `brand-50` | `#f5f7ff` | AI chatbot suggested prompt backgrounds |
| `brand-100` | `#ebf0ff` | Text selection highlight, hover states |
| `brand-500` | `#3b82f6` | Primary blue — buttons, active links, WS indicator badges |
| `brand-600` | `#2563eb` | Button backgrounds, primary CTAs |
| `brand-700` | `#1d4ed8` | Button hover states |
| `brand-accent` | `#4f46e5` | Indigo accent — used sparingly for emphasis |
| `neutral-50` | `#fafafa` | Page backgrounds (off-white) |
| `neutral-100` | `#f4f4f5` | Dashboard content area backgrounds |
| `neutral-800` | `#18181b` | Primary text (near-black) |

The palette avoids neon, glassmorphism, and random gradients — targeting the visual language of enterprise tools like Linear, Vercel Dashboard, and Stripe.

### Typography

```js
fontFamily: {
  sans: ['Inter', 'Geist', 'system-ui', 'sans-serif']
}
```

Inter is loaded via Google Fonts (`<link>` in `index.html`) with weights 300–700. The font stack falls back to Geist → system-ui to ensure consistent rendering on all platforms.

### Component Library

`components/ui/index.tsx` exports 7 reusable primitives:

| Component | Props | Description |
|-----------|-------|-------------|
| `Badge` | `variant: info\|success\|warning\|error\|neutral` | Color-coded status labels (e.g., match status, priority, density level) |
| `Card` | `className?, onClick?` | White card with zinc border, subtle shadow, hover shadow lift. Optional click handler adds pointer cursor. |
| `Modal` | `isOpen, onClose, title` | Centered overlay with backdrop blur (`bg-zinc-950/40`), border, max-height scroll. Close button in header. |
| `Skeleton` | `className?` | Pulsing `bg-zinc-200` loading placeholder. Used in data-fetching states. |
| `Spinner` | `size: sm\|md\|lg` | Rotating border spinner with brand-500 accent. Three size variants. |
| `Select` | `label?, ...HTMLSelectAttributes` | Native `<select>` with Chevron icon overlay, brand focus ring, consistent styling. Uses `forwardRef`. |
| `CustomSelect` | `label?, value, onChange, options[]` | Custom dropdown with floating overlay list, check icon for selected item, click-outside dismiss. Zero dependency. |

### CSS Architecture

`index.css` layers:

1. **Tailwind Directives** — `@tailwind base/components/utilities`
2. **Base Layer** — Body applies `bg-neutral-50 text-neutral-800 antialiased selection:bg-brand-100`
3. **Custom Scrollbar** — 6px width, zinc-colored thumb, rounded, transparent track
4. **Animations** — `pulse-slow` keyframe for live score indicator blink
5. **Enterprise Utility Classes:**
   - `.enterprise-card` — White card with zinc border, rounded, shadow + hover shadow
   - `.enterprise-input` — Consistent input styling with brand focus border
   - `.enterprise-btn-primary` — Brand-600 button with focus ring offset
   - `.enterprise-btn-secondary` — White button with zinc border, neutral text

---

## Pages & Layouts

### Public Layout

`layouts/Layouts.tsx → PublicLayout`

**Structure:**
```
┌──────────────────────────────────────────────────┐
│ [Emergency Alert Banner] (red, pulsing, dismiss) │  ← active alerts from /api/alerts
├──────────────────────────────────────────────────┤
│ Header: Logo │ Nav Links │ Staff Access / Logout │  ← sticky, responsive
├──────────────────────────────────────────────────┤
│                                                  │
│          <Outlet /> (page content)               │
│                                                  │
├──────────────────────────────────────────────────┤
│ [AI Chatbot FAB] ─── bottom-right floating       │
│ [Toast Notifications] ── bottom-left stack       │
└──────────────────────────────────────────────────┘
```

**Real-time behavior:**
- On mount: fetches active alerts, connects WebSocket, subscribes to `alert` and `alert_resolved` events
- New alerts append to the banner; resolved alerts are removed from state
- Mobile nav: hamburger menu toggles a slide-down drawer

**Navigation links:** Live Tournament, Stadium Map, Transport & Parking, Food & FAQ

### Dashboard Layout

`layouts/Layouts.tsx → DashboardLayout`

**Structure:**
```
┌────────────────┬────────────────────────────────────┐
│   Sidebar      │  Header: Title │ IoT Status │ Link │
│  ┌──────────┐  ├────────────────────────────────────┤
│  │ Logo     │  │                                    │
│  │ User     │  │         <Outlet /> (page content)  │
│  │ Nav      │  │                                    │
│  │ Sign Out │  │                                    │
│  └──────────┘  │                                    │
└────────────────┴────────────────────────────────────┘
```

- **Auth gate:** Redirects to `/login` if `isAuthenticated` is false
- **Sidebar:** Dark zinc-900 with brand accent active state (left border), user avatar, role/department badges
- **Header:** Shows "Administration Operations" or "Staff Command" based on role. Green IoT indicator badge.
- **Admin nav:** "Operations Analytics" + "Task Coordination"
- **Staff nav:** "My Action Board"

### Public Pages

| Page | File | Key Features |
|------|------|-------------|
| **Home** | `Home.tsx` | Live match cards with scores, elapsed time, status badges. Crowd density zone summary. Active emergency alerts. Quick-nav cards to other sections. |
| **Match Details** | `MatchDetails.tsx` | Single match view with live commentary timeline (goals, cards, substitutions). Score display with team names. `useParams()` for match ID. |
| **Venue Map** | `VenueMapPage.tsx` | Wraps `StadiumMap` component. Displays accessible routes info. |
| **Transport & Parking** | `TransportParking.tsx` | Transport routes table (mode icon, delay, status badge, capacity). Parking occupancy bars with percentage and type badges. |
| **Food & FAQ** | `FoodFAQ.tsx` | Food stall cards with name, location, wait time, menu items, open/closed status. FAQ accordion section. |
| **Sustainability** | `Sustainability.tsx` | Displays live power usage (kW), water consumption (L), and waste generated (kg) metrics with interactive Recharts line graphs and an eco-advisor panel powered by Gemini RAG recommendations. |

### Staff Console

**`StaffDashboard.tsx`** (~530 lines) — the operational nerve center:

- **Kanban Board** — Three columns (Todo / In Progress / Done) populated from `GET /api/tasks`. Status updates via `PUT /api/tasks/:id`. Department filter dropdown.
- **Medical Dispatch** — Incoming requests table. Assign/resolve buttons trigger `PUT /api/medical/:id/assign` and `.../resolve`.
- **Alert Broadcasting** — Form to create new alerts with title, content, type (dropdown), severity (dropdown). Submits to `POST /api/alerts`.
- **Announcement Submission** — Form with title, content, target audience selector. Submits to `POST /api/announcements` (queued for admin approval).
- **Lost & Found** — List of reported items with claim buttons for staff.

### Admin Command Center

**`AdminDashboard.tsx`** (~570 lines) — full operational oversight:

- **Metrics Dashboard** — Fetches from `GET /api/metrics`. Displays attendance, peak crowd, parking %, completed tasks, pending medical, waste/energy/water metrics.
- **User Management** — Create user form (username, password, role selector, department). Users list table.
- **Tournament Management** — Create tournament, create match, update match score, add match events (goals, cards).
- **Telemetry Overrides** — Update crowd zone occupancy, parking counts, transport status. Useful for simulation and testing.
- **Announcement Moderation** — Approve pending announcements (triggers public broadcast).
- **Audit Logs** — Chronological table of all admin actions fetched from `GET /api/audit-logs`.

---

## Generative AI Integrations (Google Gemini)

Orynt leverages the Google Gemini 2.5 Flash API to deliver three core intelligent services:

### 1. Interactive AI Assistant (RAG Chatbot)
The chatbot (`components/Chat/AIChatbot.tsx`) is a persistent floating assistant grounded in live stadium context.
- **Real-Time Text Streaming**: Fetches streaming text responses word-by-word via Server-Sent Events (SSE) from the backend `/api/ai/chat` endpoint.
- **Multi-Turn Conversation Memory**: Retains state of past exchanges (last 10 messages) in React state and passes it to the backend contents history array, allowing fans to query contextually (e.g. "Where is the nearest first aid?" followed by "How do I get there?").
- **Source Attribution**: Visualizes sources badges (e.g., "Smart Parking Sensors", "Transport API Feed") at the bottom of the message bubble for complete grounding transparency.

### 2. Staff Incident Dispatch & Copilot
For authenticated staff members (volunteers, security, medical, etc.), the AI chatbot behaves as a functional Copilot:
- **Gemini Function Calling (Structured Tool Use)**: Parses natural language intents to execute specific tools directly.
- **Action Execution**: Allows commands like *"Create a cleaning task to clean up a spill near Food Stall A"* or *"Broadcast a transit delay alert for Metro Line 1"* to automatically call `create_task` or `broadcast_alert` on the backend, update Firestore, and broadcast state changes via WebSockets.

### 3. Sustainability Eco-Advisor
Integrated into the **Sustainability Dashboard** (`pages/Public/Sustainability.tsx`):
- **Live Analysis**: Pulls current stadium telemetry (attendance, energy usage in kW, water flow in Liters, waste in kilograms).
- **Dynamic Optimization Tips**: Sends this telemetry context to Gemini, generating three highly targeted, actionable recommendations (e.g. recommending dimming lights in empty zones, adjusting water pressure, or optimizing cleanup shifts).
- **Graceful Fallback**: If the Gemini API is offline, it drops back to structured heuristics based on attendance thresholds.

---

## Interactive Stadium Map

`components/Map/StadiumMap.tsx` renders a custom SVG stadium layout:

- **Clickable zones** — Each stadium section (North Stand, East Stand, South Stand, West Stand, VIP Concourse) is an interactive SVG path
- **Crowd density overlays** — Zones are color-coded based on `densityLevel`: green (low), amber (medium), red (high)
- **Facility markers** — Food stalls, restrooms, medical stations, emergency exits displayed as labeled icons
- **Accessible routes** — Highlighted paths for wheelchair access (elevators, ramps, priority exits)
- **Dynamic Wayfinding & Smart Exit Routing** — When a user selects a congested zone, or when a tournament match is completed, the map queries the backend exit routing service and overlays glowing animated paths directing fans toward the safest, under-utilized gate (e.g., directing traffic to Gate B rather than Gate A to avoid crushes).
- **Real-time updates** — Subscribes to `crowd_density` WebSocket events to update zone colors without page refresh
- **Tooltip/detail panels** — Clicking a zone shows occupancy numbers, nearby facilities, and navigation hints

---

## Real-Time Notifications

The notification system spans three layers:

1. **WebSocket Service** (`websocket.ts`) — Intercepts `alert`, `announcement`, and `medical_request` message types and pushes them to `NotificationStore`
2. **NotificationStore** (Zustand) — Maintains a capped queue of 8 toast messages with type, severity, and timestamp
3. **ToastContainer** (`Layouts.tsx`) — Renders toasts in the bottom-left corner with severity-coded left borders:
   - Critical: red left border + `ShieldAlert` icon
   - High: amber left border + `ShieldAlert` icon
   - Info: blue left border + `Info` icon
   - Slide-in animation from left
   - Individual dismiss buttons

Additionally, `PublicLayout` maintains a separate **Emergency Alert Banner** at the top of the page for active critical alerts, which is dismiss-per-alert and updates in real time via WebSocket subscriptions.

---

## TypeScript Types

`types/index.ts` defines 15 interfaces that mirror the backend's Go structs:

| Interface | Key Fields | Used By |
|-----------|-----------|---------|
| `User` | id, username, role (union type of 8 roles), department | Auth store, dashboard sidebar, user management |
| `TokenResponse` | accessToken, refreshToken, user | Login response handling |
| `Stadium` | id, name, capacity, accessibleRoutes, facilities (nested) | Venue map, stadium details |
| `Tournament` | id, name, status (planned\|active\|completed), dates | Tournament listing, admin CRUD |
| `Match` | id, teams, scores, status (scheduled\|live\|completed), events[] | Match cards, live commentary |
| `MatchComment` | time, type (goal\|card\|substitution\|info), detail | Commentary timeline |
| `CrowdZone` | id, zoneName, currentOccupancy, maxCapacity, densityLevel | Map overlays, home page summary |
| `Parking` | id, zoneName, totalSpots, occupiedSpots, type | Parking bars |
| `Transport` | id, routeName, mode, nextArrival, delayMinutes, status, capacity | Transport table |
| `Alert` | id, title, content, type, severity, active | Emergency banner, toast notifications |
| `Announcement` | id, title, content, targetAudience, approved, approvedBy | Public notices, admin moderation |
| `FoodStall` | id, name, location, foodType, status, waitTimeMinutes, menu[] | Food directory |
| `MedicalRequest` | id, requester, location, description, status, assignedTo | Medical dispatch queue |
| `Task` | id, title, description, assignedTo, status, priority, department | Kanban board |
| `LostFoundItem` | id, itemName, description, category, status, contacts | Lost & found list |
| `AuditLog` | id, userId, username, action, resource, details, timestamp | Admin audit table |
| `LiveTelemetryMetrics` | attendance, peakCrowd, parking, waste, energy, water, tasks, medical | Admin analytics dashboard |
| `AIChatResponse` | response, sources[] | Chatbot message rendering |
| `SustainabilityMetrics` | current, historical, recommendations | Sustainability dashboard |
| `SustainabilityHistoricalPoint` | time, attendance, energyUsageKw, waterConsumptionL, wasteGeneratedKg | Sustainability timeline |

All union types (e.g., `status: 'scheduled' | 'live' | 'completed'`) match the backend's string constants exactly, enabling compile-time safety on conditional rendering logic.

---

## Accessibility

- **Semantic HTML** — Proper `<header>`, `<main>`, `<nav>`, `<aside>` elements throughout layouts
- **Keyboard navigation** — All interactive elements (buttons, links, form inputs) are focusable with visible focus rings (`focus:ring-2 focus:ring-brand-500`)
- **Color contrast** — Text uses `neutral-800` (#18181b) on `neutral-50` (#fafafa) backgrounds, exceeding WCAG AA contrast ratios
- **Responsive design** — Desktop-first with mobile breakpoints. Hamburger menu on small screens. Dashboard sidebar collapses.
- **Screen reader support** — `aria` attributes on interactive elements, meaningful link text, form labels
- **Accessible routes** — Stadium map highlights wheelchair-accessible paths (elevators, ramps, priority exits) and the chatbot can provide accessibility-specific guidance
- **Anti-aliased text** — `antialiased` class applied to `<body>` for smooth font rendering across displays
- **Reduced motion** — Animations use CSS transitions that respect `prefers-reduced-motion` via Tailwind's built-in support
