import React, { useState, useEffect } from 'react'
import { LiveTelemetryMetrics, Match, CrowdZone, Parking, Transport, AuditLog } from '../../types'
import { api } from '../../services/api'
import { Card, CustomSelect } from '../../components/ui'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area } from 'recharts'
import { Users, Sparkles, Plus } from 'lucide-react'

export const AdminDashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<LiveTelemetryMetrics | null>(null)
  const [matches, setMatches] = useState<Match[]>([])
  const [zones, setZones] = useState<CrowdZone[]>([])
  const [parking, setParking] = useState<Parking[]>([])
  const [transport, setTransport] = useState<Transport[]>([])
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([])

  // Mock timeline data for AreaChart visualization
  const graphTimelineData = [
    { time: '18:00', visitors: 12000, congestionIndex: 15 },
    { time: '19:00', visitors: 28000, congestionIndex: 32 },
    { time: '20:00', visitors: 49000, congestionIndex: 68 },
    { time: '20:30', visitors: 58000, congestionIndex: 82 },
    { time: '21:00', visitors: 61000, congestionIndex: 88 },
    { time: '21:30', visitors: 54000, congestionIndex: 72 },
    { time: '22:00', visitors: 35000, congestionIndex: 45 }
  ]

  // Form states for Simulation Overrides
  const [simMatchID, setSimMatchID] = useState('')
  const [simHomeScore, setSimHomeScore] = useState(0)
  const [simAwayScore, setSimAwayScore] = useState(0)
  const [simElapsed, setSimElapsed] = useState('75:00')

  const [simEventMatchID, setSimEventMatchID] = useState('')
  const [simEventTime, setSimEventTime] = useState('78\'')
  const [simEventType, setSimEventType] = useState('goal')
  const [simEventDetail, setSimEventDetail] = useState('Goal scored by #10 (Messi)!')

  const [simParkingID, setSimParkingID] = useState('p1')
  const [simParkingOccupied, setSimParkingOccupied] = useState(485)

  const [simTransportID, setSimTransportID] = useState('t_line_1')
  const [simTransportDelay, setSimTransportDelay] = useState(0)
  const [simTransportStatus, setSimTransportStatus] = useState('on_time')

  // User management
  const [newUsername, setNewUsername] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [newUserRole, setNewUserRole] = useState('volunteer')
  const [newUserDept, setNewUserDept] = useState('Visitor Services')

  const loadData = async () => {
    try {
      const [metricsData, matchData, zoneData, parkData, transData, logsData] = await Promise.all([
        api.get<LiveTelemetryMetrics>('/metrics'),
        api.get<Match[]>('/matches'),
        api.get<CrowdZone[]>('/stadium/zones'),
        api.get<Parking[]>('/parking'),
        api.get<Transport[]>('/transport'),
        api.get<AuditLog[]>('/audit-logs')
      ])
      setMetrics(metricsData)
      setMatches(matchData)
      setZones(zoneData)
      setParking(parkData)
      setTransport(transData)
      setAuditLogs(logsData)

      // Autofill forms initially
      if (matchData.length > 0) {
        setSimMatchID(matchData[0].id)
        setSimHomeScore(matchData[0].homeScore)
        setSimAwayScore(matchData[0].awayScore)
        setSimEventMatchID(matchData[0].id)
      }
    } catch (err) {
      console.error(err)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const handleUpdateScore = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.put(`/matches/${simMatchID}`, {
        homeScore: simHomeScore,
        awayScore: simAwayScore,
        timeElapsed: simElapsed
      })
      loadData()
    } catch (err) {
      console.error(err)
    }
  }

  const handleAddEvent = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.post(`/matches/${simEventMatchID}/events`, {
        time: simEventTime,
        type: simEventType,
        detail: simEventDetail
      })
      loadData()
      setSimEventDetail('')
    } catch (err) {
      console.error(err)
    }
  }

  const handleUpdateParking = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.put(`/parking/${simParkingID}`, { occupied: simParkingOccupied })
      loadData()
    } catch (err) {
      console.error(err)
    }
  }

  const handleUpdateTransport = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.put(`/transport/${simTransportID}`, {
        delay: simTransportDelay,
        status: simTransportStatus
      })
      loadData()
    } catch (err) {
      console.error(err)
    }
  }

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.post('/users', {
        username: newUsername,
        password: newPassword,
        role: newUserRole,
        department: newUserDept
      })
      loadData()
      setNewUsername('')
      setNewPassword('')
    } catch (err) {
      console.error(err)
    }
  }

  // BarChart stand formatting
  const standsChartData = zones.map((z) => ({
    name: z.zoneName.split(' ')[0], // Stand Name first word
    Capacity: z.maxCapacity,
    Occupancy: z.currentOccupancy
  }))

  // Dynamic and Static options for CustomSelect components
  const matchOptions = matches.map((m) => ({
    value: m.id,
    label: `${m.homeTeam} vs ${m.awayTeam}`
  }))

  const parkingOptions = parking.map((p) => ({
    value: p.id,
    label: p.zoneName
  }))

  const transportOptions = transport.map((t) => ({
    value: t.id,
    label: t.routeName
  }))

  const eventTypeOptions = [
    { value: 'goal', label: 'Goal Scored' },
    { value: 'card', label: 'Referee Card' },
    { value: 'substitution', label: 'Substitution' },
    { value: 'info', label: 'General Info' }
  ]

  const transportStatusOptions = [
    { value: 'on_time', label: 'On Time' },
    { value: 'delayed', label: 'Delayed' },
    { value: 'suspended', label: 'Suspended' }
  ]

  const userRoleOptions = [
    { value: 'volunteer', label: 'Volunteer' },
    { value: 'security', label: 'Security' },
    { value: 'medical', label: 'Medical' },
    { value: 'cleaning', label: 'Cleaning' },
    { value: 'ops', label: 'Operations Coordinator' }
  ]

  return (
    <div className="space-y-8 font-sans">
      {/* Telemetry Stat Panels */}
      {metrics && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <Card className="p-4 space-y-2">
            <span className="text-[10px] uppercase font-mono font-bold text-zinc-400">Total Attendance</span>
            <div className="flex items-baseline gap-2">
              <span className="text-2xl font-bold font-mono">{metrics.attendance.toLocaleString()}</span>
              <span className="text-[10px] text-zinc-400">visitors</span>
            </div>
          </Card>
          <Card className="p-4 space-y-2">
            <span className="text-[10px] uppercase font-mono font-bold text-zinc-400">Smart Grid Energy</span>
            <div className="flex items-baseline gap-2">
              <span className="text-2xl font-bold font-mono">{metrics.energyUsageKw.toLocaleString()}</span>
              <span className="text-[10px] text-zinc-400">KW/h</span>
            </div>
          </Card>
          <Card className="p-4 space-y-2">
            <span className="text-[10px] uppercase font-mono font-bold text-zinc-400">Smart Water Loop</span>
            <div className="flex items-baseline gap-2">
              <span className="text-2xl font-bold font-mono">{metrics.waterConsumptionL.toLocaleString()}</span>
              <span className="text-[10px] text-zinc-400">Liters</span>
            </div>
          </Card>
          <Card className="p-4 space-y-2">
            <span className="text-[10px] uppercase font-mono font-bold text-zinc-400">Operations Tasks Done</span>
            <div className="flex items-baseline gap-2">
              <span className="text-2xl font-bold font-mono">
                {metrics.completedTasks}/{metrics.totalTasks}
              </span>
              <span className="text-[10px] text-zinc-400">dispatches</span>
            </div>
          </Card>
        </div>
      )}

      {/* Analytics Recharts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="p-6 space-y-3">
          <h3 className="font-bold text-sm text-neutral-800">Attendance Congestion Timeline</h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={graphTimelineData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorVis" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.4} />
                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#f4f4f5" />
                <XAxis dataKey="time" stroke="#a1a1aa" fontSize={10} tickLine={false} />
                <YAxis stroke="#a1a1aa" fontSize={10} tickLine={false} />
                <Tooltip />
                <Area type="monotone" dataKey="visitors" stroke="#2563eb" fillOpacity={1} fill="url(#colorVis)" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </Card>

        <Card className="p-6 space-y-3">
          <h3 className="font-bold text-sm text-neutral-800">Stadium Stand Capacities</h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={standsChartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f4f4f5" />
                <XAxis dataKey="name" stroke="#a1a1aa" fontSize={10} tickLine={false} />
                <YAxis stroke="#a1a1aa" fontSize={10} tickLine={false} />
                <Tooltip />
                <Bar dataKey="Capacity" fill="#e4e4e7" radius={[4, 4, 0, 0]} />
                <Bar dataKey="Occupancy" fill="#3b82f6" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </Card>
      </div>

      {/* Simulator & Operations Console Panels */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Simulation Controllers Column */}
        <div className="lg:col-span-2 space-y-6">
          <h2 className="text-base font-bold text-neutral-800 flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-brand-500" />
            <span>IoT Live Telemetry Simulator (Hackathon Overrides)</span>
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Match score controller */}
            <Card className="p-4 space-y-3">
              <span className="font-bold text-xs text-neutral-800">Match Score Override</span>
              <form onSubmit={handleUpdateScore} className="space-y-3 text-xs">
                <CustomSelect
                  label="Select Match"
                  value={simMatchID}
                  onChange={setSimMatchID}
                  options={matchOptions}
                  className="p-1 mt-0.5 text-xs border-zinc-200"
                  containerClassName="space-y-0.5"
                />
                <div className="grid grid-cols-3 gap-2">
                  <div>
                    <label className="text-[10px] text-zinc-400">Home Score</label>
                    <input
                      type="number"
                      value={simHomeScore || 0}
                      onChange={(e) => setSimHomeScore(parseInt(e.target.value) || 0)}
                      className="w-full h-9 border border-zinc-200 rounded-lg px-3 text-sm text-neutral-800 focus:outline-none focus:border-brand-500 mt-0.5"
                    />
                  </div>
                  <div>
                    <label className="text-[10px] text-zinc-400 font-medium">Away Score</label>
                    <input
                      type="number"
                      value={simAwayScore || 0}
                      onChange={(e) => setSimAwayScore(parseInt(e.target.value) || 0)}
                      className="w-full h-9 border border-zinc-200 rounded-lg px-3 text-sm text-neutral-800 focus:outline-none focus:border-brand-500 mt-0.5"
                    />
                  </div>
                  <div>
                    <label className="text-[10px] text-zinc-400 font-medium">Elapsed Time</label>
                    <input
                      type="text"
                      value={simElapsed}
                      onChange={(e) => setSimElapsed(e.target.value)}
                      className="w-full h-9 border border-zinc-200 rounded-lg px-3 text-sm text-neutral-800 focus:outline-none focus:border-brand-500 mt-0.5"
                    />
                  </div>
                </div>
                <button type="submit" className="w-full py-1.5 rounded bg-brand-600 text-white font-semibold">
                  Publish Score Update
                </button>
              </form>
            </Card>

            {/* Match events commentary controller */}
            <Card className="p-4 space-y-3">
              <span className="font-bold text-xs text-neutral-800">Add Live Event Commentary</span>
              <form onSubmit={handleAddEvent} className="space-y-3 text-xs">
                <CustomSelect
                  label="Select Match"
                  value={simEventMatchID}
                  onChange={setSimEventMatchID}
                  options={matchOptions}
                  className="p-1 mt-0.5 text-xs border-zinc-200"
                  containerClassName="space-y-0.5"
                />
                <div className="grid grid-cols-3 gap-2">
                  <div>
                    <label className="text-[10px] text-zinc-400 font-medium">Time</label>
                    <input
                      type="text"
                      value={simEventTime}
                      onChange={(e) => setSimEventTime(e.target.value)}
                      className="w-full h-9 border border-zinc-200 rounded-lg px-3 text-sm text-neutral-800 focus:outline-none focus:border-brand-500 mt-0.5"
                    />
                  </div>
                  <CustomSelect
                    label="Event Type"
                    value={simEventType}
                    onChange={setSimEventType}
                    options={eventTypeOptions}
                    className="p-1 text-xs border-zinc-200"
                    containerClassName="col-span-2 space-y-0.5"
                  />
                </div>
                <div>
                  <label className="text-[10px] text-zinc-400 font-medium">Commentary Text</label>
                  <input
                    type="text"
                    value={simEventDetail}
                    onChange={(e) => setSimEventDetail(e.target.value)}
                    className="w-full h-9 border border-zinc-200 rounded-lg px-3 text-sm text-neutral-800 focus:outline-none focus:border-brand-500 mt-0.5"
                    required
                  />
                </div>
                <button type="submit" className="w-full py-1.5 rounded bg-brand-600 text-white font-semibold">
                  Publish Live Commentary
                </button>
              </form>
            </Card>

            {/* Parking spots simulation */}
            <Card className="p-4 space-y-3">
              <span className="font-bold text-xs text-neutral-800">Parking Sensors Override</span>
              <form onSubmit={handleUpdateParking} className="space-y-3 text-xs">
                <div className="grid grid-cols-2 gap-2">
                  <CustomSelect
                    label="Parking Lot"
                    value={simParkingID}
                    onChange={setSimParkingID}
                    options={parkingOptions}
                    className="p-1 text-xs border-zinc-200"
                    containerClassName="space-y-0.5"
                  />
                  <div>
                    <label className="text-[10px] text-zinc-400 font-medium">Occupied Spots</label>
                    <input
                      type="number"
                      value={simParkingOccupied || 0}
                      onChange={(e) => setSimParkingOccupied(parseInt(e.target.value) || 0)}
                      className="w-full h-9 border border-zinc-200 rounded-lg px-3 text-sm text-neutral-800 focus:outline-none focus:border-brand-500 mt-0.5"
                    />
                  </div>
                </div>
                <button type="submit" className="w-full py-1.5 rounded bg-brand-600 text-white font-semibold">
                  Update Parking Count
                </button>
              </form>
            </Card>

            {/* Transport delay simulation */}
            <Card className="p-4 space-y-3">
              <span className="font-bold text-xs text-neutral-800">Transit Delays Override</span>
              <form onSubmit={handleUpdateTransport} className="space-y-3 text-xs">
                <div className="grid grid-cols-2 gap-2">
                  <CustomSelect
                    label="Transit Route"
                    value={simTransportID}
                    onChange={setSimTransportID}
                    options={transportOptions}
                    className="p-1 text-xs border-zinc-200"
                    containerClassName="space-y-0.5"
                  />
                  <div>
                    <label className="text-[10px] text-zinc-400 font-medium">Delay (Minutes)</label>
                    <input
                      type="number"
                      value={simTransportDelay || 0}
                      onChange={(e) => setSimTransportDelay(parseInt(e.target.value) || 0)}
                      className="w-full h-9 border border-zinc-200 rounded-lg px-3 text-sm text-neutral-800 focus:outline-none focus:border-brand-500 mt-0.5"
                    />
                  </div>
                </div>
                <CustomSelect
                  label="Route Status"
                  value={simTransportStatus}
                  onChange={setSimTransportStatus}
                  options={transportStatusOptions}
                  className="p-1 text-xs border-zinc-200"
                  containerClassName="space-y-0.5"
                />
                <button type="submit" className="w-full py-1.5 rounded bg-brand-600 text-white font-semibold">
                  Update Transit Feed
                </button>
              </form>
            </Card>
          </div>
        </div>

        {/* User accounts creator & audit log */}
        <div className="space-y-6">
          <h2 className="text-base font-bold text-neutral-800 flex items-center gap-2">
            <Users className="w-5 h-5 text-brand-500" />
            <span>Staff Administration</span>
          </h2>

          {/* User creator */}
          <Card className="p-4 space-y-3">
            <span className="font-bold text-xs text-neutral-800">Add Staff Account</span>
            <form onSubmit={handleCreateUser} className="space-y-3 text-xs">
              <div>
                <label className="text-[10px] text-zinc-400 font-medium">Username</label>
                <input
                  type="text"
                  value={newUsername}
                  onChange={(e) => setNewUsername(e.target.value)}
                  className="w-full h-9 border border-zinc-200 rounded-lg px-3 text-sm text-neutral-800 focus:outline-none focus:border-brand-500 mt-0.5"
                  required
                />
              </div>
              <div>
                <label className="text-[10px] text-zinc-400 font-medium">Password</label>
                <input
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="w-full h-9 border border-zinc-200 rounded-lg px-3 text-sm text-neutral-800 focus:outline-none focus:border-brand-500 mt-0.5"
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-2">
                <CustomSelect
                  label="Role"
                  value={newUserRole}
                  onChange={setNewUserRole}
                  options={userRoleOptions}
                  className="p-1 text-xs border-zinc-200"
                  containerClassName="space-y-0.5"
                />
                <div>
                  <label className="text-[10px] text-zinc-400 font-medium">Department</label>
                  <input
                    type="text"
                    value={newUserDept}
                    onChange={(e) => setNewUserDept(e.target.value)}
                    className="w-full h-9 border border-zinc-200 rounded-lg px-3 text-sm text-neutral-800 focus:outline-none focus:border-brand-500 mt-0.5"
                    placeholder="Visitor Services"
                  />
                </div>
              </div>
              <button type="submit" className="w-full py-1.5 rounded bg-brand-600 text-white font-semibold flex items-center justify-center gap-1">
                <Plus className="w-3.5 h-3.5" />
                <span>Register User</span>
              </button>
            </form>
          </Card>

          {/* Audit Logs */}
          <div className="space-y-3">
            <span className="font-bold text-xs text-neutral-800 block">Console Action Logs</span>
            <div className="bg-white border border-zinc-200 rounded-xl p-4 max-h-[160px] overflow-y-auto space-y-2 divide-y divide-zinc-50 shadow-sm">
              {auditLogs.length > 0 ? (
                [...auditLogs].reverse().map((log) => (
                  <div key={log.id} className="text-[10px] pt-1.5 first:pt-0 leading-normal text-zinc-500">
                    <span className="font-semibold text-neutral-800">@{log.username}</span> overrides{' '}
                    <span className="font-mono text-brand-600">{log.resource}:{log.details}</span> ({log.action})
                  </div>
                ))
              ) : (
                <p className="text-[10px] text-zinc-400 text-center py-4">No admin override audits logged.</p>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
