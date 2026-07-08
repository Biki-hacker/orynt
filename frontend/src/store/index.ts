import { create } from 'zustand'
import { User } from '../types'

interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  setAuth: (user: User, token: string) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>((set) => {
  // Read initial state from localStorage for persistence across browser refreshes
  const storedUser = localStorage.getItem('orynt_user');
  const storedToken = localStorage.getItem('orynt_token');

  return {
    user: storedUser ? JSON.parse(storedUser) : null,
    accessToken: storedToken || null,
    isAuthenticated: !!storedToken,
    setAuth: (user, token) => {
      localStorage.setItem('orynt_user', JSON.stringify(user));
      localStorage.setItem('orynt_token', token);
      set({ user, accessToken: token, isAuthenticated: true });
    },
    clearAuth: () => {
      localStorage.removeItem('orynt_user');
      localStorage.removeItem('orynt_token');
      set({ user: null, accessToken: null, isAuthenticated: false });
    }
  }
})

export interface ToastMessage {
  id: string;
  title: string;
  content: string;
  type: 'emergency' | 'weather' | 'crowd' | 'transport' | 'info';
  severity: 'critical' | 'high' | 'info';
  createdAt: Date;
}

interface NotificationState {
  notifications: ToastMessage[];
  addNotification: (notification: Omit<ToastMessage, 'id' | 'createdAt'>) => void;
  removeNotification: (id: string) => void;
  clearAll: () => void;
}

export const useNotificationStore = create<NotificationState>((set) => ({
  notifications: [],
  addNotification: (n) => {
    const id = Math.random().toString(36).substring(2, 9);
    const newNotification: ToastMessage = {
      ...n,
      id,
      createdAt: new Date()
    };
    set((state) => ({
      // Keep only last 8 toast notifications to avoid screen cluttering
      notifications: [newNotification, ...state.notifications].slice(0, 8)
    }));
  },
  removeNotification: (id) => {
    set((state) => ({
      notifications: state.notifications.filter((n) => n.id !== id)
    }));
  },
  clearAll: () => set({ notifications: [] })
}))
