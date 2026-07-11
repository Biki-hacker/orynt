import React, { useState, useRef, useEffect } from 'react'
import { ChevronDown, Check } from 'lucide-react'

// Badge Component
interface BadgeProps {
  children: React.ReactNode;
  variant?: 'info' | 'success' | 'warning' | 'error' | 'neutral';
}

export const Badge: React.FC<BadgeProps> = ({ children, variant = 'neutral' }) => {
  const styles = {
    info: 'bg-blue-50 text-blue-700 border-blue-200',
    success: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    warning: 'bg-amber-50 text-amber-700 border-amber-200',
    error: 'bg-red-50 text-red-700 border-red-200',
    neutral: 'bg-zinc-100 text-zinc-700 border-zinc-200'
  }

  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${styles[variant]}`}>
      {children}
    </span>
  )
}

// Card Component
interface CardProps {
  children: React.ReactNode;
  className?: string;
  onClick?: () => void;
}

export const Card: React.FC<CardProps> = ({ children, className = '', onClick }) => {
  return (
    <div
      onClick={onClick}
      className={`bg-white border border-zinc-200 rounded-lg shadow-sm hover:shadow-md transition-shadow duration-200 p-5 ${onClick ? 'cursor-pointer' : ''} ${className}`}
    >
      {children}
    </div>
  )
}

// Modal Component
interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
}

export const Modal: React.FC<ModalProps> = ({ isOpen, onClose, title, children }) => {
  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-zinc-950/40 backdrop-blur-sm">
      <div className="bg-white border border-zinc-200 rounded-lg shadow-lg w-full max-w-lg overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        <div className="flex items-center justify-between px-5 py-4 border-b border-zinc-100">
          <h3 className="font-semibold text-neutral-800">{title}</h3>
          <button onClick={onClose} aria-label="Close dialog" className="text-zinc-400 hover:text-zinc-600 transition-colors">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div className="px-5 py-4 overflow-y-auto max-h-[80vh]">{children}</div>
      </div>
    </div>
  )
}

// Skeleton loading Component
export const Skeleton: React.FC<{ className?: string }> = ({ className = '' }) => {
  return <div className={`animate-pulse bg-zinc-200 rounded ${className}`} />
}

// Spinner loading Component
export const Spinner: React.FC<{ size?: 'sm' | 'md' | 'lg' }> = ({ size = 'md' }) => {
  const sizes = {
    sm: 'w-4 h-4 border-2',
    md: 'w-6 h-6 border-2',
    lg: 'w-8 h-8 border-3'
  }
  return (
    <div className={`animate-spin rounded-full border-t-brand-500 border-zinc-200 ${sizes[size]}`} />
  )
}

// Select Component
interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  containerClassName?: string;
}

export const Select = React.forwardRef<HTMLSelectElement, SelectProps>(
  ({ label, containerClassName = '', className = '', children, ...props }, ref) => {
    return (
      <div className={`space-y-1.5 ${containerClassName}`}>
        {label && (
          <label className="text-xs font-semibold text-zinc-500 block">
            {label}
          </label>
        )}
        <div className="relative">
          <select
            ref={ref}
            className={`w-full appearance-none bg-white border border-zinc-200 rounded-lg p-2 pr-10 text-sm focus:outline-none focus:border-brand-500 transition-all cursor-pointer font-medium text-neutral-800 focus:ring-2 focus:ring-brand-500/10 ${className}`}
            {...props}
          >
            {children}
          </select>
          <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3 text-zinc-400">
            <ChevronDown className="w-4 h-4" />
          </div>
        </div>
      </div>
    )
  }
)
Select.displayName = 'Select'

// CustomSelect Component (Modern floating overlay list)
interface Option {
  value: string;
  label: string;
}

interface CustomSelectProps {
  label?: string;
  value: string;
  onChange: (value: string) => void;
  options: Option[];
  className?: string;
  containerClassName?: string;
}

export const CustomSelect: React.FC<CustomSelectProps> = ({
  label,
  value,
  onChange,
  options,
  className = '',
  containerClassName = ''
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  const selectedOption = options.find((opt) => opt.value === value) || options[0]

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => dropdownRef.current ? document.removeEventListener('mousedown', handleClickOutside) : undefined
  }, [])

  return (
    <div ref={dropdownRef} className={`space-y-0.5 ${containerClassName}`}>
      {label && (
        <label className="text-[10px] font-medium text-zinc-400 block">
          {label}
        </label>
      )}
      <div className="relative">
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className={`w-full h-9 flex items-center justify-between bg-white border border-zinc-200 rounded-lg px-3 text-sm text-neutral-800 text-left focus:outline-none focus:ring-2 focus:ring-brand-500/10 focus:border-brand-500 transition-all cursor-pointer font-medium ${className}`}
        >
          <span className="truncate">{selectedOption ? selectedOption.label : 'Select...'}</span>
          <ChevronDown className={`w-4 h-4 text-zinc-400 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`} />
        </button>

        {isOpen && (
          <div className="absolute z-50 w-full mt-1 bg-white border border-zinc-200 rounded-lg shadow-lg max-h-60 overflow-y-auto py-1 animate-in fade-in duration-100">
            {options.map((opt) => {
              const isSelected = opt.value === value
              return (
                <button
                  key={opt.value}
                  type="button"
                  onClick={() => {
                    onChange(opt.value)
                    setIsOpen(false)
                  }}
                  className={`w-full flex items-center justify-between px-3 py-2 text-xs text-left text-neutral-700 hover:bg-zinc-50 transition-colors font-medium ${
                    isSelected ? 'bg-brand-50/50 text-brand-700 hover:bg-brand-50' : ''
                  }`}
                >
                  <span className="truncate">{opt.label}</span>
                  {isSelected && <Check className="w-3.5 h-3.5 text-brand-600 flex-shrink-0 ml-2" />}
                </button>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
