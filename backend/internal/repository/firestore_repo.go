package repository

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
	"orynt/internal/models"

	"golang.org/x/crypto/bcrypt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MemoryDB holds in-memory data tables protected by mutexes
type MemoryDB struct {
	mu            sync.RWMutex
	users         map[string]*models.User
	tournaments   map[string]*models.Tournament
	matches       map[string]*models.Match
	stadiums      map[string]*models.Stadium
	crowdZones    map[string]*models.CrowdZone
	parkingZones  map[string]*models.Parking
	transport     map[string]*models.Transport
	alerts        map[string]*models.Alert
	announcements map[string]*models.Announcement
	tasks         map[string]*models.Task
	lostFound     map[string]*models.LostFoundItem
	foodStalls    map[string]*models.FoodStall
	medical       map[string]*models.MedicalRequest
	auditLogs     []models.AuditLog
}

// FirestoreRepository implements DBRepository, uses MemoryDB as fallback
type FirestoreRepository struct {
	isMemory bool
	mem      *MemoryDB
	client   *firestore.Client
}

// NewFirestoreRepository initializes database repository
func NewFirestoreRepository(useFirestore bool, projectID string, credsFile string) *FirestoreRepository {
	repo := &FirestoreRepository{
		isMemory: !useFirestore,
		mem: &MemoryDB{
			users:         make(map[string]*models.User),
			tournaments:   make(map[string]*models.Tournament),
			matches:       make(map[string]*models.Match),
			stadiums:      make(map[string]*models.Stadium),
			crowdZones:    make(map[string]*models.CrowdZone),
			parkingZones:  make(map[string]*models.Parking),
			transport:     make(map[string]*models.Transport),
			alerts:        make(map[string]*models.Alert),
			announcements: make(map[string]*models.Announcement),
			tasks:         make(map[string]*models.Task),
			lostFound:     make(map[string]*models.LostFoundItem),
			foodStalls:    make(map[string]*models.FoodStall),
			medical:       make(map[string]*models.MedicalRequest),
			auditLogs:     make([]models.AuditLog, 0),
		},
	}

	if useFirestore {
		ctx := context.Background()
		var opts []option.ClientOption
		if credsFile != "" {
			opts = append(opts, option.WithCredentialsFile(credsFile))
		}

		client, err := firestore.NewClient(ctx, projectID, opts...)
		if err != nil {
			log.Printf("[FIRESTORE] Failed to create client: %v. Falling back to in-memory mode.", err)
			repo.isMemory = true
		} else {
			repo.client = client
			log.Printf("[FIRESTORE] Connected to Google Cloud Firestore project: %s", projectID)
		}
	}

	// Seeding setup
	if repo.isMemory {
		repo.seedData()
	} else {
		// Auto seed Firestore if it's empty
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		iter := repo.client.Collection("users").Limit(1).Documents(ctx)
		_, err := iter.Next()
		if err == iterator.Done {
			log.Println("[FIRESTORE] Users collection is empty. Seeding initial admin and staff data...")
			repo.seedData()
		} else if err != nil {
			log.Printf("[FIRESTORE] Could not check users collection status: %v. Seeding skipped.", err)
		}
	}

	return repo
}

func (r *FirestoreRepository) seedEntity(collection string, id string, entity interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.client.Collection(collection).Doc(id).Set(ctx, entity)
	if err != nil {
		log.Printf("[FIRESTORE] Error seeding %s/%s: %v", collection, id, err)
	}
}

func isNotFound(err error) bool {
	return status.Code(err) == codes.NotFound
}

func (r *FirestoreRepository) seedData() {
	// Seed Admin and Staff Users
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	staffHash, _ := bcrypt.GenerateFromPassword([]byte("staff123"), bcrypt.DefaultCost)

	// 1. Users
	users := []*models.User{
		{ID: "admin", Username: "admin", Password: string(adminHash), Role: "admin", Department: "Management", CreatedAt: time.Now()},
		{ID: "volunteer1", Username: "volunteer1", Password: string(staffHash), Role: "volunteer", Department: "Visitor Services", CreatedAt: time.Now()},
		{ID: "security1", Username: "security1", Password: string(staffHash), Role: "security", Department: "Safety", CreatedAt: time.Now()},
		{ID: "medical1", Username: "medical1", Password: string(staffHash), Role: "medical", Department: "First Aid", CreatedAt: time.Now()},
		{ID: "cleaning1", Username: "cleaning1", Password: string(staffHash), Role: "cleaning", Department: "Facilities", CreatedAt: time.Now()},
		{ID: "ops1", Username: "ops1", Password: string(staffHash), Role: "ops", Department: "Coordination", CreatedAt: time.Now()},
	}
	for _, u := range users {
		if r.isMemory {
			r.mem.users[u.ID] = u
		} else {
			r.seedEntity("users", u.ID, u)
		}
	}

	// 2. Stadiums
	stadiums := []*models.Stadium{
		{
			ID:               "stadium_01",
			Name:             "Orynt Arena",
			Capacity:         65000,
			AccessibleRoutes: []string{"Elevator 1 (North Stand)", "Ramp B (South Gate)", "Priority Exit 4"},
			Facilities: map[string][]models.Facility{
				"food": {
					{ID: "f1", Name: "Corner Grill", Location: "Section 102", Type: "food"},
					{ID: "f2", Name: "Slam Dunk Brews", Location: "Section 115", Type: "food"},
					{ID: "f3", Name: "Vegan Bites", Location: "Section 204", Type: "food"},
				},
				"restrooms": {
					{ID: "r1", Name: "Washrooms (Male/Female/Accessible)", Location: "Section 108", Type: "restroom"},
					{ID: "r2", Name: "Washrooms (Male/Female/Accessible)", Location: "Section 218", Type: "restroom"},
				},
				"medical": {
					{ID: "m1", Name: "First Aid Station A", Location: "Gate 2 Floor 1", Type: "medical"},
					{ID: "m2", Name: "Emergency Response Center", Location: "Gate 5 Ground Floor", Type: "medical"},
				},
				"exits": {
					{ID: "e1", Name: "Exit Gate A", Location: "North Side", Type: "exit"},
					{ID: "e2", Name: "Exit Gate B (Accessible)", Location: "South-West Side", Type: "exit"},
				},
			},
		},
	}
	for _, s := range stadiums {
		if r.isMemory {
			r.mem.stadiums[s.ID] = s
		} else {
			r.seedEntity("stadiums", s.ID, s)
		}
	}

	// 3. Tournaments
	tournaments := []*models.Tournament{
		{ID: "t1", Name: "National Championship Cup 2026", Status: "active", StartDate: time.Now().Add(-24 * time.Hour), EndDate: time.Now().Add(5 * 24 * time.Hour)},
		{ID: "t2", Name: "Pro League Season Finale", Status: "planned", StartDate: time.Now().Add(10 * 24 * time.Hour), EndDate: time.Now().Add(15 * 24 * time.Hour)},
	}
	for _, t := range tournaments {
		if r.isMemory {
			r.mem.tournaments[t.ID] = t
		} else {
			r.seedEntity("tournaments", t.ID, t)
		}
	}

	// 4. Matches
	matches := []*models.Match{
		{
			ID:           "m_01",
			TournamentID: "t1",
			HomeTeam:     "Spartans FC",
			AwayTeam:     "Titans United",
			HomeScore:    2,
			AwayScore:    1,
			Status:       "live",
			ScheduledAt:  time.Now().Add(-1 * time.Hour),
			TimeElapsed:  "67:23",
			Events: []models.MatchComment{
				{Time: "12'", Type: "info", Detail: "Match kicks off under clear conditions at Orynt Arena."},
				{Time: "24'", Type: "goal", Detail: "GOAL! Spartans FC score. Target hit by striker #9 (J. Cole)."},
				{Time: "45'", Type: "card", Detail: "Yellow Card: Titans United #4 (R. Vance) for a sliding tackle."},
				{Time: "55'", Type: "goal", Detail: "GOAL! Titans United equalizes. Header by #11 (S. Ray) off a corner kick."},
				{Time: "62'", Type: "goal", Detail: "GOAL! Spartans FC retake the lead. Volley by #7 (L. Messi) from 20 yards out."},
			},
		},
		{
			ID:           "m_02",
			TournamentID: "t1",
			HomeTeam:     "Strikers SC",
			AwayTeam:     "Wanderers FC",
			HomeScore:    0,
			AwayScore:    0,
			Status:       "scheduled",
			ScheduledAt:  time.Now().Add(3 * time.Hour),
			TimeElapsed:  "00:00",
			Events:       []models.MatchComment{},
		},
	}
	for _, m := range matches {
		if r.isMemory {
			r.mem.matches[m.ID] = m
		} else {
			r.seedEntity("matches", m.ID, m)
		}
	}

	// 5. Crowd Zones
	crowdZones := []*models.CrowdZone{
		{ID: "z1", StadiumID: "stadium_01", ZoneName: "North Stand (Section 101-110)", CurrentOccupancy: 8400, MaxCapacity: 10000, DensityLevel: "high", LastUpdated: time.Now()},
		{ID: "z2", StadiumID: "stadium_01", ZoneName: "East Stand (Section 111-120)", CurrentOccupancy: 4200, MaxCapacity: 12000, DensityLevel: "low", LastUpdated: time.Now()},
		{ID: "z3", StadiumID: "stadium_01", ZoneName: "South Stand (Section 201-210)", CurrentOccupancy: 7800, MaxCapacity: 10000, DensityLevel: "medium", LastUpdated: time.Now()},
		{ID: "z4", StadiumID: "stadium_01", ZoneName: "West Stand (Section 211-220)", CurrentOccupancy: 2200, MaxCapacity: 12000, DensityLevel: "low", LastUpdated: time.Now()},
		{ID: "z5", StadiumID: "stadium_01", ZoneName: "VIP Concourse", CurrentOccupancy: 4800, MaxCapacity: 5000, DensityLevel: "high", LastUpdated: time.Now()},
	}
	for _, cz := range crowdZones {
		if r.isMemory {
			r.mem.crowdZones[cz.ID] = cz
		} else {
			r.seedEntity("crowd_zones", cz.ID, cz)
		}
	}

	// 6. Parking
	parking := []*models.Parking{
		{ID: "p1", ZoneName: "Parking Lot A (North)", TotalSpots: 500, OccupiedSpots: 482, Type: "public", Location: "Adjacent Gate A"},
		{ID: "p2", ZoneName: "Parking Lot B (South)", TotalSpots: 800, OccupiedSpots: 340, Type: "public", Location: "Adjacent Gate B"},
		{ID: "p3", ZoneName: "VIP Parking Deck", TotalSpots: 150, OccupiedSpots: 145, Type: "vip", Location: "Main Entrance Tower"},
		{ID: "p4", ZoneName: "Accessible Lot A", TotalSpots: 80, OccupiedSpots: 22, Type: "accessible", Location: "Closest to ramp Elevator 1"},
	}
	for _, p := range parking {
		if r.isMemory {
			r.mem.parkingZones[p.ID] = p
		} else {
			r.seedEntity("parking", p.ID, p)
		}
	}

	// 7. Transport
	transport := []*models.Transport{
		{ID: "t_line_1", RouteName: "City Center Metro Shuttle", Mode: "shuttle", NextArrival: time.Now().Add(4 * time.Minute), DelayMinutes: 0, Status: "on_time", Capacity: "high"},
		{ID: "t_line_2", RouteName: "East Suburbs Express (Bus 304)", Mode: "bus", NextArrival: time.Now().Add(12 * time.Minute), DelayMinutes: 5, Status: "delayed", Capacity: "normal"},
		{ID: "t_line_3", RouteName: "North Stadium Link (Train)", Mode: "train", NextArrival: time.Now().Add(8 * time.Minute), DelayMinutes: 0, Status: "on_time", Capacity: "full"},
	}
	for _, t := range transport {
		if r.isMemory {
			r.mem.transport[t.ID] = t
		} else {
			r.seedEntity("transport", t.ID, t)
		}
	}

	// 8. Food Stalls
	foodStalls := []*models.FoodStall{
		{ID: "stall_01", Name: "Starlight Hotdogs & Burgers", Location: "Section 104 Concourse", FoodType: "Fast Food", Status: "open", WaitTimeMinutes: 15, Menu: []string{"Double Cheeseburger", "Spicy Chili Dog", "Crispy Fries", "Soda"}},
		{ID: "stall_02", Name: "Orynt Brewery & Bar", Location: "Section 118 Pavilion", FoodType: "Beverages", Status: "open", WaitTimeMinutes: 5, Menu: []string{"Craft IPA", "Draft Lager", "Pretzels with Dip"}},
		{ID: "stall_03", Name: "Noodle House & Bowls", Location: "Section 206 Level 2", FoodType: "Asian Express", Status: "open", WaitTimeMinutes: 25, Menu: []string{"Teriyaki Chicken Bowl", "Vegetable Spring Rolls", "Iced Green Tea"}},
	}
	for _, fs := range foodStalls {
		if r.isMemory {
			r.mem.foodStalls[fs.ID] = fs
		} else {
			r.seedEntity("food_stalls", fs.ID, fs)
		}
	}

	// 9. Tasks
	tasks := []*models.Task{
		{ID: "task_01", Title: "Inspect North Stand Turnstile Gate 3", Description: "Reports of slow card scanning. Investigate hardware or network issue.", AssignedTo: "volunteer1", Status: "todo", Priority: "medium", Department: "Visitor Services", CreatedAt: time.Now().Add(-2 * time.Hour)},
		{ID: "task_02", Title: "Clear Spillage near Section 106 Restroom", Description: "Large soda spill. Needs wet floor caution sign and mopping.", AssignedTo: "cleaning1", Status: "in_progress", Priority: "high", Department: "Facilities", CreatedAt: time.Now().Add(-1 * time.Hour)},
		{ID: "task_03", Title: "Monitor VIP Parking Deck Access Control", Description: "Ensure only VIP pass holders are admitted. High congestion expected.", AssignedTo: "volunteer1", Status: "done", Priority: "low", Department: "Visitor Services", CreatedAt: time.Now().Add(-4 * time.Hour)},
	}
	for _, tk := range tasks {
		if r.isMemory {
			r.mem.tasks[tk.ID] = tk
		} else {
			r.seedEntity("tasks", tk.ID, tk)
		}
	}

	// 10. Alerts
	alerts := []*models.Alert{
		{ID: "alert_01", Title: "Extreme Crowd Build-up at Gate A", Content: "High congestion detected at North Gate A. Visitors are advised to proceed to Gate B or C for faster entry.", Type: "crowd", Severity: "high", CreatedAt: time.Now().Add(-15 * time.Minute), Active: true},
		{ID: "alert_02", Title: "First-Aid Dispatch: Section 104", Content: "Medical team is responding to a heat exhaustion report in Section 104, Row 12.", Type: "emergency", Severity: "critical", CreatedAt: time.Now().Add(-5 * time.Minute), Active: true},
	}
	for _, a := range alerts {
		if r.isMemory {
			r.mem.alerts[a.ID] = a
		} else {
			r.seedEntity("alerts", a.ID, a)
		}
	}

	// 11. Announcements
	announcements := []*models.Announcement{
		{ID: "ann_01", Title: "Free Shuttle Service Available After Match", Content: "Complimentary shuttle services back to City Center Station will run every 5 minutes from Gate B for 2 hours post-match.", TargetAudience: "public", Approved: true, ApprovedBy: "admin", CreatedAt: time.Now().Add(-30 * time.Minute)},
		{ID: "ann_02", Title: "Staff Briefing: Shift B Handover", Content: "All volunteers on Shift B, please report to Room 12-A for briefing at 22:30.", TargetAudience: "staff", Approved: true, ApprovedBy: "ops1", CreatedAt: time.Now().Add(-10 * time.Minute)},
	}
	for _, an := range announcements {
		if r.isMemory {
			r.mem.announcements[an.ID] = an
		} else {
			r.seedEntity("announcements", an.ID, an)
		}
	}
}

// Users implementations
func (r *FirestoreRepository) GetUser(ctx context.Context, id string) (*models.User, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		user, exists := r.mem.users[id]
		if !exists {
			return nil, errors.New("user not found")
		}
		return user, nil
	}

	doc, err := r.client.Collection("users").Doc(id).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *FirestoreRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		for _, user := range r.mem.users {
			if user.Username == username {
				return user, nil
			}
		}
		return nil, errors.New("user not found")
	}

	iter := r.client.Collection("users").Where("username", "==", username).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *FirestoreRepository) CreateUser(ctx context.Context, user *models.User) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		if _, exists := r.mem.users[user.ID]; exists {
			return errors.New("user already exists")
		}
		r.mem.users[user.ID] = user
		return nil
	}

	_, err := r.GetUser(ctx, user.ID)
	if err == nil {
		return errors.New("user already exists")
	}

	_, err = r.client.Collection("users").Doc(user.ID).Set(ctx, user)
	return err
}

func (r *FirestoreRepository) ListUsers(ctx context.Context) ([]models.User, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.User, 0, len(r.mem.users))
		for _, u := range r.mem.users {
			list = append(list, *u)
		}
		return list, nil
	}

	docs, err := r.client.Collection("users").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.User, 0, len(docs))
	for _, doc := range docs {
		var u models.User
		if err := doc.DataTo(&u); err == nil {
			list = append(list, u)
		}
	}
	return list, nil
}

// Tournaments & Matches
func (r *FirestoreRepository) GetTournament(ctx context.Context, id string) (*models.Tournament, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		t, exists := r.mem.tournaments[id]
		if !exists {
			return nil, errors.New("tournament not found")
		}
		return t, nil
	}

	doc, err := r.client.Collection("tournaments").Doc(id).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil, errors.New("tournament not found")
		}
		return nil, err
	}
	var t models.Tournament
	if err := doc.DataTo(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *FirestoreRepository) ListTournaments(ctx context.Context) ([]models.Tournament, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.Tournament, 0, len(r.mem.tournaments))
		for _, t := range r.mem.tournaments {
			list = append(list, *t)
		}
		return list, nil
	}

	docs, err := r.client.Collection("tournaments").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.Tournament, 0, len(docs))
	for _, doc := range docs {
		var t models.Tournament
		if err := doc.DataTo(&t); err == nil {
			list = append(list, t)
		}
	}
	return list, nil
}

func (r *FirestoreRepository) CreateTournament(ctx context.Context, tournament *models.Tournament) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.tournaments[tournament.ID] = tournament
		return nil
	}

	_, err := r.client.Collection("tournaments").Doc(tournament.ID).Set(ctx, tournament)
	return err
}

func (r *FirestoreRepository) GetMatch(ctx context.Context, id string) (*models.Match, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		m, exists := r.mem.matches[id]
		if !exists {
			return nil, errors.New("match not found")
		}
		return m, nil
	}

	doc, err := r.client.Collection("matches").Doc(id).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil, errors.New("match not found")
		}
		return nil, err
	}
	var m models.Match
	if err := doc.DataTo(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *FirestoreRepository) ListMatches(ctx context.Context, tournamentID string) ([]models.Match, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.Match, 0)
		for _, m := range r.mem.matches {
			if tournamentID == "" || m.TournamentID == tournamentID {
				list = append(list, *m)
			}
		}
		return list, nil
	}

	var q firestore.Query
	if tournamentID != "" {
		q = r.client.Collection("matches").Where("tournamentId", "==", tournamentID)
	} else {
		q = r.client.Collection("matches").Query
	}

	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.Match, 0, len(docs))
	for _, doc := range docs {
		var m models.Match
		if err := doc.DataTo(&m); err == nil {
			list = append(list, m)
		}
	}
	return list, nil
}

func (r *FirestoreRepository) CreateMatch(ctx context.Context, match *models.Match) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.matches[match.ID] = match
		return nil
	}

	_, err := r.client.Collection("matches").Doc(match.ID).Set(ctx, match)
	return err
}

func (r *FirestoreRepository) UpdateMatch(ctx context.Context, match *models.Match) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		if _, exists := r.mem.matches[match.ID]; !exists {
			return errors.New("match not found")
		}
		r.mem.matches[match.ID] = match
		return nil
	}

	_, err := r.client.Collection("matches").Doc(match.ID).Set(ctx, match)
	return err
}

// Stadiums & Zones
func (r *FirestoreRepository) GetStadium(ctx context.Context, id string) (*models.Stadium, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		s, exists := r.mem.stadiums[id]
		if !exists {
			return nil, errors.New("stadium not found")
		}
		return s, nil
	}

	doc, err := r.client.Collection("stadiums").Doc(id).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil, errors.New("stadium not found")
		}
		return nil, err
	}
	var s models.Stadium
	if err := doc.DataTo(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *FirestoreRepository) CreateStadium(ctx context.Context, stadium *models.Stadium) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.stadiums[stadium.ID] = stadium
		return nil
	}

	_, err := r.client.Collection("stadiums").Doc(stadium.ID).Set(ctx, stadium)
	return err
}

func (r *FirestoreRepository) ListCrowdZones(ctx context.Context, stadiumID string) ([]models.CrowdZone, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.CrowdZone, 0)
		for _, z := range r.mem.crowdZones {
			if stadiumID == "" || z.StadiumID == stadiumID {
				list = append(list, *z)
			}
		}
		return list, nil
	}

	var q firestore.Query
	if stadiumID != "" {
		q = r.client.Collection("crowd_zones").Where("stadiumId", "==", stadiumID)
	} else {
		q = r.client.Collection("crowd_zones").Query
	}

	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.CrowdZone, 0, len(docs))
	for _, doc := range docs {
		var z models.CrowdZone
		if err := doc.DataTo(&z); err == nil {
			list = append(list, z)
		}
	}
	return list, nil
}

func (r *FirestoreRepository) UpdateCrowdZoneOccupancy(ctx context.Context, zoneID string, occupancy int) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		z, exists := r.mem.crowdZones[zoneID]
		if !exists {
			return errors.New("zone not found")
		}
		z.CurrentOccupancy = occupancy
		ratio := float64(occupancy) / float64(z.MaxCapacity)
		if ratio >= 0.8 {
			z.DensityLevel = "high"
		} else if ratio >= 0.4 {
			z.DensityLevel = "medium"
		} else {
			z.DensityLevel = "low"
		}
		z.LastUpdated = time.Now()
		return nil
	}

	doc, err := r.client.Collection("crowd_zones").Doc(zoneID).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return errors.New("zone not found")
		}
		return err
	}
	var z models.CrowdZone
	if err := doc.DataTo(&z); err != nil {
		return err
	}

	z.CurrentOccupancy = occupancy
	ratio := float64(occupancy) / float64(z.MaxCapacity)
	if ratio >= 0.8 {
		z.DensityLevel = "high"
	} else if ratio >= 0.4 {
		z.DensityLevel = "medium"
	} else {
		z.DensityLevel = "low"
	}
	z.LastUpdated = time.Now()

	_, err = r.client.Collection("crowd_zones").Doc(zoneID).Set(ctx, &z)
	return err
}

func (r *FirestoreRepository) CreateCrowdZone(ctx context.Context, zone *models.CrowdZone) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.crowdZones[zone.ID] = zone
		return nil
	}

	_, err := r.client.Collection("crowd_zones").Doc(zone.ID).Set(ctx, zone)
	return err
}

// Transport & Parking
func (r *FirestoreRepository) ListParkingZones(ctx context.Context) ([]models.Parking, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.Parking, 0, len(r.mem.parkingZones))
		for _, p := range r.mem.parkingZones {
			list = append(list, *p)
		}
		return list, nil
	}

	docs, err := r.client.Collection("parking").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.Parking, 0, len(docs))
	for _, doc := range docs {
		var p models.Parking
		if err := doc.DataTo(&p); err == nil {
			list = append(list, p)
		}
	}
	return list, nil
}

func (r *FirestoreRepository) UpdateParkingSpots(ctx context.Context, lotID string, occupied int) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		p, exists := r.mem.parkingZones[lotID]
		if !exists {
			return errors.New("parking zone not found")
		}
		if occupied > p.TotalSpots {
			occupied = p.TotalSpots
		}
		p.OccupiedSpots = occupied
		return nil
	}

	doc, err := r.client.Collection("parking").Doc(lotID).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return errors.New("parking zone not found")
		}
		return err
	}
	var p models.Parking
	if err := doc.DataTo(&p); err != nil {
		return err
	}

	if occupied > p.TotalSpots {
		occupied = p.TotalSpots
	}
	p.OccupiedSpots = occupied

	_, err = r.client.Collection("parking").Doc(lotID).Set(ctx, &p)
	return err
}

func (r *FirestoreRepository) CreateParkingZone(ctx context.Context, lot *models.Parking) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.parkingZones[lot.ID] = lot
		return nil
	}

	_, err := r.client.Collection("parking").Doc(lot.ID).Set(ctx, lot)
	return err
}

func (r *FirestoreRepository) ListTransportLines(ctx context.Context) ([]models.Transport, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.Transport, 0, len(r.mem.transport))
		for _, t := range r.mem.transport {
			list = append(list, *t)
		}
		return list, nil
	}

	docs, err := r.client.Collection("transport").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.Transport, 0, len(docs))
	for _, doc := range docs {
		var t models.Transport
		if err := doc.DataTo(&t); err == nil {
			list = append(list, t)
		}
	}
	return list, nil
}

func (r *FirestoreRepository) UpdateTransportArrival(ctx context.Context, routeID string, delay int, status string) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		t, exists := r.mem.transport[routeID]
		if !exists {
			return errors.New("transport route not found")
		}
		t.DelayMinutes = delay
		t.Status = status
		t.NextArrival = time.Now().Add(time.Duration(10+delay) * time.Minute)
		return nil
	}

	doc, err := r.client.Collection("transport").Doc(routeID).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return errors.New("transport route not found")
		}
		return err
	}
	var t models.Transport
	if err := doc.DataTo(&t); err != nil {
		return err
	}

	t.DelayMinutes = delay
	t.Status = status
	t.NextArrival = time.Now().Add(time.Duration(10+delay) * time.Minute)

	_, err = r.client.Collection("transport").Doc(routeID).Set(ctx, &t)
	return err
}

func (r *FirestoreRepository) CreateTransportLine(ctx context.Context, route *models.Transport) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.transport[route.ID] = route
		return nil
	}

	_, err := r.client.Collection("transport").Doc(route.ID).Set(ctx, route)
	return err
}

// Alerts & Announcements
func (r *FirestoreRepository) ListAlerts(ctx context.Context, activeOnly bool) ([]models.Alert, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.Alert, 0)
		for _, a := range r.mem.alerts {
			if !activeOnly || a.Active {
				list = append(list, *a)
			}
		}
		return list, nil
	}

	var q firestore.Query
	if activeOnly {
		q = r.client.Collection("alerts").Where("active", "==", true)
	} else {
		q = r.client.Collection("alerts").Query
	}

	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.Alert, 0, len(docs))
	for _, doc := range docs {
		var a models.Alert
		if err := doc.DataTo(&a); err == nil {
			list = append(list, a)
		}
	}
	return list, nil
}

func (r *FirestoreRepository) CreateAlert(ctx context.Context, alert *models.Alert) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.alerts[alert.ID] = alert
		return nil
	}

	_, err := r.client.Collection("alerts").Doc(alert.ID).Set(ctx, alert)
	return err
}

func (r *FirestoreRepository) UpdateAlertStatus(ctx context.Context, id string, active bool) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		a, exists := r.mem.alerts[id]
		if !exists {
			return errors.New("alert not found")
		}
		a.Active = active
		return nil
	}

	doc, err := r.client.Collection("alerts").Doc(id).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return errors.New("alert not found")
		}
		return err
	}
	var a models.Alert
	if err := doc.DataTo(&a); err != nil {
		return err
	}

	a.Active = active

	_, err = r.client.Collection("alerts").Doc(id).Set(ctx, &a)
	return err
}

func (r *FirestoreRepository) ListAnnouncements(ctx context.Context, role string) ([]models.Announcement, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.Announcement, 0)
		for _, a := range r.mem.announcements {
			if !a.Approved {
				continue
			}
			if role == "public" && a.TargetAudience == "staff" {
				continue
			}
			list = append(list, *a)
		}
		return list, nil
	}

	docs, err := r.client.Collection("announcements").Where("approved", "==", true).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.Announcement, 0, len(docs))
	for _, doc := range docs {
		var a models.Announcement
		if err := doc.DataTo(&a); err == nil {
			if role == "public" && a.TargetAudience == "staff" {
				continue
			}
			list = append(list, a)
		}
	}
	return list, nil
}

func (r *FirestoreRepository) CreateAnnouncement(ctx context.Context, announcement *models.Announcement) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.announcements[announcement.ID] = announcement
		return nil
	}

	_, err := r.client.Collection("announcements").Doc(announcement.ID).Set(ctx, announcement)
	return err
}

func (r *FirestoreRepository) ApproveAnnouncement(ctx context.Context, id string, adminUsername string) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		a, exists := r.mem.announcements[id]
		if !exists {
			return errors.New("announcement not found")
		}
		a.Approved = true
		a.ApprovedBy = adminUsername
		return nil
	}

	doc, err := r.client.Collection("announcements").Doc(id).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return errors.New("announcement not found")
		}
		return err
	}
	var a models.Announcement
	if err := doc.DataTo(&a); err != nil {
		return err
	}

	a.Approved = true
	a.ApprovedBy = adminUsername

	_, err = r.client.Collection("announcements").Doc(id).Set(ctx, &a)
	return err
}

// Tasks & Operations
func (r *FirestoreRepository) ListTasks(ctx context.Context, department string) ([]models.Task, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.Task, 0)
		for _, t := range r.mem.tasks {
			if department == "" || t.Department == department {
				list = append(list, *t)
			}
		}
		return list, nil
	}

	var q firestore.Query
	if department != "" {
		q = r.client.Collection("tasks").Where("department", "==", department)
	} else {
		q = r.client.Collection("tasks").Query
	}

	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.Task, 0, len(docs))
	for _, doc := range docs {
		var t models.Task
		if err := doc.DataTo(&t); err == nil {
			list = append(list, t)
		}
	}
	return list, nil
}

func (r *FirestoreRepository) CreateTask(ctx context.Context, task *models.Task) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.tasks[task.ID] = task
		return nil
	}

	_, err := r.client.Collection("tasks").Doc(task.ID).Set(ctx, task)
	return err
}

func (r *FirestoreRepository) UpdateTaskStatus(ctx context.Context, id string, status string) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		t, exists := r.mem.tasks[id]
		if !exists {
			return errors.New("task not found")
		}
		t.Status = status
		return nil
	}

	doc, err := r.client.Collection("tasks").Doc(id).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return errors.New("task not found")
		}
		return err
	}
	var t models.Task
	if err := doc.DataTo(&t); err != nil {
		return err
	}

	t.Status = status

	_, err = r.client.Collection("tasks").Doc(id).Set(ctx, &t)
	return err
}

func (r *FirestoreRepository) ListLostFound(ctx context.Context) ([]models.LostFoundItem, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.LostFoundItem, 0, len(r.mem.lostFound))
		for _, lf := range r.mem.lostFound {
			list = append(list, *lf)
		}
		return list, nil
	}

	docs, err := r.client.Collection("lost_found").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.LostFoundItem, 0, len(docs))
	for _, doc := range docs {
		var lf models.LostFoundItem
		if err := doc.DataTo(&lf); err == nil {
			list = append(list, lf)
		}
	}
	return list, nil
}

func (r *FirestoreRepository) CreateLostFound(ctx context.Context, item *models.LostFoundItem) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.lostFound[item.ID] = item
		return nil
	}

	_, err := r.client.Collection("lost_found").Doc(item.ID).Set(ctx, item)
	return err
}

func (r *FirestoreRepository) UpdateLostFoundStatus(ctx context.Context, id string, status string) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		lf, exists := r.mem.lostFound[id]
		if !exists {
			return errors.New("lost & found item not found")
		}
		lf.Status = status
		return nil
	}

	doc, err := r.client.Collection("lost_found").Doc(id).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return errors.New("lost & found item not found")
		}
		return err
	}
	var lf models.LostFoundItem
	if err := doc.DataTo(&lf); err != nil {
		return err
	}

	lf.Status = status

	_, err = r.client.Collection("lost_found").Doc(id).Set(ctx, &lf)
	return err
}

func (r *FirestoreRepository) ListFoodStalls(ctx context.Context) ([]models.FoodStall, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.FoodStall, 0, len(r.mem.foodStalls))
		for _, f := range r.mem.foodStalls {
			list = append(list, *f)
		}
		return list, nil
	}

	docs, err := r.client.Collection("food_stalls").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.FoodStall, 0, len(docs))
	for _, doc := range docs {
		var f models.FoodStall
		if err := doc.DataTo(&f); err == nil {
			list = append(list, f)
		}
	}
	return list, nil
}

func (r *FirestoreRepository) CreateFoodStall(ctx context.Context, stall *models.FoodStall) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.foodStalls[stall.ID] = stall
		return nil
	}

	_, err := r.client.Collection("food_stalls").Doc(stall.ID).Set(ctx, stall)
	return err
}

func (r *FirestoreRepository) UpdateFoodStallWaitTime(ctx context.Context, id string, minutes int) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		f, exists := r.mem.foodStalls[id]
		if !exists {
			return errors.New("food stall not found")
		}
		f.WaitTimeMinutes = minutes
		return nil
	}

	doc, err := r.client.Collection("food_stalls").Doc(id).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return errors.New("food stall not found")
		}
		return err
	}
	var f models.FoodStall
	if err := doc.DataTo(&f); err != nil {
		return err
	}

	f.WaitTimeMinutes = minutes

	_, err = r.client.Collection("food_stalls").Doc(id).Set(ctx, &f)
	return err
}

// Medical
func (r *FirestoreRepository) ListMedicalRequests(ctx context.Context) ([]models.MedicalRequest, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.MedicalRequest, 0, len(r.mem.medical))
		for _, m := range r.mem.medical {
			list = append(list, *m)
		}
		return list, nil
	}

	docs, err := r.client.Collection("medical").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.MedicalRequest, 0, len(docs))
	for _, doc := range docs {
		var m models.MedicalRequest
		if err := doc.DataTo(&m); err == nil {
			list = append(list, m)
		}
	}
	return list, nil
}

func (r *FirestoreRepository) CreateMedicalRequest(ctx context.Context, req *models.MedicalRequest) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.medical[req.ID] = req
		return nil
	}

	_, err := r.client.Collection("medical").Doc(req.ID).Set(ctx, req)
	return err
}

func (r *FirestoreRepository) UpdateMedicalRequestStatus(ctx context.Context, id string, status string, assignedTo string) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		req, exists := r.mem.medical[id]
		if !exists {
			return errors.New("medical request not found")
		}
		req.Status = status
		if assignedTo != "" {
			req.AssignedTo = assignedTo
		}
		return nil
	}

	doc, err := r.client.Collection("medical").Doc(id).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return errors.New("medical request not found")
		}
		return err
	}
	var req models.MedicalRequest
	if err := doc.DataTo(&req); err != nil {
		return err
	}

	req.Status = status
	if assignedTo != "" {
		req.AssignedTo = assignedTo
	}

	_, err = r.client.Collection("medical").Doc(id).Set(ctx, &req)
	return err
}

// Audit Logs & Settings
func (r *FirestoreRepository) CreateAuditLog(ctx context.Context, logVal *models.AuditLog) error {
	if r.isMemory {
		r.mem.mu.Lock()
		defer r.mem.mu.Unlock()
		r.mem.auditLogs = append(r.mem.auditLogs, *logVal)
		return nil
	}

	_, err := r.client.Collection("audit_logs").Doc(logVal.ID).Set(ctx, logVal)
	return err
}

func (r *FirestoreRepository) ListAuditLogs(ctx context.Context) ([]models.AuditLog, error) {
	if r.isMemory {
		r.mem.mu.RLock()
		defer r.mem.mu.RUnlock()
		list := make([]models.AuditLog, len(r.mem.auditLogs))
		copy(list, r.mem.auditLogs)
		return list, nil
	}

	docs, err := r.client.Collection("audit_logs").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	list := make([]models.AuditLog, 0, len(docs))
	for _, doc := range docs {
		var al models.AuditLog
		if err := doc.DataTo(&al); err == nil {
			list = append(list, al)
		}
	}
	return list, nil
}
