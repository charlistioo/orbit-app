"use client";

import { useState } from "react";
import { GoogleLogin } from "@react-oauth/google";
import { Orbit } from "lucide-react";
import { useAuth } from "@/lib/auth";

// Google Sign-In also acts as "register" - the backend auto-creates a
// profile on first sign-in (unchanged, from TASK-003); onboarding
// (Nama/Umur/Budget) is a separate step handled by AppShell's redirect.
export default function LoginPage() {
  const { signIn } = useAuth();
  const [error, setError] = useState("");

  return (
    <div className="flex min-h-[calc(100vh-3rem)] flex-col items-center justify-center gap-8 text-center">
      <div className="flex flex-col items-center gap-3">
        <Orbit size={36} color="var(--orbit-accent)" strokeWidth={1.6} />
        <h1 className="text-xl font-semibold" style={{ color: "var(--orbit-ink)" }}>
          ORBIT
        </h1>
        <p className="text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
          Spend freely, stay aware
        </p>
      </div>

      <GoogleLogin
        onSuccess={async (credential) => {
          setError("");
          if (!credential.credential) {
            setError("Google tidak mengembalikan token yang valid");
            return;
          }
          try {
            await signIn(credential.credential);
          } catch (err) {
            setError(err instanceof Error ? err.message : "Gagal masuk");
          }
        }}
        onError={() => setError("Gagal masuk dengan Google")}
      />

      {error ? (
        <p className="text-sm" style={{ color: "#b3261e" }}>
          {error}
        </p>
      ) : null}
    </div>
  );
}
