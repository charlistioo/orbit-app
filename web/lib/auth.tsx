"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from "react";
import { api, MeResponse, SignedInUser } from "./api";

const TOKEN_KEY = "orbit_session_token";

type AuthState = {
  token: string | null;
  onboardingCompleted: boolean | null;
  loading: boolean;
  signIn: (idToken: string) => Promise<void>;
  signOut: () => void;
  refreshMe: () => Promise<void>;
};

const AuthContext = createContext<AuthState | null>(null);

const AUTH_URL = `${process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080/api/v1"}/auth/google`;

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(null);
  const [onboardingCompleted, setOnboardingCompleted] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshMe = useCallback(async () => {
    const stored = typeof window !== "undefined" ? localStorage.getItem(TOKEN_KEY) : null;
    if (!stored) {
      setToken(null);
      setOnboardingCompleted(null);
      setLoading(false);
      return;
    }
    try {
      const me = await api.get<MeResponse>("/me", stored);
      setToken(stored);
      setOnboardingCompleted(me.onboarding_completed);
    } catch {
      localStorage.removeItem(TOKEN_KEY);
      setToken(null);
      setOnboardingCompleted(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refreshMe();
  }, [refreshMe]);

  const signIn = useCallback(async (idToken: string) => {
    const res = await fetch(AUTH_URL, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id_token: idToken }),
    });
    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body.error ?? "Sign-in failed");
    }
    const data: { session_token: string; user: SignedInUser } = await res.json();
    localStorage.setItem(TOKEN_KEY, data.session_token);
    setToken(data.session_token);
    const me = await api.get<MeResponse>("/me", data.session_token);
    setOnboardingCompleted(me.onboarding_completed);
  }, []);

  const signOut = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    setToken(null);
    setOnboardingCompleted(null);
  }, []);

  return (
    <AuthContext.Provider value={{ token, onboardingCompleted, loading, signIn, signOut, refreshMe }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
