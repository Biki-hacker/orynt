import React, { useState, useEffect } from 'react'
import { CrowdZone, Facility, ExitRoute } from '../../types'
import { api } from '../../services/api'
import { wsService } from '../../services/websocket'
import { Shield, AlertTriangle, Accessibility, Coffee, BadgeAlert, Compass, Navigation } from 'lucide-react'

interface StadiumMapProps {
  onSectionSelect?: (zone: CrowdZone | null) => void;
}

export const StadiumMap: React.FC<StadiumMapProps> = ({ onSectionSelect }) => {
  const [zones, setZones] = useState<CrowdZone[]>([])
  const [selectedZone, setSelectedZone] = useState<CrowdZone | null>(null)
  const [activeFilter, setActiveFilter] = useState<'all' | 'food' | 'restroom' | 'medical' | 'exit' | 'accessible'>('all')
  const [exitRoutes, setExitRoutes] = useState<ExitRoute[]>([])
  const [showExitAssistance, setShowExitAssistance] = useState<boolean>(false)

  useEffect(() => {
    // Fetch initial zones
    const loadZones = async () => {
      try {
        const data = await api.get<CrowdZone[]>('/stadium/zones')
        setZones(data)
      } catch (err) {
        console.error(err)
      }
    }

    const loadExitRoutes = async () => {
      try {
        const data = await api.get<ExitRoute[]>('/stadium/exit-routing')
        setExitRoutes(data)
      } catch (err) {
        console.error(err)
      }
    }

    loadZones()
    loadExitRoutes()

    // Listen to real-time density updates
    const unsub = wsService.subscribe('crowd_density', (updatedZone: CrowdZone) => {
      setZones((prev) => prev.map((z) => (z.id === updatedZone.id ? updatedZone : z)))
      setSelectedZone((curr) => (curr && curr.id === updatedZone.id ? updatedZone : curr))
      loadExitRoutes() // Refresh exit routes when crowd density changes dynamically
    })

    return () => unsub()
  }, [])

  const getZoneDensityColor = (zoneId: string) => {
    const zone = zones.find((z) => z.id === zoneId)
    if (!zone) return 'fill-zinc-200 stroke-zinc-300'
    if (zone.densityLevel === 'high') return 'fill-red-400/80 hover:fill-red-400 stroke-red-600'
    if (zone.densityLevel === 'medium') return 'fill-amber-300/80 hover:fill-amber-300 stroke-amber-500'
    return 'fill-emerald-300/80 hover:fill-emerald-300 stroke-emerald-500'
  }

  const getZoneOccupancyPct = (zoneId: string) => {
    const zone = zones.find((z) => z.id === zoneId)
    if (!zone) return 0
    return Math.round((zone.currentOccupancy / zone.maxCapacity) * 100)
  }

  const handleZoneClick = (zoneId: string) => {
    const zone = zones.find((z) => z.id === zoneId)
    if (zone) {
      setSelectedZone(zone)
      if (onSectionSelect) onSectionSelect(zone)
    }
  }

  // Facilities coordinate mapping for SVG overlays
  const facilities: (Facility & { x: number; y: number })[] = [
    { id: 'f1', name: 'Corner Grill', location: 'Section 102', type: 'food', x: 200, y: 70 },
    { id: 'f2', name: 'Slam Dunk Brews', location: 'Section 115', type: 'food', x: 420, y: 200 },
    { id: 'f3', name: 'Vegan Bites', location: 'Section 204', type: 'food', x: 180, y: 330 },
    { id: 'r1', name: 'Washrooms East', location: 'Section 108', x: 380, y: 100, type: 'restroom' },
    { id: 'r2', name: 'Washrooms West', location: 'Section 218', x: 120, y: 220, type: 'restroom' },
    { id: 'm1', name: 'First Aid Station A', location: 'Gate 2 Floor 1', x: 290, y: 60, type: 'medical' },
    { id: 'm2', name: 'Response Center', location: 'Gate 5 Ground', x: 300, y: 340, type: 'medical' },
    { id: 'e1', name: 'Exit Gate A', location: 'North Side', x: 300, y: 15, type: 'exit' },
    { id: 'e2', name: 'Exit Gate B (Access)', location: 'South-West', x: 80, y: 350, type: 'exit' }
  ]

  const filters = [
    { id: 'all', label: 'All Zones', icon: Compass },
    { id: 'food', label: 'Food Stalls', icon: Coffee },
    { id: 'accessible', label: 'Accessible Routes', icon: Accessibility },
    { id: 'medical', label: 'First Aid', icon: BadgeAlert }
  ]

  const selectedRoute = selectedZone ? exitRoutes.find((r) => r.zoneId === selectedZone.id) : null

  return (
    <div className="flex flex-col lg:flex-row gap-6 bg-white border border-zinc-200 rounded-lg p-6 shadow-sm">
      {/* SVG Stadium Map */}
      <div className="flex-1 flex flex-col items-center">
        {/* Toggle Filters */}
        <div className="flex flex-wrap gap-2 mb-6 w-full justify-center">
          {filters.map((f) => {
            const Icon = f.icon
            return (
              <button
                key={f.id}
                onClick={() => setActiveFilter(f.id as any)}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-semibold border transition-all duration-150 ${
                  activeFilter === f.id
                    ? 'bg-brand-600 border-brand-600 text-white shadow-sm'
                    : 'bg-zinc-50 border-zinc-200 text-zinc-600 hover:bg-zinc-100'
                }`}
              >
                <Icon className="w-3.5 h-3.5" />
                <span>{f.label}</span>
              </button>
            )
          })}
          
          <button
            onClick={() => setShowExitAssistance(!showExitAssistance)}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-semibold border transition-all duration-150 ${
              showExitAssistance
                ? 'bg-blue-600 border-blue-600 text-white shadow-sm hover:bg-blue-700'
                : 'bg-zinc-50 border-zinc-200 text-zinc-600 hover:bg-zinc-100'
            }`}
          >
            <Navigation className="w-3.5 h-3.5 rotate-45" />
            <span>Exit Routing Assistance</span>
          </button>
        </div>

        {/* Stadium Layout */}
        <div className="relative w-full max-w-lg aspect-square border border-zinc-100 rounded-lg bg-zinc-50 flex items-center justify-center p-4">
          <svg viewBox="0 0 500 400" className="w-full h-full select-none">
            {/* Background Details */}
            <rect width="500" height="400" fill="none" />

            {/* Stadium Stands */}
            {/* North Stand */}
            <path
              d="M 100 120 A 160 120 0 0 1 400 120 L 450 80 A 240 180 0 0 0 50 80 Z"
              className={`cursor-pointer transition-colors duration-150 stroke-2 ${getZoneDensityColor('z1')}`}
              onClick={() => handleZoneClick('z1')}
            />
            {/* East Stand */}
            <path
              d="M 400 120 A 160 120 0 0 1 400 280 L 450 320 A 240 180 0 0 0 450 80 Z"
              className={`cursor-pointer transition-colors duration-150 stroke-2 ${getZoneDensityColor('z2')}`}
              onClick={() => handleZoneClick('z2')}
            />
            {/* South Stand */}
            <path
              d="M 400 280 A 160 120 0 0 1 100 280 L 50 320 A 240 180 0 0 0 450 320 Z"
              className={`cursor-pointer transition-colors duration-150 stroke-2 ${getZoneDensityColor('z3')}`}
              onClick={() => handleZoneClick('z3')}
            />
            {/* West Stand */}
            <path
              d="M 100 280 A 160 120 0 0 1 100 120 L 50 80 A 240 180 0 0 0 50 320 Z"
              className={`cursor-pointer transition-colors duration-150 stroke-2 ${getZoneDensityColor('z4')}`}
              onClick={() => handleZoneClick('z4')}
            />
            {/* VIP Suites (Inner Ring) */}
            <ellipse
              cx="250"
              cy="200"
              rx="110"
              ry="75"
              className={`cursor-pointer transition-colors duration-150 stroke-2 ${getZoneDensityColor('z5')}`}
              onClick={() => handleZoneClick('z5')}
            />

            {/* Central Football Field */}
            <ellipse cx="250" cy="200" rx="80" ry="50" fill="#10b981" stroke="#ffffff" strokeWidth="2" />
            {/* Field Inner details */}
            <line x1="250" y1="150" x2="250" y2="250" stroke="#ffffff" strokeWidth="1.5" />
            <circle cx="250" cy="200" r="20" fill="none" stroke="#ffffff" strokeWidth="1.5" />

            {/* Accessible Elevator routes */}
            {activeFilter === 'accessible' && (
              <g className="animate-pulse">
                {/* Draw Route from Gate B to elevator */}
                <path d="M 80 350 L 150 280 L 250 280" fill="none" stroke="#3b82f6" strokeWidth="4" strokeLinecap="round" strokeDasharray="6,4" />
                <path d="M 300 15 L 300 120 L 250 120" fill="none" stroke="#3b82f6" strokeWidth="4" strokeLinecap="round" strokeDasharray="6,4" />
                {/* Elevator labels */}
                <circle cx="150" cy="280" r="6" fill="#3b82f6" />
                <circle cx="300" cy="120" r="6" fill="#3b82f6" />
              </g>
            )}

            {/* Defs for arrow markers */}
            <defs>
              <marker
                id="exit-arrow"
                viewBox="0 0 10 10"
                refX="6"
                refY="5"
                markerWidth="5"
                markerHeight="5"
                orient="auto-start-reverse"
              >
                <path d="M 0 2 L 7 5 L 0 8 z" fill="context-stroke" />
              </marker>
            </defs>

            {/* CSS Animation for exit flow */}
            <style>{`
              @keyframes exitFlow {
                to {
                  stroke-dashoffset: -20;
                }
              }
              .exit-route-path {
                stroke-dasharray: 8, 6;
                animation: exitFlow 1s linear infinite;
              }
            `}</style>

            {/* Smart Exit Routing Paths */}
            {showExitAssistance && exitRoutes.map((route) => {
              const pathString = route.pathPoints.reduce((acc, p, idx) => {
                return acc + (idx === 0 ? `M ${p.x} ${p.y}` : ` L ${p.x} ${p.y}`)
              }, '')

              const isSelected = selectedZone?.id === route.zoneId
              
              // Set stroke colors based on congestion
              let strokeColor = '#3b82f6' // clear / low
              if (route.congestionLevel === 'high') strokeColor = '#ef4444'
              else if (route.congestionLevel === 'medium') strokeColor = '#f59e0b'

              return (
                <path
                  key={route.zoneId}
                  d={pathString}
                  fill="none"
                  stroke={strokeColor}
                  strokeWidth={isSelected ? 5 : 2.5}
                  strokeOpacity={selectedZone ? (isSelected ? 1.0 : 0.2) : 0.75}
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  className="exit-route-path"
                  style={{
                    filter: isSelected ? `drop-shadow(0 0 4px ${strokeColor})` : 'none',
                    transition: 'stroke-width 0.2s, stroke-opacity 0.2s',
                  }}
                  markerEnd="url(#exit-arrow)"
                />
              )
            })}

            {/* Facilities Overlay Icons */}
            {facilities.map((fac) => {
              if (activeFilter !== 'all' && activeFilter !== fac.type && activeFilter !== 'accessible') {
                return null
              }

              // Set colors
              let color = 'bg-zinc-500'
              if (fac.type === 'food') color = 'bg-amber-500'
              if (fac.type === 'restroom') color = 'bg-blue-500'
              if (fac.type === 'medical') color = 'bg-red-500'
              if (fac.type === 'exit') color = 'bg-zinc-700'

              return (
                <foreignObject key={fac.id} x={fac.x - 10} y={fac.y - 10} width="20" height="20">
                  <div
                    title={`${fac.name} (${fac.location})`}
                    className={`w-5 h-5 rounded-full ${color} text-white flex items-center justify-center text-[8px] font-bold shadow-md cursor-help border border-white hover:scale-125 transition-transform duration-100`}
                  >
                    {fac.type === 'food' && 'F'}
                    {fac.type === 'restroom' && 'WC'}
                    {fac.type === 'medical' && 'H'}
                    {fac.type === 'exit' && 'E'}
                  </div>
                </foreignObject>
              )
            })}
          </svg>
        </div>
      </div>

      {/* Map Sidebar Info Panel */}
      <div className="w-full lg:w-72 border-t lg:border-t-0 lg:border-l border-zinc-100 pt-6 lg:pt-0 lg:pl-6 flex flex-col justify-between">
        {selectedZone ? (
          <div>
            <span className="text-xs uppercase font-mono font-bold text-brand-600 tracking-wider">
              Selected Section
            </span>
            <h3 className="font-bold text-lg text-neutral-800 mt-0.5">{selectedZone.zoneName}</h3>

            <div className="mt-5 space-y-4">
              <div>
                <span className="text-xs text-zinc-400">Crowd Density Status</span>
                <div className="flex items-center gap-2 mt-1">
                  <div
                    className={`w-3.5 h-3.5 rounded-full ${
                      selectedZone.densityLevel === 'high'
                        ? 'bg-red-500'
                        : selectedZone.densityLevel === 'medium'
                        ? 'bg-amber-500'
                        : 'bg-emerald-500'
                    }`}
                  />
                  <span className="text-sm font-semibold capitalize text-neutral-800">
                    {selectedZone.densityLevel} Density ({getZoneOccupancyPct(selectedZone.id)}% Capacity)
                  </span>
                </div>
              </div>

              <div>
                <span className="text-xs text-zinc-400">Section Occupancy Count</span>
                <p className="text-sm font-medium text-neutral-800 mt-0.5">
                  {selectedZone.currentOccupancy.toLocaleString()} / {selectedZone.maxCapacity.toLocaleString()} visitors
                </p>
              </div>

              {selectedZone.densityLevel === 'high' && (
                <div className="flex gap-2 p-3 bg-red-50 rounded border border-red-100 mt-2">
                  <AlertTriangle className="w-4 h-4 text-red-600 flex-shrink-0 mt-0.5" />
                  <p className="text-xs text-red-700 leading-normal">
                    <strong>High Congestion Warning:</strong> Exit routes in this stand may experience minor queue delay times. Follow green arrows to exits.
                  </p>
                </div>
              )}

              {selectedZone.id === 'z1' && (
                <div className="flex gap-2 p-3 bg-blue-50 rounded border border-blue-100 mt-2">
                  <Accessibility className="w-4 h-4 text-blue-600 flex-shrink-0 mt-0.5" />
                  <p className="text-xs text-blue-700 leading-normal">
                    <strong>Accessible Route:</strong> Elevator 1 is active directly on the concourse level of this stand.
                  </p>
                </div>
              )}

              {showExitAssistance && selectedRoute && (
                <div className="mt-4 p-3 bg-zinc-50 border border-zinc-200 rounded-lg">
                  <div className="flex items-center gap-1.5 text-xs font-bold text-zinc-500 uppercase tracking-wide">
                    <Navigation className="w-3.5 h-3.5 text-blue-600 rotate-45 animate-pulse" />
                    <span>Smart Exit Routing</span>
                  </div>
                  
                  <div className="mt-2.5 space-y-2">
                    <div>
                      <span className="text-[10px] text-zinc-400 block uppercase font-medium">Recommended Exit</span>
                      <span className="text-sm font-bold text-neutral-800">
                        {selectedRoute.exitName}
                      </span>
                    </div>
                    
                    <div>
                      <span className="text-[10px] text-zinc-400 block uppercase font-medium">Route Congestion</span>
                      <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full inline-block mt-0.5 ${
                        selectedRoute.congestionLevel === 'high'
                          ? 'bg-red-100 text-red-700'
                          : selectedRoute.congestionLevel === 'medium'
                          ? 'bg-amber-100 text-amber-700'
                          : 'bg-emerald-100 text-emerald-700'
                      }`}>
                        {selectedRoute.congestionLevel.toUpperCase()}
                      </span>
                    </div>
                    
                    {selectedZone.id === 'z1' && selectedRoute.recommendedExit === 'e2' && (
                      <div className="p-2 bg-blue-50 border border-blue-100 rounded text-[11px] text-blue-800 leading-normal flex gap-1.5 items-start">
                        <AlertTriangle className="w-3.5 h-3.5 text-blue-600 flex-shrink-0 mt-0.5" />
                        <span>
                          <strong>Dynamic Divert Active:</strong> Rerouted to Gate B to prevent crowd crush at Gate A.
                        </span>
                      </div>
                    )}

                    <div>
                      <span className="text-[10px] text-zinc-400 block uppercase font-medium">Wayfinding Guidance</span>
                      <p className="text-xs text-neutral-600 leading-normal mt-0.5">
                        {selectedRoute.reason}
                      </p>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-center text-zinc-400 py-12">
            <Compass className="w-10 h-10 mb-2 stroke-1" />
            <p className="text-sm">Click any stand on the map to display density sensors and occupancy metrics.</p>
          </div>
        )}

        <div className="mt-8 pt-4 border-t border-zinc-100">
          <div className="flex items-center gap-1.5 text-xs text-zinc-400">
            <Shield className="w-3.5 h-3.5 text-brand-500" />
            <span>Automatic refresh: Live IoT Sensors</span>
          </div>
        </div>
      </div>
    </div>
  )
}
