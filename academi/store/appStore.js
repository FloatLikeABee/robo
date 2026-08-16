import { create } from 'zustand';
import { getTheme } from '../themes/theme';

const useAppStore = create((set, get) => ({
  isDarkMode: true,
  themeMode: 'dark',
  theme: getTheme('dark'),
  aiTone: 'casual',
  aiDepth: 'intermediate',
  notifications: [],
  savedContent: [],
  learningStreak: 12,
  contributions: 45,
  
  toggleTheme: () => set((state) => ({
    isDarkMode: !state.isDarkMode,
    themeMode: state.isDarkMode ? 'light' : 'dark',
    theme: getTheme(state.isDarkMode ? 'light' : 'dark'),
  })),

  initializeTheme: () => set((state) => ({
    theme: getTheme(state.themeMode),
  })),
  
  setAITone: (tone) => set({ aiTone: tone }),
  setAIDepth: (depth) => set({ aiDepth: depth }),
  
  addNotification: (notification) => set((state) => ({
    notifications: [notification, ...state.notifications],
  })),
  
  markNotificationRead: (id) => set((state) => ({
    notifications: state.notifications.map((n) =>
      n.id === id ? { ...n, read: true } : n
    ),
  })),
  
  saveContent: (content) => set((state) => ({
    savedContent: [...state.savedContent, content],
  })),
}));

export default useAppStore;
