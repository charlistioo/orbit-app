"use client";

import { useEffect, useState } from "react";
import { PiggyBank, X, TrendingUp, TrendingDown } from "lucide-react";
import Link from "next/link";
import { Card, ProgressBar, Button, buttonVariants } from "@heroui/react";
import { useAuth } from "@/lib/auth";
import {
  api,
  BankData,
  Category,
  DailyPlan,
  TimelineEvent,
  currentTierFloor,
  formatRupiah,
  nextBadgeTier,
  todayStr,
} from "@/lib/api";

// Screen 8 (Consumption Bank) - a dedicated page instead of only a
// dashboard card. Every number is real: balance/badge from
// GET /consumption-bank, per-category "Direncanakan/Terpakai/
// Tersimpan" from today's plan (GET /plans/{date}) combined with
// today's actual spend per category, computed client-side by filtering
// the already-fetched Timeline - no new backend aggregation invented.
export default function BankPage() {
  const { token } = useAuth();
  const [bank, setBank] = useState<BankData | null>(null);
  const [plan, setPlan] = useState<DailyPlan | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [error, setError] = useState("");
  const [showToday, setShowToday] = useState(false);

  useEffect(() => {
    if (!token) return;
    Promise.all([
      api.get<BankData>("/consumption-bank", token),
      api.get<{ categories: Category[] }>("/categories", token),
      api.get<{ events: TimelineEvent[] }>("/transactions", token),
      api.get<DailyPlan>(`/plans/${todayStr()}`, token).catch(() => null),
    ])
      .then(([b, c, t, p]) => {
        setBank(b);
        setCategories(c.categories);
        setEvents(t.events);
        setPlan(p);
      })
      .catch(() => setError("Gagal memuat data Bank Konsumsi"));
  }, [token]);

  if (error) return <p className="pt-2 text-sm" style={{ color: "#b3261e" }}>{error}</p>;
  if (!bank) return null;

  const categoryName = (id: string) => categories.find((c) => c.ID === id)?.Name ?? "Kategori";
  const actualFor = (categoryId: string) =>
    events
      .filter((e) => e.type === "transaction" && e.transaction?.CategoryID === categoryId && e.occurred_at.slice(0, 10) === todayStr())
      .reduce((sum, e) => sum + (e.transaction?.Amount ?? 0), 0);

  const floor = currentTierFloor(bank.balance);
  const next = nextBadgeTier(bank.balance);
  const tierProgress = next ? (bank.balance - floor) / (next.threshold - floor) : 1;

  return (
    <div className="flex flex-col gap-5 pt-2">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold" style={{ color: "var(--orbit-ink)" }}>
          Bank Konsumsimu
        </h1>
        <Link href="/" className={buttonVariants({ variant: "ghost", size: "sm", isIconOnly: true })} aria-label="Tutup">
          <X size={18} />
        </Link>
      </div>

      <Card variant="secondary" className="!rounded-2xl">
        <Card.Content className="flex flex-col items-center gap-3 py-7 text-center">
          <div
            className="flex h-14 w-14 items-center justify-center rounded-full"
            style={{ backgroundColor: "var(--orbit-sage-soft)" }}
          >
            <PiggyBank size={26} color="var(--orbit-sage)" />
          </div>
          <span className="text-xs font-medium uppercase tracking-wide" style={{ color: "var(--orbit-ink-muted)" }}>
            Fleksibilitas Tersimpan
          </span>
          <span className="text-3xl font-bold" style={{ color: "var(--orbit-ink)" }}>
            {formatRupiah(bank.balance)}
          </span>
          {bank.badge ? (
            <span
              className="rounded-full px-3 py-1 text-xs font-medium"
              style={{ backgroundColor: "var(--orbit-accent-soft)", color: "var(--orbit-accent)" }}
            >
              {bank.badge}
            </span>
          ) : null}

          <div className="w-full pt-2">
            <ProgressBar value={tierProgress * 100} size="sm" aria-label="Progres menuju tier berikutnya" />
            <div className="flex justify-between pt-1 text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
              <span>Ambang {formatRupiah(floor)}</span>
              <span>{next ? `Menuju ${next.name} (${formatRupiah(next.threshold)})` : "Tier tertinggi"}</span>
            </div>
          </div>
        </Card.Content>
      </Card>

      {plan && plan.Categories.length > 0 && (
        <Button
          variant="secondary"
          fullWidth
          onPress={() => setShowToday((v) => !v)}
        >
          {showToday ? "Sembunyikan Penghematan Hari Ini" : "Lihat Penghematan Anda Hari Ini"}
        </Button>
      )}

      {showToday && plan && plan.Categories.length > 0 && (() => {
        const rows = plan.Categories.map((pc) => {
          const actual = actualFor(pc.CategoryID);
          const diff = pc.PlannedAmount - actual;
          return { pc, actual, diff, hemat: diff >= 0 };
        });
        const hematCount = rows.filter((r) => r.hemat).length;
        const borosCount = rows.length - hematCount;
        const netTotal = rows.reduce((sum, r) => sum + r.diff, 0);

        return (
          <div className="flex flex-col gap-3">
            <div className="grid grid-cols-2 gap-3">
              <Card variant="default" className="!rounded-2xl">
                <Card.Content className="flex flex-col items-center gap-1 py-4">
                  <TrendingUp size={18} color="var(--orbit-sage)" />
                  <span className="text-lg font-semibold" style={{ color: "var(--orbit-ink)" }}>
                    {hematCount}
                  </span>
                  <span className="text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
                    Kategori Hemat
                  </span>
                </Card.Content>
              </Card>
              <Card variant="default" className="!rounded-2xl">
                <Card.Content className="flex flex-col items-center gap-1 py-4">
                  <TrendingDown size={18} color="var(--orbit-warn)" />
                  <span className="text-lg font-semibold" style={{ color: "var(--orbit-ink)" }}>
                    {borosCount}
                  </span>
                  <span className="text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
                    Kategori Boros
                  </span>
                </Card.Content>
              </Card>
            </div>

            <Card variant="secondary" className="!rounded-2xl">
              <Card.Content className="flex items-center justify-between py-4">
                <span className="text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
                  Total hari ini
                </span>
                <span
                  className="text-base font-semibold"
                  style={{ color: netTotal >= 0 ? "var(--orbit-sage)" : "var(--orbit-warn)" }}
                >
                  {netTotal >= 0 ? "Hemat " : "Boros "}
                  {formatRupiah(Math.abs(netTotal))}
                </span>
              </Card.Content>
            </Card>

            <div className="flex flex-col gap-2">
              {rows.map(({ pc, actual, diff, hemat }) => (
                <Card key={pc.ID} variant="default" className="!rounded-2xl">
                  <Card.Content className="flex flex-col gap-1.5 py-4">
                    <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
                      {categoryName(pc.CategoryID)}
                    </span>
                    <div className="flex justify-between text-xs">
                      <span style={{ color: "var(--orbit-ink-muted)" }}>Direncanakan</span>
                      <span style={{ color: "var(--orbit-ink)" }}>{formatRupiah(pc.PlannedAmount)}</span>
                    </div>
                    <div className="flex justify-between text-xs">
                      <span style={{ color: "var(--orbit-ink-muted)" }}>Terpakai</span>
                      <span style={{ color: "var(--orbit-ink)" }}>{formatRupiah(actual)}</span>
                    </div>
                    <div className="flex justify-between text-xs">
                      <span style={{ color: "var(--orbit-ink-muted)" }}>{hemat ? "Hemat" : "Boros"}</span>
                      <span style={{ color: hemat ? "var(--orbit-sage)" : "var(--orbit-warn)" }}>
                        {formatRupiah(Math.abs(diff))}
                      </span>
                    </div>
                  </Card.Content>
                </Card>
              ))}
            </div>
          </div>
        );
      })()}

      <p className="text-center text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
        Saldo ini tumbuh dari rencana yang tidak terpakai, dan bisa dipakai menutup selisih pengeluaran (maks. 10% saldo per transaksi).
      </p>
    </div>
  );
}
