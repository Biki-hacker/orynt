import React, { useState, useEffect } from 'react'
import { FoodStall } from '../../types'
import { api } from '../../services/api'
import { wsService } from '../../services/websocket'
import { Badge } from '../../components/ui'
import { Coffee, HelpCircle, Clock, MapPin, ChevronDown, ChevronUp } from 'lucide-react'

export const FoodFAQ: React.FC = () => {
  const [stalls, setStalls] = useState<FoodStall[]>([])
  const [loading, setLoading] = useState(true)
  const [openFAQ, setOpenFAQ] = useState<number | null>(null)

  useEffect(() => {
    const loadStalls = async () => {
      try {
        const data = await api.get<FoodStall[]>('/food-stalls')
        setStalls(data)
      } catch (err) {
        console.error(err)
      } finally {
        setLoading(false)
      }
    }
    loadStalls()

    // Subscribe to food wait time updates
    const unsub = wsService.subscribe('food_stall', (stall: FoodStall) => {
      setStalls((prev) => prev.map((item) => (item.id === stall.id ? stall : item)))
    })

    return () => unsub()
  }, [])

  const faqs = [
    {
      q: 'How do I access wheelchair assistance in the stands?',
      a: 'Orynt Arena features dedicated priority seating. Elevator 1 (North Stand) and Ramp B (South Gate) provide fully step-free pathways. You can request escort helpers by asking any volunteer in a blue vest, or querying our AI chatbot.'
    },
    {
      q: 'Where do I claim lost keys, baggages, or electronic devices?',
      a: "All items found within the stadium grounds are cataloged in our live operations database. Please check the 'Lost & Found' log on this portal, or report a lost item to the customer assistance desk at Gate 1."
    },
    {
      q: 'How are stadium crowd densities calculated?',
      a: 'Our smart stadium stands are equipped with optical entry turnstile sensors and IoT density cameras that calculate the ratio of active visitors to the stand maximum capacity in real-time.'
    },
    {
      q: 'What should I do in case of a stadium evacuation warning?',
      a: 'In case of a critical alarm, an emergency notification banner will flash red on this app. Remain calm, follow green arrow indicators to Exit Gates A or B, and obey directions given by security personnel.'
    }
  ]

  if (loading) {
    return <div className="animate-pulse h-96 bg-zinc-200 rounded-xl" />
  }

  return (
    <div className="space-y-8 font-sans">
      <div className="space-y-1">
        <span className="text-xs uppercase font-mono font-bold text-brand-600 tracking-wider">
          Orynt Services Directory
        </span>
        <h1 className="text-2xl font-extrabold tracking-tight text-neutral-800">Food Stalls & FAQ</h1>
        <p className="text-zinc-500 text-sm max-w-2xl leading-relaxed">
          Check menu availabilities and queue wait times across our stadium dining locations. Read through the FAQ for general venue policies.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Food Stalls Column */}
        <div className="lg:col-span-2 space-y-4">
          <h2 className="text-lg font-bold text-neutral-800 flex items-center gap-2">
            <Coffee className="w-5 h-5 text-brand-500" />
            <span>Food & Beverage Stalls</span>
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {stalls.map((s) => (
              <div key={s.id} className="bg-white border border-zinc-200 rounded-xl p-5 shadow-sm space-y-3 flex flex-col justify-between">
                <div>
                  <div className="flex justify-between items-start">
                    <h3 className="font-bold text-neutral-800 text-sm">{s.name}</h3>
                    <Badge variant={s.status === 'open' ? 'success' : 'error'}>
                      {s.status.toUpperCase()}
                    </Badge>
                  </div>
                  <span className="text-xs text-zinc-400 font-medium block mt-0.5">{s.foodType}</span>

                  <div className="flex items-center gap-4 mt-3 py-2 border-y border-zinc-50">
                    <div className="flex items-center gap-1 text-zinc-500 text-xs">
                      <Clock className="w-3.5 h-3.5" />
                      <span>{s.status === 'open' ? `${s.waitTimeMinutes}M Queue` : 'Closed'}</span>
                    </div>
                    <div className="flex items-center gap-1 text-zinc-500 text-xs">
                      <MapPin className="w-3.5 h-3.5" />
                      <span>{s.location}</span>
                    </div>
                  </div>
                </div>

                <div className="mt-4">
                  <span className="text-[10px] text-zinc-400 font-semibold uppercase block tracking-wider">
                    Popular Items
                  </span>
                  <div className="flex flex-wrap gap-1 mt-1.5">
                    {s.menu.map((item, idx) => (
                      <span
                        key={idx}
                        className="text-[10px] bg-zinc-50 text-zinc-600 border border-zinc-200 rounded px-2 py-0.5 font-medium"
                      >
                        {item}
                      </span>
                    ))}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* FAQs Column */}
        <div className="space-y-4">
          <h2 className="text-lg font-bold text-neutral-800 flex items-center gap-2">
            <HelpCircle className="w-5 h-5 text-brand-500" />
            <span>General Venue FAQ</span>
          </h2>

          <div className="space-y-3 bg-white border border-zinc-200 rounded-xl p-5 shadow-sm">
            {faqs.map((faq, idx) => {
              const isOpen = openFAQ === idx
              return (
                <div key={idx} className="border-b border-zinc-100 last:border-b-0 pb-3 last:pb-0 pt-3 first:pt-0">
                  <button
                    onClick={() => setOpenFAQ(isOpen ? null : idx)}
                    className="w-full flex items-center justify-between text-left font-bold text-neutral-800 text-xs hover:text-brand-600 transition-colors"
                  >
                    <span>{faq.q}</span>
                    {isOpen ? <ChevronUp className="w-3.5 h-3.5 text-zinc-400" /> : <ChevronDown className="w-3.5 h-3.5 text-zinc-400" />}
                  </button>
                  {isOpen && (
                    <p className="text-zinc-500 text-xs leading-relaxed mt-2 pl-1 animate-in fade-in slide-in-from-top-2 duration-150">
                      {faq.a}
                    </p>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}
