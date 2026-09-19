"use client";

import { useEffect, useState } from "react";
import { Lightbulb, Sparkles } from "lucide-react";
import { Card } from "@heroui/react";
import { useAuth } from "@/lib/auth";
import { api, DailyPlan, InsightResult, formatRupiah } from "@/lib/api";

type Tab = "harian" | "mingguan";

// Screen 10 & 11 (Daily / Weekly Insight) - every insight string comes
// straight from GET /insights/daily and /insights/weekly (TASK-012's
// rule-based engine: never claims a trend without enough history,
// mood is only ever a temporal note, never a cause). The weekly bar
// comparison is computed here from real numbers (this week's daily
// plans + Timeline actuals), not invented.
export default function WawasanPage() {
  const [tab, setTab] = useState<Tab>("harian");

  return (
    <div className="flex flex-col gap-5 pt-2">
      <div className="flex items-center gap-2">
        <Lightbulb size={20} color="var(--orbit-accent)" />
        <h1 className="text-lg font-semibold" style={{ color: "var(--orbit-ink)" }}>
          Wawasan
        </h1>
      </div>

      <div className="flex gap-2 rounded-full p-1" style={{ backgroundColor: "var(--orbit-accent-soft)" }}>
        {(["harian", "mingguan"] as Tab[]).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className="flex-1 rounded-full py-2 text-sm font-medium capitalize transition-colors"
            style={{
              backgroundColor: tab === t ? "var(--orbit-surface)" : "transparent",
              color: tab === t ? "var(--orbit-accent)" : "var(--orbit-ink-muted)",
            }}
          >
            {t}
          </button>
        ))}
      </div>

      {tab === "harian" ? <DailyInsight /> : <WeeklyInsight />}
    </div>
  );
}

function InsightList({ title, insights }: { title: string; insights: string[] }) {
  return (
    <Card variant="default" className="!rounded-2xl">
      <Card.Content className="flex flex-col gap-3 py-5">
        <span className="text-xs font-medium uppercase tracking-wide" style={{ color: "var(--orbit-ink-muted)" }}>
          {title}
        </span>
        {insights.length === 0 ? (
          <p className="text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
            Belum ada insight untuk periode ini.
          </p>
        ) : (
          insights.map((text, i) => (
            <div key={i} className="flex items-start gap-2">
              <Sparkles size={15} color="var(--orbit-accent)" className="mt-0.5 shrink-0" />
              <p className="text-sm" style={{ color: "var(--orbit-ink)" }}>
                {text}
              </p>
            </div>
          ))
        )}
      </Card.Content>
    </Card>
  );
}

function DailyInsight() {
  const { token } = useAuth();
  const [result, setResult] = useState<InsightResult | null>(null);

  useEffect(() => {
    if (!token) return;
    api.get<InsightResult>("/insights/daily", token).then(setResult);
  }, [token]);

  if (!result) return null;
  return <InsightList title="Ulasan Harimu" insights={result.insights} />;
}

function WeeklyInsight() {
  const { token } = useAuth();
  const [result, setResult] = useState<InsightResult | null>(null);
  const [plannedTotal, setPlannedTotal] = useState<number | null>(null);
  const [actualTotal, setActualTotal] = useState<number | null>(null);

  useEffect(() => {
    if (!token) return;
    api.get<InsightResult>("/insights/weekly", token).then(setResult);

    // Sum the last 7 days' baseline (planned) totals - real numbers
    // from each day's own plan, 0 for days with no plan at all.
    const dates = Array.from({ length: 7 }, (_, i) => {
      const d = new Date();
      d.setDate(d.getDate() - i);
      return d.toISOString().slice(0, 10);
    });
    Promise.all(
      dates.map((d) =>
        api
          .get<DailyPlan>(`/plans/${d}`, token)
          .then((p) => p.Categories.reduce((s, c) => s + c.PlannedAmount, 0))
          .catch(() => 0)
      )
    ).then((totals) => setPlannedTotal(totals.reduce((a, b) => a + b, 0)));

    api
      .get<{ events: { type: string; occurred_at: string; transaction?: { Amount: number } }[] }>("/transactions", token)
      .then((d) => {
        const weekAgo = dates[6];
        const sum = d.events
          .filter((e) => e.type === "transaction" && e.occurred_at.slice(0, 10) >= weekAgo)
          .reduce((s, e) => s + (e.transaction?.Amount ?? 0), 0);
        setActualTotal(sum);
      });
  }, [token]);

  const maxVal = Math.max(plannedTotal ?? 0, actualTotal ?? 0, 1);

  return (
    <div className="flex flex-col gap-4">
      {plannedTotal != null && actualTotal != null && (
        <Card variant="secondary" className="!rounded-2xl">
          <Card.Content className="flex flex-col gap-4 py-5">
            <span className="text-xs font-medium uppercase tracking-wide" style={{ color: "var(--orbit-ink-muted)" }}>
              Pengeluaran 7 Hari Terakhir
            </span>
            <div className="flex items-end justify-center gap-8">
              <Bar label="Direncanakan" value={plannedTotal} max={maxVal} color="var(--orbit-ink-muted)" />
              <Bar label="Aktual" value={actualTotal} max={maxVal} color="var(--orbit-accent)" />
            </div>
          </Card.Content>
        </Card>
      )}
      {result ? <InsightList title="Pola Pengeluaranmu" insights={result.insights} /> : null}
    </div>
  );
}

function Bar({ label, value, max, color }: { label: string; value: number; max: number; color: string }) {
  const heightPct = Math.max(4, (value / max) * 100);
  return (
    <div className="flex flex-col items-center gap-2">
      <span className="text-xs font-medium" style={{ color: "var(--orbit-ink)" }}>
        {formatRupiah(value)}
      </span>
      <div className="flex h-28 w-12 items-end rounded-lg" style={{ backgroundColor: "var(--orbit-accent-soft)" }}>
        <div className="w-full rounded-lg transition-all" style={{ height: `${heightPct}%`, backgroundColor: color }} />
      </div>
      <span className="text-[11px]" style={{ color: "var(--orbit-ink-muted)" }}>
        {label}
      </span>
    </div>
  );
}
