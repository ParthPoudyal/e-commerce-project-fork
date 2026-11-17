import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '../types';
import { setAuthToken, removeAuthToken, getAuthToken } from '../services/api';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  signIn: (user: User, token?: string) => void;
  signOut: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      signIn: (user, token) => {
        const authToken = token || getAuthToken() || 'mock-jwt-token';
        setAuthToken(authToken);
        set({ user, token: authToken, isAuthenticated: true });
      },
      signOut: () => {
        removeAuthToken();
        set({ user: null, token: null, isAuthenticated: false });
      },
    }),
    {
      name: 'auction-studio-auth',
      onRehydrateStorage: () => (state) => {
        if (state?.token) {
          setAuthToken(state.token);
        } else {
          const stored = getAuthToken();
          if (stored && state) {
            state.token = stored;
            if (state.user) {
              state.isAuthenticated = true;
            }
          }
        }
      },
    }
  )
);
