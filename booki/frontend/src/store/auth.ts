import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';
import { syncPlatformSessionCookie } from '../api/client';

type AuthState = {
  accessToken: string | null;
  refreshToken: string | null;
  /** UsersPanel JWT from login / platform-session — required for /api/messages/* on UsersPanel. */
  platformJwt: string | null;
  setTokens: (a: string | null, r: string | null, platformJwt?: string | null) => void;
  logout: () => void;
};

export const useAuth = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      platformJwt: null,
      setTokens: (accessToken, refreshToken, platformJwt) =>
        set({
          accessToken,
          refreshToken,
          ...(platformJwt !== undefined ? { platformJwt: platformJwt ?? null } : {}),
        }),
      logout: () => {
        syncPlatformSessionCookie(null);
        set({ accessToken: null, refreshToken: null, platformJwt: null });
      },
    }),
    {
      name: 'academi-auth',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        platformJwt: state.platformJwt,
      }),
    }
  )
);
