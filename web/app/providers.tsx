"use client";

import { ReactNode } from "react";
import { GoogleOAuthProvider } from "@react-oauth/google";
import { AuthProvider } from "@/lib/auth";
import AppShell from "@/components/AppShell";

const GOOGLE_CLIENT_ID =
  process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID ??
  "595683522917-al9clne9te05q8gb55a1t8v2t0kvf39u.apps.googleusercontent.com";

export function Providers({ children }: { children: ReactNode }) {
  return (
    <GoogleOAuthProvider clientId={GOOGLE_CLIENT_ID}>
      <AuthProvider>
        <AppShell>{children}</AppShell>
      </AuthProvider>
    </GoogleOAuthProvider>
  );
}
