import { describe, it, expect, beforeEach } from 'vitest'
import { useAuthStore, useNotificationStore } from './index'
import { User } from '../types'

describe('useAuthStore', () => {
  beforeEach(() => {
    localStorage.clear()
    useAuthStore.getState().clearAuth()
  })

  it('should have initial null state', () => {
    const state = useAuthStore.getState()
    expect(state.user).toBeNull()
    expect(state.accessToken).toBeNull()
    expect(state.isAuthenticated).toBe(false)
  })

  it('should setAuth and save to localStorage', () => {
    const mockUser: User = {
      id: 'u_123',
      username: 'testuser',
      role: 'volunteer',
      department: 'Cleaning',
      createdAt: '2026-07-07T21:40:00Z',
    }
    const mockToken = 'mock-jwt-token-xyz'

    useAuthStore.getState().setAuth(mockUser, mockToken)

    const state = useAuthStore.getState()
    expect(state.user).toEqual(mockUser)
    expect(state.accessToken).toBe(mockToken)
    expect(state.isAuthenticated).toBe(true)

    expect(localStorage.getItem('orynt_user')).toBe(JSON.stringify(mockUser))
    expect(localStorage.getItem('orynt_token')).toBe(mockToken)
  })

  it('should clearAuth and remove from localStorage', () => {
    const mockUser: User = {
      id: 'u_123',
      username: 'testuser',
      role: 'volunteer',
      department: 'Cleaning',
      createdAt: '2026-07-07T21:40:00Z',
    }
    const mockToken = 'mock-jwt-token-xyz'

    useAuthStore.getState().setAuth(mockUser, mockToken)
    useAuthStore.getState().clearAuth()

    const state = useAuthStore.getState()
    expect(state.user).toBeNull()
    expect(state.accessToken).toBeNull()
    expect(state.isAuthenticated).toBe(false)

    expect(localStorage.getItem('orynt_user')).toBeNull()
    expect(localStorage.getItem('orynt_token')).toBeNull()
  })
})

describe('useNotificationStore', () => {
  beforeEach(() => {
    useNotificationStore.getState().clearAll()
  })

  it('should start with an empty notification list', () => {
    expect(useNotificationStore.getState().notifications).toEqual([])
  })

  it('should add a notification and generate an ID', () => {
    useNotificationStore.getState().addNotification({
      title: 'Alert title',
      content: 'Alert content detail',
      type: 'crowd',
      severity: 'high',
    })

    const notifications = useNotificationStore.getState().notifications
    expect(notifications.length).toBe(1)
    expect(notifications[0].title).toBe('Alert title')
    expect(notifications[0].id).toBeDefined()
    expect(notifications[0].createdAt).toBeInstanceOf(Date)
  })

  it('should limit the notifications list to a maximum of 8', () => {
    const store = useNotificationStore.getState()
    for (let i = 0; i < 12; i++) {
      store.addNotification({
        title: `Alert ${i}`,
        content: `Content ${i}`,
        type: 'info',
        severity: 'info',
      })
    }

    const notifications = useNotificationStore.getState().notifications
    expect(notifications.length).toBe(8)
    expect(notifications[0].title).toBe('Alert 11') // Last added
  })

  it('should remove a notification by ID', () => {
    const store = useNotificationStore.getState()
    store.addNotification({
      title: 'Alert to remove',
      content: 'Content',
      type: 'weather',
      severity: 'critical',
    })

    const initialNotifications = useNotificationStore.getState().notifications
    const idToRemove = initialNotifications[0].id

    store.removeNotification(idToRemove)

    const finalNotifications = useNotificationStore.getState().notifications
    expect(finalNotifications.length).toBe(0)
  })
})
