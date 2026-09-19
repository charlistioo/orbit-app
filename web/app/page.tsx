"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Orbit, Receipt, Smile, SlidersHorizontal, PiggyBank, Plus } from "lucide-react";
import { Card, Button, Input, ProgressBar, ProgressCircle } from "@heroui/react";
import { useAuth } from "@/lib/auth";
import {
  api,
  ApiError,
  Category,
  DailyPlan,
  HomeData,
  Profile,
  TimelineEvent,
  formatRupiah,
  todayStr,
} from "@/lib/api";

const MOOD_LABEL: Record<string, string> = {
  very_good: "Sangat baik",
  good: "Merasa baik",
  neutral: "Biasa saja",
  stressed: "Stres",
  sad: "Sedih",
};

function greeting(): string {
  const h = new Date().getHours();
  if (h < 11) return "Selamat pagi";
  if (h < 15) return "Selamat siang";
  if (h < 18) return "Selamat sore";
  return "Selamat malam";
}

export default function BerandaPage() {
  const { token } = useAuth();
  const [loading, setLoading] = useState(true);
  const [hasPlan, setHasPlan] = useState<boolean | null>(null);
  const [profile, setProfile] = useState<Profile | null>(null);

  useEffect(() => {
    if (!token) return;
    (async () => {
      try {
        await api.get<DailyPlan>(`/plans/${todayStr()}`, token);
        setHasPlan(true);
      } catch (err) {
        if (err instanceof ApiError && err.status === 404) setHasPlan(false);
        else setHasPlan(false);
      }
      try {
        setProfile(await api.get<Profile>("/me/profile", token));
      } catch {
        // non-critical
      }
      setLoading(false);
    })();
  }, [token]);

  if (loading || hasPlan === null) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <Orbit className="animate-spin" size={26} color="var(--orbit-accent)" />
      </div>
    );
  }

  return hasPlan ? <MainDashboard /> : <PlanToday defaultBaseline={profile?.default_baseline_amount ?? null} onCreated={() => setHasPlan(true)} />;
}

// ---------- Screen 2b: Rencanakan Harimu ----------

function PlanToday({
  defaultBaseline,
  onCreated,
}: {
  defaultBaseline: number | null;
  onCreated: () => void;
}) {
  const { token } = useAuth();
  const [categories, setCategories] = useState<Category[]>([]);
  const [totalText, setTotalText] = useState(defaultBaseline ? String(Math.round(defaultBaseline)) : "");
  const [rows, setRows] = useState<{ categoryId: string; name: string; amount: string }[]>([]);
  const [newCategoryName, setNewCategoryName] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!token) return;
    api.get<{ categories: Category[] }>("/categories", token).then((d) => {
      setCategories(d.categories);
      setRows(d.categories.map((c) => ({ categoryId: c.ID, name: c.Name, amount: "" })));
    });
  }, [token]);

  const addCategory = async () => {
    if (!newCategoryName.trim()) return;
    const created = await api.post<Category>("/categories", token, { name: newCategoryName.trim() });
    setCategories((prev) => [created, ...prev]);
    setRows((prev) => [{ categoryId: created.ID, name: created.Name, amount: "" }, ...prev]);
    setNewCategoryName("");
  };

  const totalPlanned = rows.reduce((sum, r) => sum + (parseInt(r.amount, 10) || 0), 0);

  const submit = async () => {
    const baseline = parseInt(totalText.replace(/\D/g, ""), 10);
    if (!baseline || baseline <= 0) return setError("Isi total budget yang disarankan");
    setError("");
    setSaving(true);
    try {
      await api.post("/plans", token, {
        plan_date: todayStr(),
        baseline_amount: baseline,
        categories: rows
          .filter((r) => parseInt(r.amount, 10) > 0)
          .map((r) => ({ category_id: r.categoryId, planned_amount: parseInt(r.amount, 10) })),
      });
      onCreated();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal menyimpan rencana");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col gap-5 pt-2">
      <h1 className="text-xl font-semibold" style={{ color: "var(--orbit-ink)" }}>
        Rencanakan harimu
      </h1>

      <Card variant="secondary" className="!rounded-2xl">
        <Card.Content className="flex flex-col gap-2 py-5">
          <span className="text-xs font-medium uppercase tracking-wide" style={{ color: "var(--orbit-ink-muted)" }}>
            Total suggested budget
          </span>
          <Input
            inputMode="numeric"
            fullWidth
            value={totalText}
            onChange={(e) => setTotalText(e.target.value.replace(/\D/g, ""))}
            placeholder="90000"
          />
        </Card.Content>
      </Card>

      <div className="flex flex-col gap-2">
        <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
          Alokasi kategori
        </span>
        <Card variant="default" className="!rounded-2xl">
          <Card.Content className="flex flex-col py-1">
            {rows.length === 0 ? (
              <p className="py-3 text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
                Belum ada kategori - tambahkan satu di bawah.
              </p>
            ) : (
              rows.map((row, i) => (
                <div
                  key={row.categoryId}
                  className="flex items-center gap-3 border-b py-2.5 last:border-b-0"
                  style={{ borderColor: "var(--orbit-border)" }}
                >
                  <span className="flex-1 truncate text-sm" style={{ color: "var(--orbit-ink)" }}>
                    {row.name}
                  </span>
                  <Input
                    inputMode="numeric"
                    className="!w-28"
                    value={row.amount}
                    placeholder="0"
                    onChange={(e) => {
                      const v = e.target.value.replace(/\D/g, "");
                      setRows((prev) => prev.map((r, idx) => (idx === i ? { ...r, amount: v } : r)));
                    }}
                  />
                </div>
              ))
            )}
          </Card.Content>
        </Card>
        <div className="flex items-center gap-2 pt-1">
          <Input
            fullWidth
            placeholder="Kategori baru"
            value={newCategoryName}
            onChange={(e) => setNewCategoryName(e.target.value)}
          />
          <Button variant="secondary" isIconOnly onPress={addCategory} aria-label="Tambah kategori">
            <Plus size={18} />
          </Button>
        </div>
      </div>

      <div className="flex items-center justify-between border-t pt-3" style={{ borderColor: "var(--orbit-border)" }}>
        <span className="text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
          Total Direncanakan
        </span>
        <span className="text-base font-semibold" style={{ color: "var(--orbit-ink)" }}>
          {formatRupiah(totalPlanned)}
        </span>
      </div>

      <p className="text-center text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
        Ini adalah rencanamu, bukan batasanmu.
      </p>

      {error ? <p className="text-sm" style={{ color: "#b3261e" }}>{error}</p> : null}

      <Button
        variant="primary"
        size="lg"
        fullWidth
        isDisabled={saving}
        onPress={submit}
        className="!bg-[var(--orbit-accent)]"
      >
        {saving ? "Menyimpan..." : "Mulai Hariku"}
      </Button>
    </div>
  );
}

// ---------- Screen 3: Main Dashboard ----------

function MainDashboard() {
  const { token } = useAuth();
  const [home, setHome] = useState<HomeData | null>(null);
  const [plan, setPlan] = useState<DailyPlan | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [error, setError] = useState("");

  const load = async () => {
    if (!token) return;
    try {
      const [homeData, planData, catData, timelineData] = await Promise.all([
        api.get<HomeData>("/home", token),
        api.get<DailyPlan>(`/plans/${todayStr()}`, token),
        api.get<{ categories: Category[] }>("/categories", token),
        api.get<{ events: TimelineEvent[] }>("/transactions", token),
      ]);
      setHome(homeData);
      setPlan(planData);
      setCategories(catData.categories);
      setEvents(timelineData.events);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal memuat data");
    }
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  if (error) return <p className="text-sm" style={{ color: "#b3261e" }}>{error}</p>;
  if (!home || !plan) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <Orbit className="animate-spin" size={26} color="var(--orbit-accent)" />
      </div>
    );
  }

  const categoryName = (id: string) => categories.find((c) => c.ID === id)?.Name ?? "Kategori";
  const actualFor = (categoryId: string) =>
    events
      .filter((e) => e.type === "transaction" && e.transaction?.CategoryID === categoryId && e.occurred_at.slice(0, 10) === home.date)
      .reduce((sum, e) => sum + (e.transaction?.Amount ?? 0), 0);

  const pct = home.plan > 0 ? Math.max(0, Math.min(100, (home.remaining / home.plan) * 100)) : 0;

  return (
    <div className="flex flex-col gap-5 pt-2">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold" style={{ color: "var(--orbit-ink)" }}>
          {greeting()}
        </h1>
      </div>

      <Card variant="secondary" className="!rounded-2xl">
        <Card.Content className="flex flex-col gap-4 py-5">
          <span className="text-xs font-medium uppercase tracking-wide" style={{ color: "var(--orbit-ink-muted)" }}>
            Rencana Hari Ini
          </span>
          <span className="text-2xl font-semibold" style={{ color: "var(--orbit-ink)" }}>
            {formatRupiah(home.plan)}
          </span>

          <div className="flex items-center gap-4">
            <ProgressCircle value={pct} size="lg" aria-label="Sisa anggaran hari ini">
              <div className="flex flex-col items-center justify-center text-center">
                <span className="text-[10px]" style={{ color: "var(--orbit-ink-muted)" }}>
                  Tersisa
                </span>
                <span className="text-sm font-semibold" style={{ color: "var(--orbit-accent)" }}>
                  {formatRupiah(home.remaining)}
                </span>
              </div>
            </ProgressCircle>
            <div className="flex flex-1 flex-col gap-1">
              <span className="text-sm" style={{ color: "var(--orbit-ink)" }}>
                Terpakai {formatRupiah(home.actual)}
              </span>
              <span className="text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
                Tersisa {formatRupiah(home.remaining)}
              </span>
            </div>
          </div>

          <div className="flex flex-col gap-2">
            {plan.Categories.map((pc) => {
              const actual = actualFor(pc.CategoryID);
              const catPct = pc.PlannedAmount > 0 ? Math.min(100, (actual / pc.PlannedAmount) * 100) : 0;
              return (
                <div key={pc.ID} className="flex flex-col gap-1">
                  <div className="flex items-center justify-between text-xs">
                    <span style={{ color: "var(--orbit-ink)" }}>{categoryName(pc.CategoryID)}</span>
                    <span style={{ color: "var(--orbit-ink-muted)" }}>
                      {formatRupiah(actual)} / {formatRupiah(pc.PlannedAmount)}
                    </span>
                  </div>
                  <ProgressBar value={catPct} size="sm" aria-label={categoryName(pc.CategoryID)} />
                </div>
              );
            })}
          </div>
        </Card.Content>
      </Card>

      <div className="grid grid-cols-3 gap-2">
        <ActionButton href="/catat" icon={Receipt} label="Catat Pengeluaran" />
        <ActionButton href="/mood" icon={Smile} label="Update Mood" />
        <ActionButton href="/rencana/sesuaikan" icon={SlidersHorizontal} label="Sesuaikan Rencana" />
      </div>

      <Card variant="default" className="!rounded-2xl">
        <Card.Content className="flex items-center justify-between py-4">
          <span className="text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
            Suasana Hati Saat Ini
          </span>
          <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
            {home.mood ? (MOOD_LABEL[home.mood] ?? home.mood) : "Belum dicatat"}
          </span>
        </Card.Content>
      </Card>

      <Link href="/bank">
        <Card variant="default" className="!rounded-2xl">
          <Card.Content className="flex items-center gap-3 py-4">
            <PiggyBank size={22} color="var(--orbit-sage)" />
            <div className="flex flex-1 flex-col">
              <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
                Bank Konsumsi
              </span>
              <span className="text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
                Tersimpan dari rencana
              </span>
            </div>
            <span className="text-sm font-semibold" style={{ color: "var(--orbit-sage)" }}>
              {formatRupiah(home.consumption_bank.balance)}
            </span>
          </Card.Content>
        </Card>
      </Link>
    </div>
  );
}

function ActionButton({
  href,
  icon: Icon,
  label,
}: {
  href: string;
  icon: typeof Receipt;
  label: string;
}) {
  return (
    <Link
      href={href}
      className="flex !h-auto flex-col items-center justify-center gap-1.5 rounded-2xl border !py-4 text-center transition-colors"
      style={{ borderColor: "var(--orbit-border)", backgroundColor: "var(--orbit-surface)" }}
    >
      <Icon size={20} color="var(--orbit-accent)" />
      <span className="text-[11px] font-medium leading-tight" style={{ color: "var(--orbit-ink)" }}>
        {label}
      </span>
    </Link>
  );
}
