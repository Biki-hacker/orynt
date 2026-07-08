import React from 'react'
import { StadiumMap } from '../../components/Map/StadiumMap'
import { Compass, PhoneCall } from 'lucide-react'

export const VenueMapPage: React.FC = () => {
  return (
    <div className="space-y-6 font-sans">
      <div className="space-y-1.5">
        <span className="text-xs uppercase font-mono font-bold text-brand-600 tracking-wider">
          Orynt Navigation Portal
        </span>
        <h1 className="text-2xl font-extrabold tracking-tight text-neutral-800">Interactive Venue Directory</h1>
        <p className="text-zinc-500 text-sm max-w-2xl leading-relaxed">
          Monitor real-time crowd standings, identify nearby facilities (food, washrooms, medical response zones), and find accessible pathways directly on the stadium vector blueprint.
        </p>
      </div>

      {/* Stadium Map */}
      <StadiumMap />

      {/* Auxiliary Help Directory */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-8">
        <div className="bg-white border border-zinc-200 rounded-xl p-5 shadow-sm space-y-2">
          <h4 className="font-bold text-sm text-neutral-800 flex items-center gap-1.5">
            <Compass className="w-4 h-4 text-brand-500" />
            <span>Map Indicators</span>
          </h4>
          <p className="text-xs text-zinc-500 leading-normal">
            Stands are color-coded based on active congestion sensors: Green (low density), Yellow (medium stand occupancy), and Red (high crowd congestion).
          </p>
        </div>

        <div className="bg-white border border-zinc-200 rounded-xl p-5 shadow-sm space-y-2">
          <h4 className="font-bold text-sm text-neutral-800 flex items-center gap-1.5">
            <Accessibility className="w-4 h-4 text-brand-500" />
            <span>Access Information</span>
          </h4>
          <p className="text-xs text-zinc-500 leading-normal">
            Elevator 1 is located directly in the North Stand for visitors with mobility impairments. Wheelchair ramps are active at South Gate.
          </p>
        </div>

        <div className="bg-white border border-zinc-200 rounded-xl p-5 shadow-sm space-y-2">
          <h4 className="font-bold text-sm text-neutral-800 flex items-center gap-1.5">
            <PhoneCall className="w-4 h-4 text-brand-500" />
            <span>Operational Help</span>
          </h4>
          <p className="text-xs text-zinc-500 leading-normal">
            Need emergency medical assistance or lost an item? Open a direct chat query with our RAG AI assistant floating in the bottom-right corner.
          </p>
        </div>
      </div>
    </div>
  )
}

// Inline imports fix for Accessibility icon inside VenueMapPage
import { Accessibility } from 'lucide-react'
