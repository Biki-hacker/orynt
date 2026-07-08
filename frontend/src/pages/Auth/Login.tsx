import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../store'
import { api } from '../../services/api'
import { TokenResponse } from '../../types'
import { KeyRound, ShieldAlert, Sparkles, UserCheck } from 'lucide-react'

export const Login: React.FC = () => {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const setAuth = useAuthStore((state) => state.setAuth)
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!username || !password) return

    setLoading(true)
    setError(null)

    try {
      const data = await api.post<TokenResponse>('/auth/login', { username, password })
      setAuth(data.user, data.accessToken)

      // Redirect based on role
      if (data.user.role === 'admin') {
        navigate('/admin')
      } else {
        navigate('/staff')
      }
    } catch (err: any) {
      setError(err.message || 'Login failed. Please verify credentials.')
    } finally {
      setLoading(false)
    }
  }

  // Quick select login helpers for hackathon testing
  const handleQuickLogin = async (userType: 'admin' | 'staff') => {
    const credentials = {
      admin: { u: 'admin', p: 'admin123' },
      staff: { u: 'volunteer1', p: 'staff123' }
    }
    const creds = credentials[userType]
    setUsername(creds.u)
    setPassword(creds.p)

    setLoading(true)
    setError(null)
    try {
      const data = await api.post<TokenResponse>('/auth/login', { username: creds.u, password: creds.p })
      setAuth(data.user, data.accessToken)
      if (data.user.role === 'admin') {
        navigate('/admin')
      } else {
        navigate('/staff')
      }
    } catch (err: any) {
      setError(err.message || 'Quick login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-md w-full mx-auto my-12 bg-white border border-zinc-200 rounded-2xl shadow-xl overflow-hidden font-sans">
      <div className="bg-zinc-900 text-white px-8 py-6 text-center">
        <h2 className="text-xl font-bold tracking-tight">Orynt Control Center</h2>
        <p className="text-xs text-zinc-400 mt-1">Authorized Operations Staff Login</p>
      </div>

      <div className="p-8 space-y-6">
        {error && (
          <div className="p-3 bg-red-50 border border-red-100 rounded-lg flex gap-2 text-xs text-red-700 items-start">
            <ShieldAlert className="w-4 h-4 flex-shrink-0 mt-0.5" />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-neutral-600 block">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="e.g. admin, volunteer1"
              required
              className="w-full bg-white border border-zinc-300 rounded px-3.5 py-2 text-sm focus:outline-none focus:border-brand-500 transition-colors"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-neutral-600 block">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              required
              className="w-full bg-white border border-zinc-300 rounded px-3.5 py-2 text-sm focus:outline-none focus:border-brand-500 transition-colors"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full flex items-center justify-center gap-1.5 bg-brand-600 hover:bg-brand-700 text-white font-bold text-sm py-2.5 rounded transition-all shadow-sm focus:outline-none focus:ring-2 focus:ring-brand-500 focus:ring-offset-2"
          >
            <KeyRound className="w-4 h-4" />
            <span>{loading ? 'Authenticating...' : 'Sign In'}</span>
          </button>
        </form>

        {/* Hackathon Quick Login Panel */}
        <div className="pt-6 border-t border-zinc-100 space-y-3">
          <span className="text-[10px] uppercase font-mono font-bold text-zinc-400 tracking-wider flex items-center gap-1">
            <Sparkles className="w-3.5 h-3.5 text-brand-500" />
            <span>Hackathon Quick Access Console</span>
          </span>
          <div className="grid grid-cols-2 gap-3">
            <button
              onClick={() => handleQuickLogin('admin')}
              className="flex items-center justify-center gap-1.5 py-2 px-3 border border-zinc-200 rounded text-xs font-semibold text-neutral-700 hover:bg-neutral-50 hover:border-zinc-300 transition-colors"
            >
              <UserCheck className="w-3.5 h-3.5 text-brand-600" />
              <span>Admin Role</span>
            </button>
            <button
              onClick={() => handleQuickLogin('staff')}
              className="flex items-center justify-center gap-1.5 py-2 px-3 border border-zinc-200 rounded text-xs font-semibold text-neutral-700 hover:bg-neutral-50 hover:border-zinc-300 transition-colors"
            >
              <UserCheck className="w-3.5 h-3.5 text-brand-600" />
              <span>Staff Role</span>
            </button>
          </div>
          <p className="text-[10px] text-zinc-400 text-center leading-normal mt-2">
            Use shortcuts above to immediately test different console panels (Admin Telemetry Overrides or Staff task workflows) without typing.
          </p>
        </div>
      </div>
    </div>
  )
}
