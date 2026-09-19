"use client";

import { ReactNode, useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { Orbit } from "lucide-react";
import { useAuth } from "@/lib/auth";
import BottomNav from "./BottomNav";

const PUBLIC_ROUTES = ["/onboarding", "/login"];
const ONBOARDING_PROFILE_ROUTE = "/onboarding/profile";

// Routes every signed-in-and-onboarded user can reach - the bottom nav
// only renders on these.
function isAppRoute(pathname: string) {
  return !PUBLIC_ROUTES.includes(pathname) && pathname !== ONBOARDING_PROFILE_ROUTE;
}

export default function AppShell({ children }: { children: ReactNode }) {
  const { token, onboardingCompleted, loading } = useAuth();
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    if (loading) return;

    if (!token) {
      if (!PUBLIC_ROUTES.includes(pathname)) router.replace("/onboarding");
      return;
    }

    if (!onboardingCompleted) {
      if (pathname !== ONBOARDING_PROFILE_ROUTE) router.replace(ONBOARDING_PROFILE_ROUTE);
      return;
    }

    if (PUBLIC_ROUTES.includes(pathname) || pathname === ONBOARDING_PROFILE_ROUTE) {
      router.replace("/");
    }
  }, [loading, token, onboardingCompleted, pathname, router]);

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Orbit className="animate-spin" size={28} color="var(--orbit-accent)" />
      </div>
    );
  }

  const showNav = token && onboardingCompleted && isAppRoute(pathname);

  return (
    <div className="flex min-h-screen flex-col">
      <main className={`mx-auto w-full max-w-md flex-1 px-4 ${showNav ? "pb-24" : "pb-4"} pt-6`}>
        {children}
      </main>
      {showNav && <BottomNav />}
    </div>
  );
}
