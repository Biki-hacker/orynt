import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Match, Alert, Announcement } from '../../types'
import { api } from '../../services/api'
import { wsService } from '../../services/websocket'
import { Play, Calendar, Bell, Trophy, Volume2, ArrowRight } from 'lucide-react'

export const Home: React.FC = () => {
  const [matches, setMatches] = useState<Match[]>([])
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [announcements, setAnnouncements] = useState<Announcement[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const loadData = async () => {
      try {
        const [matchData, alertData, annData] = await Promise.all([
          api.get<Match[]>('/matches'),
          api.get<Alert[]>('/alerts?active=true'),
          api.get<Announcement[]>('/announcements?role=public')
        ])
        setMatches(matchData)
        setAlerts(alertData)
        setAnnouncements(annData)
      } catch (err) {
        console.error(err)
      } finally {
        setLoading(false)
      }
    }
    loadData()

    // WebSockets event handlers
    const unsubScore = wsService.subscribe('match_score', (m: Match) => {
      setMatches((prev) => prev.map((item) => (item.id === m.id ? m : item)))
    })

    const unsubAlert = wsService.subscribe('alert', (a: Alert) => {
      if (a.active) {
        setAlerts((prev) => [a, ...prev])
      }
    })

    const unsubAlertResolve = wsService.subscribe('alert_resolved', (payload: { id: string }) => {
      setAlerts((prev) => prev.filter((a) => a.id !== payload.id))
    })

    const unsubAnn = wsService.subscribe('announcement', (a: Announcement) => {
      setAnnouncements((prev) => [a, ...prev])
    })

    return () => {
      unsubScore()
      unsubAlert()
      unsubAlertResolve()
      unsubAnn()
    }
  }, [])

  if (loading) {
    return (
      <div className="space-y-6 animate-pulse">
        <div className="h-40 bg-zinc-200 rounded-xl" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="h-64 bg-zinc-200 rounded-lg" />
          <div className="h-64 bg-zinc-200 rounded-lg" />
        </div>
      </div>
    )
  }

  const liveMatches = matches.filter((m) => m.status === 'live')
  const upcomingMatches = matches.filter((m) => m.status === 'scheduled')

  return (
    <div className="space-y-8">
      {/* Welcome Banner */}
      <div className="bg-zinc-900 text-white rounded-2xl p-8 shadow-lg relative overflow-hidden flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div className="relative z-10 space-y-2 max-w-xl">
          <span className="text-xs uppercase font-mono tracking-wider font-semibold text-brand-500">
            Smart Venue Hub
          </span>
          <h1 className="text-3xl font-extrabold tracking-tight">Orynt Arena Command</h1>
          <p className="text-zinc-400 text-sm leading-relaxed">
            Welcome to the Smart Tournament Dashboard. Keep track of live stadium occupancy stands, shuttle delays, parking spaces, food concessions wait times, and emergency alerts.
          </p>
        </div>
        <div className="flex gap-3">
          <Link
            to="/map"
            className="bg-brand-600 hover:bg-brand-700 text-white text-sm font-semibold rounded-lg px-5 py-2.5 transition-colors shadow-md flex items-center gap-1.5"
          >
            <span>Interactive Map</span>
            <ArrowRight className="w-4 h-4" />
          </Link>
          <Link
            to="/transport"
            className="bg-zinc-800 hover:bg-zinc-700 border border-zinc-700 text-white text-sm font-semibold rounded-lg px-5 py-2.5 transition-colors"
          >
            Parking & Transit
          </Link>
        </div>
        {/* Subtle grid background vector */}
        <div className="absolute inset-0 opacity-10 bg-[linear-gradient(to_right,#808080_1px,transparent_1px),linear-gradient(to_bottom,#808080_1px,transparent_1px)] bg-[size:24px_24px] pointer-events-none" />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Live Score Column */}
        <div className="lg:col-span-2 space-y-6">
          <div>
            <h2 className="text-lg font-bold text-neutral-800 flex items-center gap-2">
              <Trophy className="w-5 h-5 text-amber-500" />
              <span>Live Matches</span>
            </h2>
            <p className="text-xs text-zinc-500 mt-0.5">Real-time match scoring and event updates.</p>
          </div>

          {liveMatches.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {liveMatches.map((m) => (
                <div key={m.id} className="bg-white border border-zinc-200 rounded-xl p-5 shadow-sm hover:shadow-md transition-shadow relative overflow-hidden flex flex-col justify-between">
                  <div className="flex justify-between items-center">
                    <span className="flex items-center gap-1 px-2 py-0.5 rounded bg-red-50 text-[10px] font-bold text-red-600 uppercase">
                      <span className="w-1.5 h-1.5 rounded-full bg-red-600 animate-pulse-slow" />
                      <span>Live {m.timeElapsed}</span>
                    </span>
                    <span className="text-[10px] text-zinc-400 font-mono">Match ID: {m.id}</span>
                  </div>

                  <div className="my-6 space-y-3">
                    <div className="flex justify-between items-center font-bold text-lg text-neutral-800">
                      <span>{m.homeTeam}</span>
                      <span className="text-2xl font-mono text-brand-600">{m.homeScore}</span>
                    </div>
                    <div className="flex justify-between items-center font-bold text-lg text-neutral-800">
                      <span>{m.awayTeam}</span>
                      <span className="text-2xl font-mono text-brand-600">{m.awayScore}</span>
                    </div>
                  </div>

                  <Link
                    to={`/match/${m.id}`}
                    className="w-full flex items-center justify-center gap-1 bg-zinc-50 hover:bg-zinc-100 text-xs font-bold text-zinc-700 py-2 border border-zinc-200 rounded-lg transition-colors"
                  >
                    <Play className="w-3.5 h-3.5 fill-current" />
                    <span>View Commentary & Statistics</span>
                  </Link>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-white border border-zinc-200 rounded-xl p-8 text-center text-zinc-400 py-12">
              <Trophy className="w-10 h-10 mx-auto mb-2 text-zinc-300 stroke-1" />
              <p className="text-sm">There are no live matches right now. Upcoming fixtures are listed below.</p>
            </div>
          )}

          {/* Upcoming Matches */}
          {upcomingMatches.length > 0 && (
            <div className="space-y-3">
              <h3 className="text-sm font-semibold uppercase font-mono text-zinc-500 tracking-wider">
                Upcoming Schedules
              </h3>
              <div className="bg-white border border-zinc-200 rounded-xl divide-y divide-zinc-100 overflow-hidden shadow-sm">
                {upcomingMatches.map((m) => (
                  <div key={m.id} className="p-4 flex items-center justify-between text-sm">
                    <div className="flex items-center gap-3">
                      <Calendar className="w-4 h-4 text-zinc-400" />
                      <span className="font-semibold text-neutral-800">
                        {m.homeTeam} <span className="text-zinc-400 font-normal">vs</span> {m.awayTeam}
                      </span>
                    </div>
                    <span className="text-xs text-zinc-500 font-mono">
                      {new Date(m.scheduledAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Alerts & Announcement Columns */}
        <div className="space-y-8">
          {/* Active Alerts */}
          <div className="space-y-4">
            <div>
              <h2 className="text-lg font-bold text-neutral-800 flex items-center gap-2">
                <Bell className="w-5 h-5 text-red-500" />
                <span>Crowd Alerts</span>
              </h2>
              <p className="text-xs text-zinc-500 mt-0.5">Important security and transport updates.</p>
            </div>

            {alerts.length > 0 ? (
              <div className="space-y-3">
                {alerts.map((a) => (
                  <div
                    key={a.id}
                    className={`p-4 rounded-xl border flex gap-3 ${
                      a.severity === 'critical'
                        ? 'bg-red-50 border-red-200 text-red-800'
                        : 'bg-amber-50 border-amber-200 text-amber-800'
                    }`}
                  >
                    <div className="flex-1 space-y-1">
                      <h4 className="text-sm font-bold flex items-center gap-1.5">
                        <span className={`w-2 h-2 rounded-full ${a.severity === 'critical' ? 'bg-red-600' : 'bg-amber-500'}`} />
                        {a.title}
                      </h4>
                      <p className="text-xs leading-relaxed opacity-90">{a.content}</p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="bg-white border border-zinc-200 rounded-xl p-5 text-center text-zinc-400 py-8">
                <p className="text-xs">All channels normal. No active crowd alerts.</p>
              </div>
            )}
          </div>

          {/* Announcements */}
          <div className="space-y-4">
            <div>
              <h2 className="text-lg font-bold text-neutral-800 flex items-center gap-2">
                <Volume2 className="w-5 h-5 text-brand-500" />
                <span>Announcements</span>
              </h2>
              <p className="text-xs text-zinc-500 mt-0.5">Official venue guidelines and notifications.</p>
            </div>

            {announcements.length > 0 ? (
              <div className="bg-white border border-zinc-200 rounded-xl divide-y divide-zinc-100 overflow-hidden shadow-sm">
                {announcements.map((ann) => (
                  <div key={ann.id} className="p-4 space-y-1">
                    <span className="text-[9px] uppercase font-mono tracking-wider font-semibold text-brand-500">
                      {new Date(ann.createdAt).toLocaleDateString()}
                    </span>
                    <h4 className="text-sm font-bold text-neutral-800">{ann.title}</h4>
                    <p className="text-xs text-zinc-500 leading-relaxed">{ann.content}</p>
                  </div>
                ))}
              </div>
            ) : (
              <div className="bg-white border border-zinc-200 rounded-xl p-5 text-center text-zinc-400 py-8">
                <p className="text-xs">No new public announcements.</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
