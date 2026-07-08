import React, { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { Match } from '../../types'
import { api } from '../../services/api'
import { wsService } from '../../services/websocket'
import { ArrowLeft, Clock, Activity, Flag } from 'lucide-react'

export const MatchDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const [match, setMatch] = useState<Match | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const loadMatch = async () => {
      try {
        const matches = await api.get<Match[]>('/matches')
        const found = matches.find((m) => m.id === id)
        if (found) setMatch(found)
      } catch (err) {
        console.error(err)
      } finally {
        setLoading(false)
      }
    }
    loadMatch()

    // Subscribe to updates for this match
    const unsubScore = wsService.subscribe('match_score', (m: Match) => {
      if (m.id === id) {
        setMatch(m)
      }
    })

    const unsubEvent = wsService.subscribe('match_event', (m: Match) => {
      if (m.id === id) {
        setMatch(m)
      }
    })

    return () => {
      unsubScore()
      unsubEvent()
    }
  }, [id])

  if (loading) {
    return <div className="animate-pulse h-96 bg-zinc-200 rounded-xl" />
  }

  if (!match) {
    return (
      <div className="text-center py-12 bg-white border border-zinc-200 rounded-xl p-8 max-w-md mx-auto">
        <h3 className="font-bold text-lg text-neutral-800">Match Not Found</h3>
        <p className="text-zinc-500 text-sm mt-1">This match schedule does not exist or was deleted.</p>
        <Link to="/" className="mt-4 inline-block text-sm font-semibold text-brand-600 hover:underline">
          Return Home
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-8 font-sans max-w-4xl mx-auto">
      {/* Back button */}
      <Link to="/" className="inline-flex items-center gap-1.5 text-zinc-500 hover:text-neutral-800 transition-colors text-sm font-medium">
        <ArrowLeft className="w-4 h-4" />
        <span>Back to Matches</span>
      </Link>

      {/* Main Scoreboard Card */}
      <div className="bg-white border border-zinc-200 rounded-xl overflow-hidden shadow-sm">
        {/* Banner header */}
        <div className="bg-zinc-900 text-zinc-400 px-6 py-3 flex items-center justify-between text-xs font-semibold">
          <span className="uppercase tracking-widest text-[10px] text-brand-500">Live Telemetry Feed</span>
          <span className="flex items-center gap-1">
            <Clock className="w-3.5 h-3.5" />
            <span className="font-mono text-white">{match.status === 'live' ? `LIVE - ${match.timeElapsed}` : match.status.toUpperCase()}</span>
          </span>
        </div>

        {/* Scoring board */}
        <div className="p-8 flex items-center justify-between gap-6 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-zinc-50 via-white to-white">
          <div className="flex-1 text-center space-y-1">
            <h2 className="text-xl sm:text-2xl font-extrabold text-neutral-800">{match.homeTeam}</h2>
            <span className="text-xs text-zinc-400">Home Club</span>
          </div>

          <div className="flex items-center gap-5">
            <span className="text-4xl sm:text-5xl font-mono font-extrabold text-neutral-800">{match.homeScore}</span>
            <span className="text-lg font-semibold text-zinc-300">-</span>
            <span className="text-4xl sm:text-5xl font-mono font-extrabold text-neutral-800">{match.awayScore}</span>
          </div>

          <div className="flex-1 text-center space-y-1">
            <h2 className="text-xl sm:text-2xl font-extrabold text-neutral-800">{match.awayTeam}</h2>
            <span className="text-xs text-zinc-400">Away Club</span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {/* Live Commentary log */}
        <div className="md:col-span-2 space-y-4">
          <h3 className="text-sm font-bold uppercase font-mono text-zinc-500 tracking-wider flex items-center gap-1.5">
            <Activity className="w-4 h-4 text-brand-500" />
            <span>Timeline Commentary</span>
          </h3>

          {match.events && match.events.length > 0 ? (
            <div className="bg-white border border-zinc-200 rounded-xl overflow-hidden divide-y divide-zinc-100 shadow-sm">
              {[...match.events].reverse().map((e, idx) => (
                <div key={idx} className="p-4 flex gap-4 text-sm leading-relaxed items-start">
                  <span className="font-mono font-bold text-brand-600 bg-brand-50 rounded px-2 py-0.5 text-xs flex-shrink-0">
                    {e.time}
                  </span>
                  <div className="space-y-0.5">
                    <span className="font-semibold text-xs text-zinc-400 uppercase tracking-wider block">
                      {e.type}
                    </span>
                    <p className="text-neutral-800">{e.detail}</p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-white border border-zinc-200 rounded-xl p-8 text-center text-zinc-400">
              <p className="text-sm">No timeline commentary logged yet. Match updates will populate here.</p>
            </div>
          )}
        </div>

        {/* Pitch line-up / Match details card */}
        <div className="space-y-4">
          <h3 className="text-sm font-bold uppercase font-mono text-zinc-500 tracking-wider flex items-center gap-1.5">
            <Flag className="w-4 h-4 text-brand-500" />
            <span>Venue Directory</span>
          </h3>

          <div className="bg-white border border-zinc-200 rounded-xl p-5 space-y-4 shadow-sm text-sm">
            <div className="space-y-1">
              <span className="text-xs text-zinc-400">Stadium Location</span>
              <p className="font-semibold text-neutral-800">Orynt Arena, Main Field Pitch</p>
            </div>

            <div className="space-y-1">
              <span className="text-xs text-zinc-400">Access Elevators</span>
              <p className="font-medium text-neutral-700">North Stand, Section 101 Elevator 1</p>
            </div>

            <div className="space-y-1">
              <span className="text-xs text-zinc-400">Match Status Rules</span>
              <p className="text-xs text-zinc-500 leading-normal">
                Scoreboards are verified via referee data relays. WebSocket push streams automatically update metrics every 30 seconds.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
