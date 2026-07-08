import React, { useState, useEffect } from 'react'
import { Parking, Transport } from '../../types'
import { api } from '../../services/api'
import { wsService } from '../../services/websocket'
import { Badge } from '../../components/ui'
import { Car, Bus, Timer, MapPin, Gauge } from 'lucide-react'

export const TransportParking: React.FC = () => {
  const [parking, setParking] = useState<Parking[]>([])
  const [transport, setTransport] = useState<Transport[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const loadData = async () => {
      try {
        const [parkingData, transportData] = await Promise.all([
          api.get<Parking[]>('/parking'),
          api.get<Transport[]>('/transport')
        ])
        setParking(parkingData)
        setTransport(transportData)
      } catch (err) {
        console.error(err)
      } finally {
        setLoading(false)
      }
    }
    loadData()

    // WebSockets real-time listeners
    const unsubParking = wsService.subscribe('parking', (lot: Parking) => {
      setParking((prev) => prev.map((item) => (item.id === lot.id ? lot : item)))
    })

    const unsubTransport = wsService.subscribe('transport', (route: Transport) => {
      setTransport((prev) => prev.map((item) => (item.id === route.id ? route : item)))
    })

    return () => {
      unsubParking()
      unsubTransport()
    }
  }, [])

  if (loading) {
    return (
      <div className="space-y-6 animate-pulse">
        <div className="h-48 bg-zinc-200 rounded-xl" />
        <div className="h-48 bg-zinc-200 rounded-xl" />
      </div>
    )
  }

  return (
    <div className="space-y-8 font-sans">
      <div className="space-y-1">
        <span className="text-xs uppercase font-mono font-bold text-brand-600 tracking-wider">
          Orynt Transit Hub
        </span>
        <h1 className="text-2xl font-extrabold tracking-tight text-neutral-800">Transport & Parking Telemetry</h1>
        <p className="text-zinc-500 text-sm max-w-2xl leading-relaxed">
          Monitor live shuttle delays, train schedules, and vacant parking spots near Orynt Stadium. Updates are pushed live via smart parking and transit sensors.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Smart Parking Lots Section */}
        <div className="space-y-4">
          <h2 className="text-lg font-bold text-neutral-800 flex items-center gap-2">
            <Car className="w-5 h-5 text-brand-500" />
            <span>Parking Space Vacancies</span>
          </h2>

          <div className="space-y-4">
            {parking.map((p) => {
              const occupiedPct = Math.round((p.occupiedSpots / p.totalSpots) * 100)
              const freeSpots = p.totalSpots - p.occupiedSpots
              const isFull = freeSpots <= 10

              return (
                <div key={p.id} className="bg-white border border-zinc-200 rounded-xl p-5 shadow-sm space-y-3">
                  <div className="flex justify-between items-start">
                    <div>
                      <h3 className="font-bold text-sm text-neutral-800">{p.zoneName}</h3>
                      <div className="flex items-center gap-1 text-zinc-400 text-xs mt-0.5">
                        <MapPin className="w-3.5 h-3.5" />
                        <span>{p.location}</span>
                      </div>
                    </div>
                    <Badge variant={isFull ? 'error' : p.type === 'vip' ? 'info' : 'success'}>
                      {isFull ? 'Lot Full' : p.type.toUpperCase()}
                    </Badge>
                  </div>

                  {/* Occupancy bar */}
                  <div className="space-y-1.5">
                    <div className="flex justify-between text-xs text-zinc-500 font-medium">
                      <span>Occupancy: {occupiedPct}%</span>
                      <span>{freeSpots} / {p.totalSpots} spots vacant</span>
                    </div>
                    <div className="w-full h-2.5 bg-zinc-100 rounded-full overflow-hidden">
                      <div
                        className={`h-full rounded-full transition-all duration-300 ${
                          occupiedPct >= 90 ? 'bg-red-500' : occupiedPct >= 65 ? 'bg-amber-400' : 'bg-emerald-500'
                        }`}
                        style={{ width: `${occupiedPct}%` }}
                      />
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </div>

        {/* Live Transit Schedules Section */}
        <div className="space-y-4">
          <h2 className="text-lg font-bold text-neutral-800 flex items-center gap-2">
            <Bus className="w-5 h-5 text-brand-500" />
            <span>Live Shuttle & Transit Timings</span>
          </h2>

          <div className="space-y-4">
            {transport.map((t) => {
              const isDelayed = t.status === 'delayed'
              const isSuspended = t.status === 'suspended'

              return (
                <div key={t.id} className="bg-white border border-zinc-200 rounded-xl p-5 shadow-sm flex items-center justify-between gap-4">
                  <div className="flex items-center gap-4">
                    <div className={`w-10 h-10 rounded-lg flex items-center justify-center text-white ${
                      t.mode === 'train' ? 'bg-blue-600' : t.mode === 'bus' ? 'bg-indigo-600' : 'bg-emerald-600'
                    }`}>
                      <Bus className="w-5 h-5" />
                    </div>
                    <div>
                      <h3 className="font-bold text-sm text-neutral-800">{t.routeName}</h3>
                      <div className="flex items-center gap-1.5 text-zinc-400 text-xs mt-0.5">
                        <Timer className="w-3.5 h-3.5" />
                        <span>Next arrival in {new Date(t.nextArrival).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                      </div>
                    </div>
                  </div>

                  <div className="text-right space-y-1">
                    <Badge variant={isSuspended ? 'error' : isDelayed ? 'warning' : 'success'}>
                      {isSuspended ? 'SUSPENDED' : isDelayed ? `DELAYED ${t.delayMinutes}M` : 'ON TIME'}
                    </Badge>
                    <div className="flex items-center gap-1 text-[10px] text-zinc-400 font-medium justify-end">
                      <Gauge className="w-3.5 h-3.5 text-zinc-300" />
                      <span>Capacity: <span className="text-neutral-700 font-semibold uppercase">{t.capacity}</span></span>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}
