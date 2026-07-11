import { vi, describe, it, expect, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { AIChatbot } from './AIChatbot'

// Mock useAuthStore
vi.mock('../../store', () => ({
  useAuthStore: {
    getState: () => ({
      accessToken: 'test-token',
    }),
  },
}))

// Mock scrollIntoView since jsdom doesn't implement it
window.HTMLElement.prototype.scrollIntoView = vi.fn()

describe('AIChatbot Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('should render the floating chat button initially', () => {
    render(<AIChatbot />)
    const chatButton = screen.getByRole('button')
    expect(chatButton).toBeInTheDocument()
  })

  it('should open the chat panel when the button is clicked', () => {
    render(<AIChatbot />)
    const chatButton = screen.getByRole('button')
    fireEvent.click(chatButton)

    // Check if welcome message exists
    expect(
      screen.getByText(
        /Hi! I am your Orynt Smart Assistant. Ask me anything about live scores, venue navigation, facilities, transport delays, or parking space availabilities./
      )
    ).toBeInTheDocument()

    // Check if suggested prompts are displayed
    expect(screen.getByText('What is the live score?')).toBeInTheDocument()
    expect(screen.getByText('Find available parking')).toBeInTheDocument()
  })

  it('should close the chat panel when the close button is clicked', () => {
    render(<AIChatbot />)
    // Open
    fireEvent.click(screen.getByRole('button'))
    expect(screen.getByText(/Orynt Smart Assistant/)).toBeInTheDocument()

    // Close button is the first button inside the open header
    // Let's grab all buttons and click the one with the close action
    const closeBtn = screen.getAllByRole('button')[0]
    fireEvent.click(closeBtn)

    expect(screen.queryByText(/Orynt Smart Assistant/)).not.toBeInTheDocument()
  })

  it('should submit a message when clicking a suggested prompt', async () => {
    // Mock fetch for SSE stream reader
    const mockReader = {
      read: vi.fn()
        .mockResolvedValueOnce({
          value: new TextEncoder().encode(
            'data: {"response": "Parking Lot A has 100 free spots.", "sources": ["Smart Parking Sensors"]}\n'
          ),
          done: false,
        })
        .mockResolvedValueOnce({ done: true }),
    }
    const mockStream = {
      getReader: () => mockReader,
    }
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      body: mockStream,
    })

    render(<AIChatbot />)
    fireEvent.click(screen.getByRole('button'))

    // Click on a suggested prompt
    const promptBtn = screen.getByText('Find available parking')
    fireEvent.click(promptBtn)

    // Verify user message was added (one in the prompt button, one in the chat message)
    expect(screen.getAllByText('Find available parking').length).toBe(2)

    // Verify fetch was called
    expect(globalThis.fetch).toHaveBeenCalledWith('/api/ai/chat', expect.any(Object))

    // Wait for the mock streaming response to be rendered
    await waitFor(() => {
      expect(screen.getByText(/Parking Lot A has 100 free spots/)).toBeInTheDocument()
    })
  })
})
