import { useNotificationStore } from '../store'

type WSCallback = (payload: any) => void;

class WebSocketService {
  private socket: WebSocket | null = null;
  private listeners: Map<string, Set<WSCallback>> = new Map();
  private reconnectTimeout: any = null;
  private reconnectDelay = 1000;
  private maxReconnectDelay = 30000;

  constructor() {
    // Lazy initialize connection
  }

  public connect() {
    if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    const loc = window.location
    const proto = loc.protocol === 'https:' ? 'wss:' : 'ws:'
    // Use VITE_WS_HOST from .env, or fall back to localhost:8080 in dev
    const wsHost = import.meta.env.VITE_WS_HOST || (loc.port === '3000' ? 'localhost:8080' : loc.host)
    const wsUrl = `${proto}//${wsHost}/api/ws`

    console.log(`[WS] Connecting to ${wsUrl}...`)
    this.socket = new WebSocket(wsUrl)

    this.socket.onopen = () => {
      console.log('[WS] Connected successfully.')
      this.reconnectDelay = 1000 // reset backoff
    }

    this.socket.onclose = (event) => {
      console.log('[WS] Connection closed:', event.reason)
      this.scheduleReconnect()
    }

    this.socket.onerror = (err) => {
      console.error('[WS] Connection error:', err)
      this.socket?.close()
    }

    this.socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        const type = data.type
        const payload = data.payload

        console.log(`[WS] Message received: ${type}`, payload)

        // Handle Global Notifications triggers
        if (type === 'alert') {
          useNotificationStore.getState().addNotification({
            title: `⚠️ Alert: ${payload.title}`,
            content: payload.content,
            type: 'emergency',
            severity: payload.severity
          })
        } else if (type === 'announcement') {
          useNotificationStore.getState().addNotification({
            title: `📢 Announcement: ${payload.title}`,
            content: payload.content,
            type: 'info',
            severity: 'info'
          })
        } else if (type === 'medical_request') {
          useNotificationStore.getState().addNotification({
            title: '🚨 Medical Assistance Requested',
            content: `Location: ${payload.location}. Description: ${payload.description}`,
            type: 'emergency',
            severity: 'critical'
          })
        }

        // Dispatch to registered components
        const callbacks = this.listeners.get(type)
        if (callbacks) {
          callbacks.forEach((cb) => cb(payload))
        }

        // Also trigger general broadcast listeners
        const generalCallbacks = this.listeners.get('*')
        if (generalCallbacks) {
          generalCallbacks.forEach((cb) => cb({ type, payload }))
        }
      } catch (err) {
        console.error('[WS] Failed to parse message body:', err)
      }
    }
  }

  private scheduleReconnect() {
    if (this.reconnectTimeout) return;
    this.reconnectTimeout = setTimeout(() => {
      this.reconnectTimeout = null
      this.connect()
      // Exponential backoff
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay)
    }, this.reconnectDelay)
  }

  public subscribe(type: string, callback: WSCallback) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set())
    }
    this.listeners.get(type)!.add(callback)

    // return unsubscribe function
    return () => {
      const callbacks = this.listeners.get(type)
      if (callbacks) {
        callbacks.delete(callback)
        if (callbacks.size === 0) {
          this.listeners.delete(type)
        }
      }
    }
  }

  public disconnect() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }
    if (this.socket) {
      this.socket.onclose = null // avoid trigger reconnect
      this.socket.close()
      this.socket = null
    }
  }
}

export const wsService = new WebSocketService()
