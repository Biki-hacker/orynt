import React, { useState, useRef, useEffect } from 'react'
import { MessageSquare, X, Send, Sparkles } from 'lucide-react'
import { useAuthStore } from '../../store'

interface ChatMessage {
  id: string;
  sender: 'user' | 'ai';
  text: string;
  timestamp: Date;
  sources?: string[];
}

const renderFormattedText = (text: string) => {
  if (!text) return null
  const parts = text.split(/(\*\*[^*]+\*\*|\*[^*]+\*)/g)
  return parts.map((part, index) => {
    if (part.startsWith('**') && part.endsWith('**')) {
      return (
        <strong key={index} className="font-bold text-zinc-950">
          {part.slice(2, -2)}
        </strong>
      )
    } else if (part.startsWith('*') && part.endsWith('*')) {
      return (
        <em key={index} className="italic font-medium text-zinc-800">
          {part.slice(1, -1)}
        </em>
      )
    }
    return part
  })
}

export const AIChatbot: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false)
  const [messages, setMessages] = useState<ChatMessage[]>([
    {
      id: 'welcome',
      sender: 'ai',
      text: 'Hi! I am your Orynt Smart Assistant. Ask me anything about live scores, venue navigation, facilities, transport delays, or parking space availabilities.',
      timestamp: new Date()
    }
  ])
  const [inputText, setInputText] = useState('')
  const [isTyping, setIsTyping] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const suggestedPrompts = [
    'What is the live score?',
    'Find available parking',
    'Where is the nearest first aid?',
    'Show transit delays'
  ]

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    if (isOpen) {
      scrollToBottom()
    }
  }, [messages, isOpen])

  const handleSend = async (text: string) => {
    if (!text.trim()) return

    const userMsg: ChatMessage = {
      id: Math.random().toString(36).substring(7),
      sender: 'user',
      text: text,
      timestamp: new Date()
    }

    // Build history of the last 10 messages before adding the new message
    const history = messages
      .filter((m) => m.id !== 'welcome')
      .slice(-10)
      .map((m) => ({
        sender: m.sender,
        text: m.text
      }))

    setMessages((prev) => [...prev, userMsg])
    setInputText('')
    setIsTyping(true)

    const token = useAuthStore.getState().accessToken
    const headers = new Headers()
    headers.set('Content-Type', 'application/json')
    if (token) {
      headers.set('Authorization', `Bearer ${token}`)
    }

    try {
      const response = await fetch('/api/ai/chat', {
        method: 'POST',
        headers,
        body: JSON.stringify({
          message: text,
          history
        })
      })

      if (!response.ok) {
        throw new Error('Failed to start chat stream')
      }

      const reader = response.body?.getReader()
      if (!reader) {
        throw new Error('No stream reader available')
      }

      const decoder = new TextDecoder()
      let done = false
      let aiResponseText = ''
      let aiSources: string[] = []

      // Create a placeholder message in chat
      const aiMsgId = Math.random().toString(36).substring(7)
      const initialAiMsg: ChatMessage = {
        id: aiMsgId,
        sender: 'ai',
        text: '',
        timestamp: new Date(),
        sources: []
      }

      setMessages((prev) => [...prev, initialAiMsg])
      setIsTyping(false) // turn off typing animation once stream starts rendering

      let buffer = ''

      while (!done) {
        const { value, done: readerDone } = await reader.read()
        done = readerDone
        if (value) {
          const chunkStr = decoder.decode(value, { stream: true })
          buffer += chunkStr

          const lines = buffer.split('\n')
          buffer = lines.pop() || '' // keep partial line in buffer

          for (const line of lines) {
            const trimmed = line.trim()
            if (trimmed.startsWith('data:')) {
              const dataStr = trimmed.slice(5).trim()
              try {
                const parsed = JSON.parse(dataStr)
                if (parsed.response) {
                  aiResponseText += parsed.response
                }
                if (parsed.sources && parsed.sources.length > 0) {
                  aiSources = [...new Set([...aiSources, ...parsed.sources])]
                }
                // Update placeholder in real-time
                setMessages((prev) =>
                  prev.map((msg) =>
                    msg.id === aiMsgId
                      ? {
                          ...msg,
                          text: aiResponseText,
                          sources: aiSources.length > 0 ? aiSources : undefined
                        }
                      : msg
                  )
                )
              } catch (e) {
                // Ignore JSON parsing errors for incomplete chunks
              }
            }
          }
        }
      }
    } catch (err: any) {
      const errorMsg: ChatMessage = {
        id: Math.random().toString(36).substring(7),
        sender: 'ai',
        text: `Error: ${err.message || 'Unable to connect to AI server. Please try again.'}`,
        timestamp: new Date()
      }
      setMessages((prev) => [...prev, errorMsg])
    } finally {
      setIsTyping(false)
    }
  }

  return (
    <div className="fixed bottom-6 right-6 z-40 font-sans">
      {/* Floating Toggle Button */}
      {!isOpen && (
        <button
          onClick={() => setIsOpen(true)}
          aria-label="Open AI Chatbot"
          className="w-14 h-14 rounded-full bg-brand-600 hover:bg-brand-700 text-white flex items-center justify-center shadow-lg hover:scale-105 transition-all duration-200 focus:outline-none focus:ring-4 focus:ring-brand-200"
        >
          <MessageSquare className="w-6 h-6" />
        </button>
      )}

      {/* Chat Window */}
      {isOpen && (
        <div className="w-96 h-[500px] bg-white border border-zinc-200 rounded-xl shadow-2xl flex flex-col overflow-hidden animate-in slide-in-from-bottom-5 duration-200">
          {/* Header */}
          <div className="bg-zinc-900 text-white px-5 py-4 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="w-6 h-6 rounded bg-brand-500 flex items-center justify-center">
                <Sparkles className="w-3.5 h-3.5 text-white" />
              </div>
              <div>
                <h3 className="font-bold text-sm leading-none">Orynt AI Assistant</h3>
                <span className="text-[10px] text-zinc-400 font-medium">Telemetry-Grounded RAG</span>
              </div>
            </div>
            <button onClick={() => setIsOpen(false)} aria-label="Close AI Chatbot" className="text-zinc-400 hover:text-white transition-colors">
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Messages Area */}
          <div className="flex-1 p-4 overflow-y-auto space-y-3 bg-zinc-50">
            {messages.map((m) => {
              const isUser = m.sender === 'user'
              return (
                <div key={m.id} className={`flex flex-col ${isUser ? 'items-end' : 'items-start'}`}>
                  <div
                    className={`max-w-[85%] rounded-lg px-4 py-2.5 text-sm leading-relaxed shadow-sm ${
                      isUser
                        ? 'bg-brand-600 text-white rounded-br-none'
                        : 'bg-white border border-zinc-200 text-neutral-800 rounded-bl-none'
                    }`}
                  >
                    <p className="whitespace-pre-wrap">{renderFormattedText(m.text)}</p>

                    {/* Source Attribution list */}
                    {!isUser && m.sources && m.sources.length > 0 && (
                      <div className="mt-2.5 pt-2.5 border-t border-zinc-100 flex flex-wrap gap-1 items-center">
                        <span className="text-[9px] text-zinc-400 font-semibold uppercase mr-1">Sources:</span>
                        {m.sources.map((src, idx) => (
                          <span
                            key={idx}
                            className="bg-zinc-100 border border-zinc-200 text-zinc-500 rounded px-1.5 py-0.5 text-[9px] font-medium"
                          >
                            {src}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                  <span className="text-[10px] text-zinc-400 mt-1 px-1 font-mono">
                    {m.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                  </span>
                </div>
              )
            })}

            {isTyping && (
              <div className="flex items-center gap-1 bg-white border border-zinc-200 rounded-lg px-4 py-3 max-w-[70px] shadow-sm">
                <span className="w-1.5 h-1.5 rounded-full bg-zinc-400 animate-bounce" style={{ animationDelay: '0ms' }} />
                <span className="w-1.5 h-1.5 rounded-full bg-zinc-400 animate-bounce" style={{ animationDelay: '150ms' }} />
                <span className="w-1.5 h-1.5 rounded-full bg-zinc-400 animate-bounce" style={{ animationDelay: '300ms' }} />
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* Quick suggestions */}
          <div className="px-4 py-2 border-t border-zinc-100 flex gap-1.5 overflow-x-auto whitespace-nowrap bg-white scrollbar-none">
            {suggestedPrompts.map((p, idx) => (
              <button
                key={idx}
                onClick={() => handleSend(p)}
                className="text-[11px] font-semibold text-brand-600 bg-brand-50 hover:bg-brand-100 border border-brand-100 rounded-full px-3 py-1 transition-all"
              >
                {p}
              </button>
            ))}
          </div>

          {/* Input Bar */}
          <form
            onSubmit={(e) => {
              e.preventDefault()
              handleSend(inputText)
            }}
            className="p-3 border-t border-zinc-200 bg-white flex gap-2"
          >
            <input
              type="text"
              value={inputText}
              onChange={(e) => setInputText(e.target.value)}
              placeholder="Ask Orynt Stadium AI..."
              className="flex-1 bg-zinc-100 hover:bg-zinc-200/50 focus:bg-white border border-zinc-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-brand-500 transition-colors"
            />
            <button
              type="submit"
              disabled={!inputText.trim()}
              aria-label="Send message"
              className="w-9 h-9 rounded-lg bg-brand-600 hover:bg-brand-700 text-white flex items-center justify-center shadow-sm disabled:bg-zinc-100 disabled:text-zinc-300 transition-colors"
            >
              <Send className="w-4 h-4" />
            </button>
          </form>
        </div>
      )}
    </div>
  )
}
