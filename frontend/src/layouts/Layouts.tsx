import React, { useState, useEffect } from 'react'
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useAuthStore, useNotificationStore } from '../store'
import { AIChatbot } from '../components/Chat/AIChatbot'
import { wsService } from '../services/websocket'
import { api } from '../services/api'
import { Alert } from '../types'
import { User as UserIcon, LogOut, Menu, X, ShieldAlert, CheckCircle2, Info } from 'lucide-react'

// Toast Notification ToastContainer
const ToastContainer: React.FC = () => {
  const { notifications, removeNotification } = useNotificationStore()

  return (
    <div className="fixed bottom-5 left-5 z-50 flex flex-col gap-2 max-w-md w-full">
      {notifications.map((n) => (
        <div
          key={n.id}
          className={`flex items-start gap-3 p-4 bg-white border rounded-lg shadow-lg animate-in slide-in-from-left-5 duration-200 ${
            n.severity === 'critical'
              ? 'border-l-4 border-l-red-600 border-zinc-200'
              : n.severity === 'high'
              ? 'border-l-4 border-l-amber-500 border-zinc-200'
              : 'border-l-4 border-l-blue-500 border-zinc-200'
          }`}
        >
          <div className="mt-0.5">
            {n.severity === 'critical' ? (
              <ShieldAlert className="w-5 h-5 text-red-600" />
            ) : n.severity === 'high' ? (
              <ShieldAlert className="w-5 h-5 text-amber-500" />
            ) : (
              <Info className="w-5 h-5 text-blue-500" />
            )}
          </div>
          <div className="flex-1">
            <h4 className="text-sm font-semibold text-neutral-800">{n.title}</h4>
            <p className="text-xs text-zinc-500 mt-1 leading-relaxed">{n.content}</p>
          </div>
          <button
            onClick={() => removeNotification(n.id)}
            aria-label="Close notification"
            className="w-9 h-9 flex items-center justify-center text-zinc-400 hover:text-zinc-600 hover:bg-zinc-100 rounded-full transition-colors flex-shrink-0"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      ))}
    </div>
  )
}

// PUBLIC LAYOUT
export const PublicLayout: React.FC = () => {
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const { user, isAuthenticated, clearAuth } = useAuthStore()
  const [activeAlerts, setActiveAlerts] = useState<Alert[]>([])
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  useEffect(() => {
    // Connect WebSockets
    wsService.connect()

    // Fetch initial active emergency alerts
    const loadAlerts = async () => {
      try {
        const data = await api.get<Alert[]>('/alerts?active=true')
        setActiveAlerts(data)
      } catch (err) {
        console.error(err)
      }
    }
    loadAlerts()

    // Subscribe to new real-time alerts
    const unsub = wsService.subscribe('alert', (alert: Alert) => {
      if (alert.active) {
        setActiveAlerts((prev) => [alert, ...prev])
      }
    })

    const unsubResolve = wsService.subscribe('alert_resolved', (payload: { id: string }) => {
      setActiveAlerts((prev) => prev.filter((a) => a.id !== payload.id))
    })

    return () => {
      unsub()
      unsubResolve()
    }
  }, [])

  const navLinks = [
    { label: 'Live Tournament', path: '/' },
    { label: 'Stadium Map', path: '/map' },
    { label: 'Transport & Parking', path: '/transport' },
    { label: 'Food & FAQ', path: '/food-faq' },
    { label: 'Sustainability', path: '/sustainability' }
  ]

  return (
    <div className="min-h-screen flex flex-col bg-neutral-50">
      {/* Emergency Alert Banner */}
      {activeAlerts.length > 0 && (
        <div className="bg-red-600 text-white py-2.5 px-4 text-sm flex items-center justify-between font-medium animate-pulse shadow-md z-45">
          <div className="flex items-center gap-2 max-w-4xl">
            <ShieldAlert className="w-5 h-5 flex-shrink-0" />
            <span>
              <strong>EMERGENCY:</strong> {activeAlerts[0].title} — {activeAlerts[0].content}
            </span>
          </div>
          <button
            onClick={() => setActiveAlerts((prev) => prev.slice(1))}
            aria-label="Dismiss emergency alert"
            className="text-red-100 hover:text-white hover:bg-red-700/50 w-9 h-9 flex items-center justify-center rounded-full transition-colors flex-shrink-0"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Main Header */}
      <header className="bg-white border-b border-zinc-200 sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex-1 flex items-center">
            <Link to="/" className="flex items-center gap-2">
              <div className="w-8 h-8 rounded bg-brand-600 flex items-center justify-center text-white font-bold text-lg shadow-sm">
                O
              </div>
              <span className="font-bold text-xl tracking-tight text-neutral-800">Orynt</span>
            </Link>
          </div>

          <nav className="hidden md:flex items-center gap-6">
            {navLinks.map((link) => (
              <Link
                key={link.path}
                to={link.path}
                className={`text-sm font-medium transition-colors ${
                  pathname === link.path ? 'text-brand-600' : 'text-zinc-500 hover:text-neutral-800'
                }`}
              >
                {link.label}
              </Link>
            ))}
          </nav>

          <div className="flex-1 flex items-center justify-end gap-4">
            {isAuthenticated && user ? (
              <div className="flex items-center gap-3">
                <span className="text-xs font-semibold px-2.5 py-1 bg-zinc-100 rounded-full text-zinc-700 uppercase">
                  {user.role}
                </span>
                <Link
                  to={user.role === 'admin' ? '/admin' : '/staff'}
                  className="text-sm font-medium hover:underline text-neutral-700"
                >
                  Console
                </Link>
                <button
                  onClick={() => {
                    clearAuth()
                    navigate('/')
                  }}
                  aria-label="Log out"
                  className="text-zinc-400 hover:text-zinc-600 w-9 h-9 flex items-center justify-center hover:bg-zinc-100 rounded-full transition-colors"
                >
                  <LogOut className="w-4 h-4" />
                </button>
              </div>
            ) : (
              <Link
                to="/login"
                className="text-sm font-medium text-brand-600 hover:text-brand-700 border border-brand-500 hover:bg-brand-50 px-3.5 py-1.5 rounded transition-all duration-150"
              >
                Staff Access
              </Link>
            )}

            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              aria-label="Toggle navigation menu"
              className="md:hidden text-zinc-500 hover:text-neutral-800 w-10 h-10 flex items-center justify-center hover:bg-zinc-100 rounded-lg transition-colors"
            >
              {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>
        </div>

        {/* Mobile Navigation Drawer */}
        {mobileMenuOpen && (
          <div className="md:hidden bg-white border-b border-zinc-200 px-4 py-4 flex flex-col gap-3">
            {navLinks.map((link) => (
              <Link
                key={link.path}
                to={link.path}
                onClick={() => setMobileMenuOpen(false)}
                className={`text-sm font-medium py-1.5 ${
                  pathname === link.path ? 'text-brand-600' : 'text-zinc-500 hover:text-neutral-800'
                }`}
              >
                {link.label}
              </Link>
            ))}
          </div>
        )}
      </header>

      {/* Main Content */}
      <main className="flex-1 max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>

      {/* Floating Chatbot */}
      <AIChatbot />

      {/* Toast Notification Container */}
      <ToastContainer />
    </div>
  )
}

// DASHBOARD LAYOUT (ADMIN/STAFF CONSOLES)
export const DashboardLayout: React.FC = () => {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const { user, clearAuth } = useAuthStore()

  useEffect(() => {
    // If not logged in, redirect to login
    if (!useAuthStore.getState().isAuthenticated) {
      navigate('/login')
      return
    }
    // Connect WebSockets for real-time console telemetry
    wsService.connect()
  }, [navigate])

  if (!user) return null

  const isAdmin = user.role === 'admin'
  const navLinks = isAdmin
    ? [
        { label: 'Operations Analytics', path: '/admin' },
        { label: 'Task Coordination', path: '/staff' } // Admins can assign tasks too
      ]
    : [
        { label: 'My Action Board', path: '/staff' }
      ]

  return (
    <div className="min-h-screen flex bg-neutral-100 font-sans">
      {/* Sidebar navigation */}
      <aside className="w-64 bg-zinc-900 text-zinc-300 flex flex-col border-r border-zinc-800">
        <div className="h-16 flex items-center px-6 border-b border-zinc-800">
          <Link to="/" className="flex items-center gap-2">
            <div className="w-8 h-8 rounded bg-brand-600 flex items-center justify-center text-white font-bold text-lg">
              O
            </div>
            <span className="font-bold text-xl text-white tracking-tight">Orynt Hub</span>
          </Link>
        </div>

        <div className="px-4 py-4 flex items-center gap-3 border-b border-zinc-800 bg-zinc-950/20">
          <div className="w-10 h-10 rounded-full bg-zinc-700 flex items-center justify-center text-white">
            <UserIcon className="w-5 h-5" />
          </div>
          <div className="overflow-hidden">
            <h4 className="text-sm font-semibold text-white truncate">{user.username}</h4>
            <span className="text-xs text-zinc-500 uppercase font-mono tracking-wider">
              {user.role} | {user.department || 'OPS'}
            </span>
          </div>
        </div>

        <nav className="flex-1 px-3 py-6 flex flex-col gap-1.5">
          {navLinks.map((link) => (
            <Link
              key={link.path}
              to={link.path}
              className={`flex items-center px-4 py-2.5 rounded text-sm font-medium transition-all duration-150 ${
                pathname === link.path
                  ? 'bg-zinc-800 text-white border-l-4 border-brand-500 pl-3'
                  : 'hover:bg-zinc-800/50 text-zinc-400 hover:text-zinc-200'
              }`}
            >
              {link.label}
            </Link>
          ))}
        </nav>

        <div className="p-4 border-t border-zinc-800">
          <button
            onClick={() => {
              clearAuth()
              navigate('/')
            }}
            className="w-full flex items-center justify-center gap-2 py-2 px-3 rounded bg-zinc-800 hover:bg-zinc-700 text-sm font-semibold text-white transition-colors"
          >
            <LogOut className="w-4 h-4" />
            <span>Sign Out</span>
          </button>
        </div>
      </aside>

      {/* Main Wrapper */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <header className="h-16 bg-white border-b border-zinc-200 flex items-center justify-between px-8 shadow-sm">
          <h2 className="text-lg font-bold text-neutral-800">
            {isAdmin ? 'Administration Operations' : 'Staff Command'}
          </h2>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-1.5 text-zinc-500">
              <CheckCircle2 className="w-4 h-4 text-emerald-500" />
              <span className="text-xs font-medium">IoT Stream Connected</span>
            </div>
            <Link
              to="/"
              className="text-xs font-semibold text-brand-600 hover:text-brand-700 border border-brand-500 rounded px-2.5 py-1"
            >
              Public Dashboard
            </Link>
          </div>
        </header>

        <main className="flex-1 overflow-y-auto p-8">
          <Outlet />
        </main>
      </div>

      <AIChatbot />
      <ToastContainer />
    </div>
  )
}
