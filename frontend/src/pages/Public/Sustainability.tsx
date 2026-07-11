import React, { useState, useEffect } from 'react'
import { api } from '../../services/api'
import { SustainabilityMetrics } from '../../types'
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts'
import { Sparkles, Zap, Droplets, Trash2, Leaf, BarChart2, Activity } from 'lucide-react'

const formatAIOutput = (text: string) => {
  if (!text) return null
  const lines = text.split('\n')
  return (
    <ul className="space-y-3.5">
      {lines.map((line, idx) => {
        let cleanLine = line.trim()
        if (!cleanLine) return null

        // Match bullet lists
        const isBullet = cleanLine.startsWith('-') || cleanLine.startsWith('*')
        if (isBullet) {
          cleanLine = cleanLine.substring(1).trim()
        }

        // Handle bold parsing
        const parts = cleanLine.split(/(\*\*[^*]+\*\*)/g)
        const renderedText = parts.map((part, pIdx) => {
          if (part.startsWith('**') && part.endsWith('**')) {
            return (
              <strong key={pIdx} className="font-bold text-neutral-900">
                {part.slice(2, -2)}
              </strong>
            )
          }
          return part
        })

        return (
          <li key={idx} className="flex items-start gap-3">
            <span className="flex-shrink-0 w-5 h-5 rounded-full bg-emerald-100 flex items-center justify-center mt-0.5 text-emerald-600 font-bold text-xs">
              ✓
            </span>
            <span className="text-sm text-neutral-700 leading-relaxed">{renderedText}</span>
          </li>
        )
      })}
    </ul>
  )
}

export const Sustainability: React.FC = () => {
  const [data, setData] = useState<SustainabilityMetrics | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'energy' | 'water' | 'waste'>('energy')

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        setLoading(true)
        const res = await api.get<SustainabilityMetrics>('/sustainability')
        setData(res)
      } catch (err: any) {
        console.error(err)
        setError(err.message || 'Failed to fetch sustainability telemetry.')
      } finally {
        setLoading(false)
      }
    }
    fetchMetrics()
  }, [])

  if (loading) {
    return (
      <div className="space-y-8 animate-pulse font-sans">
        <div className="h-40 bg-zinc-200 rounded-2xl" />
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="h-28 bg-zinc-200 rounded-xl" />
          <div className="h-28 bg-zinc-200 rounded-xl" />
          <div className="h-28 bg-zinc-200 rounded-xl" />
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 h-96 bg-zinc-200 rounded-xl" />
          <div className="h-96 bg-zinc-200 rounded-xl" />
        </div>
      </div>
    )
  }

  if (error || !data) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 rounded-xl p-6 text-center max-w-md mx-auto mt-12 font-sans">
        <p className="font-semibold">Error Loading Sustainability Data</p>
        <p className="text-sm mt-1 text-red-600">{error || 'Telemetry offline.'}</p>
        <button
          onClick={() => window.location.reload()}
          className="mt-4 px-4 py-2 bg-red-600 text-white rounded text-sm font-semibold hover:bg-red-700 transition"
        >
          Retry
        </button>
      </div>
    )
  }

  const { current, historical, recommendations } = data

  const tabConfigs = {
    energy: {
      label: 'Energy Usage',
      icon: Zap,
      dataKey: 'energyUsageKw',
      unit: ' kW',
      stroke: '#eab308',
      fill: 'rgba(234, 179, 8, 0.1)',
      desc: 'Real-time electrical consumption matching lighting, ventilation, and digital screens.'
    },
    water: {
      label: 'Water Consumption',
      icon: Droplets,
      dataKey: 'waterConsumptionL',
      unit: ' L',
      stroke: '#3b82f6',
      fill: 'rgba(59, 130, 246, 0.1)',
      desc: 'Halftime toilet flush flows, concession usage, and pitch maintenance irrigation.'
    },
    waste: {
      label: 'Waste Generated',
      icon: Trash2,
      dataKey: 'wasteGeneratedKg',
      unit: ' kg',
      stroke: '#10b981',
      fill: 'rgba(16, 185, 129, 0.1)',
      desc: 'Total visitor waste, divided by sorting audits (recyclables vs food compost).'
    }
  }

  const activeConfig = tabConfigs[activeTab]

  return (
    <div className="space-y-8 font-sans">
      {/* Page Header Banner */}
      <div className="bg-gradient-to-r from-emerald-900 to-zinc-900 text-white rounded-2xl p-8 shadow-lg relative overflow-hidden flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div className="relative z-10 space-y-2 max-w-2xl">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-emerald-800/40 text-emerald-300 text-xs font-mono font-bold uppercase tracking-wider">
            <Leaf className="w-3.5 h-3.5" />
            <span>Green Arena Telemetry</span>
          </div>
          <h1 className="text-3xl font-extrabold tracking-tight">Orynt Arena Sustainability</h1>
          <p className="text-zinc-300 text-sm leading-relaxed">
            Monitoring energy consumption, water conservation, and waste management metrics across stand zones in real time. Backed by Google Gemini eco-advisor recommendations.
          </p>
        </div>
        <div className="flex-shrink-0 relative z-10">
          <div className="bg-white/10 backdrop-blur-md rounded-xl p-4 border border-white/10 text-center min-w-[150px]">
            <span className="text-[10px] uppercase font-mono tracking-wider text-zinc-400 block mb-1">Live Attendance</span>
            <span className="text-3xl font-mono font-extrabold text-emerald-400">
              {current.attendance.toLocaleString()}
            </span>
            <span className="text-[10px] text-zinc-400 block mt-1">active telemetry</span>
          </div>
        </div>
        <div className="absolute inset-0 opacity-10 bg-[linear-gradient(to_right,#808080_1px,transparent_1px),linear-gradient(to_bottom,#808080_1px,transparent_1px)] bg-[size:24px_24px] pointer-events-none" />
      </div>

      {/* Main Metric Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Energy Card */}
        <div className="bg-white border border-zinc-200 rounded-2xl p-5 shadow-sm flex items-start gap-4 hover:shadow-md transition">
          <div className="w-12 h-12 rounded-xl bg-amber-50 text-amber-500 flex items-center justify-center flex-shrink-0">
            <Zap className="w-6 h-6 fill-current" />
          </div>
          <div className="space-y-1">
            <span className="text-xs font-medium text-zinc-500">Live Power Demand</span>
            <h3 className="text-2xl font-mono font-extrabold text-neutral-800">
              {Math.round(current.energyUsageKw).toLocaleString()} <span className="text-sm font-normal text-zinc-400">kW</span>
            </h3>
            <p className="text-xs text-zinc-400">Floodlights, concessions, corridors</p>
          </div>
        </div>

        {/* Water Card */}
        <div className="bg-white border border-zinc-200 rounded-2xl p-5 shadow-sm flex items-start gap-4 hover:shadow-md transition">
          <div className="w-12 h-12 rounded-xl bg-blue-50 text-blue-500 flex items-center justify-center flex-shrink-0">
            <Droplets className="w-6 h-6" />
          </div>
          <div className="space-y-1">
            <span className="text-xs font-medium text-zinc-500">Live Water Flow</span>
            <h3 className="text-2xl font-mono font-extrabold text-neutral-800">
              {Math.round(current.waterConsumptionL).toLocaleString()} <span className="text-sm font-normal text-zinc-400">L</span>
            </h3>
            <p className="text-xs text-zinc-400">Sanitation and field irrigation</p>
          </div>
        </div>

        {/* Waste Card */}
        <div className="bg-white border border-zinc-200 rounded-2xl p-5 shadow-sm flex items-start gap-4 hover:shadow-md transition">
          <div className="w-12 h-12 rounded-xl bg-emerald-50 text-emerald-500 flex items-center justify-center flex-shrink-0">
            <Trash2 className="w-6 h-6" />
          </div>
          <div className="space-y-1">
            <span className="text-xs font-medium text-zinc-500">Live Waste Generated</span>
            <h3 className="text-2xl font-mono font-extrabold text-neutral-800">
              {Math.round(current.wasteGeneratedKg).toLocaleString()} <span className="text-sm font-normal text-zinc-400">kg</span>
            </h3>
            <p className="text-xs text-zinc-400">Recyclables, food, non-recyclables</p>
          </div>
        </div>
      </div>

      {/* Analytics & Recommendation Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        
        {/* Recharts Timeline */}
        <div className="lg:col-span-2 bg-white border border-zinc-200 rounded-2xl p-6 shadow-sm space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <h2 className="text-lg font-bold text-neutral-800 flex items-center gap-2">
                <BarChart2 className="w-5 h-5 text-brand-600" />
                <span>Historical Telemetry Trends</span>
              </h2>
              <p className="text-xs text-zinc-500 mt-0.5">Time-series tracking across the current match day event.</p>
            </div>
            
            {/* Tab Toggles */}
            <div className="flex bg-zinc-100 rounded-lg p-1 text-xs font-semibold">
              {(Object.keys(tabConfigs) as Array<keyof typeof tabConfigs>).map((tab) => {
                const cfg = tabConfigs[tab]
                const Icon = cfg.icon
                return (
                  <button
                    key={tab}
                    onClick={() => setActiveTab(tab)}
                    className={`flex items-center gap-1 px-3 py-1.5 rounded transition ${
                      activeTab === tab
                        ? 'bg-white text-brand-600 shadow-sm'
                        : 'text-zinc-500 hover:text-neutral-800'
                    }`}
                  >
                    <Icon className="w-3.5 h-3.5" />
                    <span>{cfg.label}</span>
                  </button>
                )
              })}
            </div>
          </div>

          <div className="p-3 bg-zinc-50 rounded-xl text-xs text-zinc-600 border border-zinc-100 leading-relaxed">
            <span className="font-semibold text-brand-600">Metric Context:</span> {activeConfig.desc}
          </div>

          {/* Recharts Chart */}
          <div className="h-72 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={historical} margin={{ top: 10, right: 10, left: 10, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f4f4f5" />
                <XAxis dataKey="time" stroke="#71717a" fontSize={11} />
                <YAxis stroke="#71717a" fontSize={11} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#ffffff', borderRadius: '12px', borderColor: '#e4e4e7', fontSize: '12px' }}
                  labelClassName="font-bold text-zinc-800 font-sans"
                />
                <Legend wrapperStyle={{ fontSize: '12px', marginTop: '10px' }} />
                <Line
                  type="monotone"
                  name={activeConfig.label}
                  dataKey={activeConfig.dataKey}
                  stroke={activeConfig.stroke}
                  strokeWidth={3}
                  dot={{ r: 4, strokeWidth: 2 }}
                  activeDot={{ r: 6 }}
                />
                <Line
                  type="monotone"
                  name="Attendance"
                  dataKey="attendance"
                  stroke="#8b5cf6"
                  strokeWidth={1.5}
                  strokeDasharray="4 4"
                  dot={false}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* AI Recommendations Panel */}
        <div className="bg-white border border-zinc-200 rounded-2xl p-6 shadow-sm flex flex-col justify-between space-y-6">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-bold text-neutral-800 flex items-center gap-2">
                <Sparkles className="w-5 h-5 text-emerald-600" />
                <span>AI Eco-Advisor</span>
              </h2>
              <span className="flex items-center gap-1 px-2.5 py-0.5 rounded-full bg-emerald-50 text-[10px] font-bold text-emerald-600 uppercase border border-emerald-100">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-600 animate-pulse" />
                <span>Live recommendations</span>
              </span>
            </div>
            
            <p className="text-xs text-zinc-500 leading-relaxed">
              Google Gemini generates real-time operational directives to optimize stadium utilities based on current crowd capacities.
            </p>

            <div className="border-t border-zinc-100 pt-4">
              {formatAIOutput(recommendations)}
            </div>
          </div>

          <div className="bg-emerald-50/50 border border-emerald-100/80 rounded-xl p-4 flex gap-3">
            <div className="mt-0.5">
              <Activity className="w-4 h-4 text-emerald-600" />
            </div>
            <div className="space-y-1">
              <h4 className="text-xs font-bold text-emerald-800">Telemetry Accuracy</h4>
              <p className="text-[10px] text-emerald-600/90 leading-relaxed">
                Sustainability metrics are derived from IoT sensors calibrated against crowd stand turnstiles and concession checkout streams.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
