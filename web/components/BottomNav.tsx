"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, History, Lightbulb, User } from "lucide-react";

const TABS = [
  { href: "/", label: "Beranda", icon: Home },
  { href: "/riwayat", label: "Riwayat", icon: History },
  { href: "/wawasan", label: "Wawasan", icon: Lightbulb },
  { href: "/profil", label: "Profil", icon: User },
];

// Bottom tab bar (TASK-016) - 4 tabs, replacing the earlier RN app's
// 5-tab+FAB layout, matching the new reference screens' navigation.
export default function BottomNav() {
  const pathname = usePathname();

  return (
    <nav className="fixed bottom-0 inset-x-0 z-40 border-t border-[var(--orbit-border)] bg-[var(--orbit-surface)]/95 backdrop-blur">
      <div className="mx-auto flex max-w-md items-center justify-around px-2 py-2 pb-[calc(env(safe-area-inset-bottom)+0.5rem)]">
        {TABS.map((tab) => {
          const active = tab.href === "/" ? pathname === "/" : pathname.startsWith(tab.href);
          const Icon = tab.icon;
          return (
            <Link
              key={tab.href}
              href={tab.href}
              className="flex min-w-[56px] flex-col items-center gap-1 rounded-xl px-3 py-1.5"
            >
              <Icon
                size={22}
                strokeWidth={active ? 2.4 : 1.8}
                color={active ? "var(--orbit-accent)" : "var(--orbit-ink-muted)"}
              />
              <span
                className="text-[11px] font-medium"
                style={{ color: active ? "var(--orbit-accent)" : "var(--orbit-ink-muted)" }}
              >
                {tab.label}
              </span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
