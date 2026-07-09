export interface User {
  id: string;
  username: string;
  role: 'admin' | 'volunteer' | 'security' | 'medical' | 'cleaning' | 'parking' | 'transport' | 'ops';
  department: string;
  createdAt: string;
}

export interface TokenResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
}

export interface Facility {
  id: string;
  name: string;
  location: string;
  type: 'food' | 'restroom' | 'medical' | 'exit';
}

export interface Stadium {
  id: string;
  name: string;
  capacity: number;
  accessibleRoutes: string[];
  facilities: {
    food: Facility[];
    restrooms: Facility[];
    medical: Facility[];
    exits: Facility[];
  };
}

export interface Tournament {
  id: string;
  name: string;
  status: 'planned' | 'active' | 'completed';
  startDate: string;
  endDate: string;
}

export interface MatchComment {
  time: string;
  type: 'goal' | 'card' | 'substitution' | 'info';
  detail: string;
}

export interface Match {
  id: string;
  tournamentId: string;
  homeTeam: string;
  awayTeam: string;
  homeScore: number;
  awayScore: number;
  status: 'scheduled' | 'live' | 'completed';
  scheduledAt: string;
  timeElapsed: string;
  events: MatchComment[];
}

export interface CrowdZone {
  id: string;
  stadiumId: string;
  zoneName: string;
  currentOccupancy: number;
  maxCapacity: number;
  densityLevel: 'low' | 'medium' | 'high';
  lastUpdated: string;
}

export interface Parking {
  id: string;
  zoneName: string;
  totalSpots: number;
  occupiedSpots: number;
  type: 'public' | 'vip' | 'staff' | 'accessible';
  location: string;
}

export interface Transport {
  id: string;
  routeName: string;
  mode: 'bus' | 'train' | 'shuttle';
  nextArrival: string;
  delayMinutes: number;
  status: 'on_time' | 'delayed' | 'suspended';
  capacity: 'normal' | 'high' | 'full';
}

export interface Alert {
  id: string;
  title: string;
  content: string;
  type: 'emergency' | 'weather' | 'crowd' | 'transport';
  severity: 'critical' | 'high' | 'info';
  createdAt: string;
  active: boolean;
}

export interface Announcement {
  id: string;
  title: string;
  content: string;
  targetAudience: 'public' | 'staff' | 'all';
  approved: boolean;
  approvedBy: string;
  createdAt: string;
}

export interface FoodStall {
  id: string;
  name: string;
  location: string;
  foodType: string;
  status: 'open' | 'closed';
  waitTimeMinutes: number;
  menu: string[];
}

export interface MedicalRequest {
  id: string;
  requester: string;
  location: string;
  description: string;
  status: 'pending' | 'assigned' | 'resolved';
  assignedTo: string;
  requestedAt: string;
}

export interface Task {
  id: string;
  title: string;
  description: string;
  assignedTo: string;
  status: 'todo' | 'in_progress' | 'done';
  priority: 'low' | 'medium' | 'high';
  department: string;
  createdAt: string;
}

export interface LostFoundItem {
  id: string;
  itemName: string;
  description: string;
  category: 'electronics' | 'clothing' | 'keys' | 'bags' | 'other';
  status: 'reported_lost' | 'found' | 'claimed';
  contactName: string;
  contactPhone: string;
  reportedAt: string;
}

export interface AuditLog {
  id: string;
  userId: string;
  username: string;
  action: string;
  resource: string;
  details: string;
  timestamp: string;
}

export interface LiveTelemetryMetrics {
  attendance: number;
  peakCrowd: number;
  parkingOccupied: number;
  parkingTotal: number;
  queueLengthAverage: number;
  shuttleDelayAverage: number;
  wasteGeneratedKg: number;
  energyUsageKw: number;
  waterConsumptionL: number;
  completedTasks: number;
  totalTasks: number;
  unclaimedLostItems: number;
  medicalPendingAlerts: number;
  timestamp: string;
}

export interface AIChatResponse {
  response: string;
  sources: string[];
}

export interface SustainabilityHistoricalPoint {
  time: string;
  attendance: number;
  energyUsageKw: number;
  waterConsumptionL: number;
  wasteGeneratedKg: number;
}

export interface SustainabilityMetrics {
  current: {
    attendance: number;
    energyUsageKw: number;
    waterConsumptionL: number;
    wasteGeneratedKg: number;
  };
  historical: SustainabilityHistoricalPoint[];
  recommendations: string;
}

export interface Point {
  x: number;
  y: number;
}

export interface ExitRoute {
  zoneId: string;
  zoneName: string;
  recommendedExit: string;
  exitName: string;
  congestionLevel: 'low' | 'medium' | 'high';
  pathPoints: Point[];
  reason: string;
}

export interface AIChatHistoryItem {
  sender: 'user' | 'ai';
  text: string;
}


