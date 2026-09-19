"use client";

import { useEffect, useMemo, useState } from "react";
import { History, Receipt, Smile, PiggyBank, SlidersHorizontal } from "lucide-react";
import { Card } from "@heroui/react";
import { useAuth } from "@/lib/auth";
import { api, TimelineEvent, formatRupiah } from "@/lib/api";

const MOOD_LABEL: Record<string, string> = {
  very_good: "Sangat Baik",
  good: "Baik",
  neutral: "Biasa Saja",
  stressed: "Stres",
  sad: "Sedih",
};

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit" });
}
function formatDateHeader(iso: string): string {
  return new Date(iso).toLocaleDateString("id-ID", { weekday: "long", day: "numeric", month: "long" });
}

// Screen 9 (Daily Timeline / Riwayat) - the merged, server-sorted
// Timeline (TASK-010's design), grouped by date client-side for
// display only. Category/item names come straight from the backend's
// enriched Timeline response (TASK-016) - no more raw ids shown.
export default function RiwayatPage() {
  const { token } = useAuth();
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!token) return;
    api
      .get<{ events: TimelineEvent[] }>("/transactions", token)
      .then((d) => setEvents(d.events))
      .catch(() => setError("Gagal memuat riwayat"));
  }, [token]);

  const grouped = useMemo(() => {
    const reversed = [...events].reverse();
    const groups: { date: string; events: TimelineEvent[] }[] = [];
    for (const e of reversed) {
      const date = e.occurred_at.slice(0, 10);
      let g = groups.find((g) => g.date === date);
      if (!g) {
        g = { date, events: [] };
        groups.push(g);
      }
      g.events.push(e);
    }
    return groups;
  }, [events]);

  return (
    <div className="flex flex-col gap-5 pt-2">
      <div className="flex items-center gap-2">
        <History size={20} color="var(--orbit-accent)" />
        <h1 className="text-lg font-semibold" style={{ color: "var(--orbit-ink)" }}>
          Riwayat
        </h1>
      </div>

      {error ? <p className="text-sm" style={{ color: "#b3261e" }}>{error}</p> : null}

      {grouped.length === 0 && !error ? (
        <Card variant="default" className="!rounded-2xl">
          <Card.Content className="py-5">
            <p className="text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
              Belum ada aktivitas tercatat.
            </p>
          </Card.Content>
        </Card>
      ) : (
        grouped.map((g) => (
          <div key={g.date} className="flex flex-col gap-2">
            <span className="text-xs font-medium uppercase tracking-wide" style={{ color: "var(--orbit-ink-muted)" }}>
              {formatDateHeader(g.date)}
            </span>
            <div className="flex flex-col">
              {g.events.map((e, i) => (
                <TimelineRow key={i} event={e} />
              ))}
            </div>
          </div>
        ))
      )}
    </div>
  );
}

function TimelineRow({ event }: { event: TimelineEvent }) {
  const icon =
    event.type === "transaction" ? (
      <Receipt size={16} color="var(--orbit-accent)" />
    ) : event.type === "mood" ? (
      <Smile size={16} color="var(--orbit-sage)" />
    ) : event.type === "bank_ledger" ? (
      <PiggyBank size={16} color="var(--orbit-sage)" />
    ) : (
      <SlidersHorizontal size={16} color="var(--orbit-ink-muted)" />
    );

  let title = "Aktivitas";
  let trailing: string | null = null;

  if (event.type === "transaction" && event.transaction) {
    title = event.transaction.ItemName || event.transaction.CategoryName;
    trailing = formatRupiah(event.transaction.Amount);
  } else if (event.type === "mood" && event.mood) {
    title = `Mood: ${MOOD_LABEL[event.mood.Mood] ?? event.mood.Mood}`;
  } else if (event.type === "bank_ledger" && event.bank_ledger) {
    const delta = event.bank_ledger.DeltaAmount;
    title = delta >= 0 ? "Buffer Bertambah" : "Buffer Diterapkan";
    trailing = `${delta >= 0 ? "+" : ""}${formatRupiah(delta)}`;
  } else if (event.type === "plan_adjustment" && event.plan_adjustment) {
    title = `Rencana ${event.plan_adjustment.CategoryName} disesuaikan`;
  }

  return (
    <div className="flex items-center gap-3 border-b py-3 last:border-b-0" style={{ borderColor: "var(--orbit-border)" }}>
      <div
        className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full"
        style={{ backgroundColor: "var(--orbit-accent-soft)" }}
      >
        {icon}
      </div>
      <div className="flex flex-1 flex-col">
        <span className="text-sm" style={{ color: "var(--orbit-ink)" }}>
          {title}
        </span>
        <span className="text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
          {formatTime(event.occurred_at)}
        </span>
      </div>
      {trailing ? (
        <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
          {trailing}
        </span>
      ) : null}
    </div>
  );
}
