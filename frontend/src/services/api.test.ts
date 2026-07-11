import { vi, describe, it, expect, beforeEach } from 'vitest'
import { api } from './api'
import { useAuthStore } from '../store'

describe('api service', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    useAuthStore.getState().clearAuth()
  })

  it('should make GET requests successfully', async () => {
    const mockData = { name: 'Orynt Arena' }
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => mockData,
    })

    const result = await api.get('/stadium')
    expect(globalThis.fetch).toHaveBeenCalledWith('/api/stadium', expect.objectContaining({ method: 'GET' }))
    expect(result).toEqual(mockData)
  })

  it('should inject Authorization header if token exists', async () => {
    useAuthStore.getState().setAuth(
      { id: 'u1', username: 'admin', role: 'admin', department: 'Ops', createdAt: '' },
      'my-jwt-token'
    )

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true }),
    })

    await api.get('/protected')

    expect(globalThis.fetch).toHaveBeenCalledWith(
      '/api/protected',
      expect.objectContaining({
        headers: expect.any(Headers),
      })
    )

    const calls = vi.mocked(globalThis.fetch).mock.calls
    const headers = calls[0][1]?.headers as Headers
    expect(headers.get('Authorization')).toBe('Bearer my-jwt-token')
  })

  it('should handle POST body and set Content-Type header', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ id: 'new_id' }),
    })

    await api.post('/lost-found', { itemName: 'keys' })

    const calls = vi.mocked(globalThis.fetch).mock.calls
    expect(calls[0][0]).toBe('/api/lost-found')
    expect(calls[0][1]?.method).toBe('POST')
    expect(calls[0][1]?.body).toBe(JSON.stringify({ itemName: 'keys' }))

    const headers = calls[0][1]?.headers as Headers
    expect(headers.get('Content-Type')).toBe('application/json')
  })

  it('should throw error on non-ok response', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 400,
      statusText: 'Bad Request',
      json: async () => ({ error: 'Invalid user fields' }),
    })

    await expect(api.post('/auth/login', {})).rejects.toThrow('Invalid user fields')
  })

  it('should return empty object on 204 status', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    })

    const result = await api.put('/tasks/1/resolve')
    expect(result).toEqual({})
  })
})
