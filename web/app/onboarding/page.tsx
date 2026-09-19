"use client";

import Link from "next/link";
import { Orbit, Wallet, Smile, TrendingDown, Eye } from "lucide-react";
import { Card, buttonVariants } from "@heroui/react";

const GOALS = [
  { icon: Wallet, label: "Belanja lebih sadar" },
  { icon: TrendingDown, label: "Kurangi belanja impulsif" },
  { icon: Smile, label: "Kendalikan self-reward" },
  { icon: Eye, label: "Pahami pola pengeluaran" },
];

// Screen 1 (Onboarding value prop) - TASK-016.
export default function OnboardingPage() {
  return (
    <div className="flex min-h-[calc(100vh-3rem)] flex-col justify-between gap-8">
      <div className="flex flex-col items-center gap-6 pt-8 text-center">
        <div
          className="flex h-20 w-20 items-center justify-center rounded-full"
          style={{ backgroundColor: "var(--orbit-accent-soft)" }}
        >
          <Orbit size={40} color="var(--orbit-accent)" strokeWidth={1.6} />
        </div>

        <div className="flex flex-col gap-2">
          <h1 className="text-2xl font-semibold leading-tight" style={{ color: "var(--orbit-ink)" }}>
            Understand your spending,
            <br />
            not just your money.
          </h1>
          <p className="mx-auto max-w-xs text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
            ORBIT membantu merencanakan konsumsi harian dan memahami
            bagaimana suasana hati memengaruhi pengeluaranmu.
          </p>
        </div>

        <div className="grid w-full grid-cols-2 gap-3">
          {GOALS.map((g) => (
            <Card key={g.label} variant="secondary" className="!rounded-2xl">
              <Card.Content className="flex flex-col items-center gap-2 py-4">
                <g.icon size={22} color="var(--orbit-sage)" strokeWidth={1.8} />
                <span className="text-xs font-medium" style={{ color: "var(--orbit-ink)" }}>
                  {g.label}
                </span>
              </Card.Content>
            </Card>
          ))}
        </div>
      </div>

      <Link
        href="/login"
        className={buttonVariants({ variant: "primary", size: "lg", fullWidth: true })}
        style={{ backgroundColor: "var(--orbit-accent)", borderColor: "var(--orbit-accent)" }}
      >
        Mulai Perjalananmu
      </Link>
    </div>
  );
}
