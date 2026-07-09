import React from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { PublicLayout, DashboardLayout } from './layouts/Layouts'
import { Home } from './pages/Public/Home'
import { MatchDetails } from './pages/Public/MatchDetails'
import { VenueMapPage } from './pages/Public/VenueMapPage'
import { TransportParking } from './pages/Public/TransportParking'
import { FoodFAQ } from './pages/Public/FoodFAQ'
import { Sustainability } from './pages/Public/Sustainability'
import { Login } from './pages/Auth/Login'
import { StaffDashboard } from './pages/Staff/StaffDashboard'
import { AdminDashboard } from './pages/Admin/AdminDashboard'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1
    }
  }
})

export const App: React.FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          {/* Public Views */}
          <Route element={<PublicLayout />}>
            <Route path="/" element={<Home />} />
            <Route path="/match/:id" element={<MatchDetails />} />
            <Route path="/map" element={<VenueMapPage />} />
            <Route path="/transport" element={<TransportParking />} />
            <Route path="/food-faq" element={<FoodFAQ />} />
            <Route path="/sustainability" element={<Sustainability />} />
            <Route path="/login" element={<Login />} />
          </Route>

          {/* Secure Admin & Staff views */}
          <Route element={<DashboardLayout />}>
            <Route path="/staff" element={<StaffDashboard />} />
            <Route path="/admin" element={<AdminDashboard />} />
          </Route>

          {/* Fallback */}
          <Route path="*" element={
            <div className="flex flex-col items-center justify-center min-h-[400px] text-center font-sans">
              <h2 className="text-xl font-bold text-neutral-800">Page Not Found</h2>
              <p className="text-zinc-500 text-sm mt-1">The page you were looking for doesn't exist.</p>
              <a href="/" className="mt-4 text-sm font-semibold text-brand-600 hover:underline">
                Return to Dashboard
              </a>
            </div>
          } />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
