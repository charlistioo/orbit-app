"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { X, Coffee } from "lucide-react";
import { Card, Button, Input, buttonVariants } from "@heroui/react";
import { useAuth } from "@/lib/auth";
import { api, ApiError, Category, DailyPlan, formatRupiah, todayStr } from "@/lib/api";
import Link from "next/link";

// Screen 7 (Adjust Plan) - edits one category's planned amount via the
// existing PATCH /plans/{plan_id}/categories/{plan_category_id}
// (unchanged since TASK-005); no maximum is enforced, plans are soft
// targets. Picking a category is step one, adjusting its amount is
// step two - both on this one screen.
function AdjustPlanContent() {
  const { token } = useAuth();
  const router = useRouter();
  const [plan, setPlan] = useState<DailyPlan | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [amount, setAmount] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!token) return;
    Promise.all([
      api.get<DailyPlan>(`/plans/${todayStr()}`, token),
      api.get<{ categories: Category[] }>("/categories", token),
    ]).then(([p, c]) => {
      setPlan(p);
      setCategories(c.categories);
    });
  }, [token]);

  const categoryName = (id: string) => categories.find((c) => c.ID === id)?.Name ?? "Kategori";
  const selected = plan?.Categories.find((pc) => pc.ID === selectedId) ?? null;

  const submit = async () => {
    if (!selected) return;
    const value = parseInt(amount, 10);
    if (!value || value <= 0) return setError("Isi jumlah rencana baru");
    setError("");
    setSaving(true);
    try {
      await api.patch(`/plans/${plan!.ID}/categories/${selected.ID}`, token, { planned_amount: value });
      router.push("/");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal menyimpan");
    } finally {
      setSaving(false);
    }
  };

  if (!plan) return null;

  const diff = selected && amount ? parseInt(amount, 10) - selected.PlannedAmount : 0;

  return (
    <div className="flex flex-col gap-5 pt-2">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold" style={{ color: "var(--orbit-ink)" }}>
          Sesuaikan Rencana
        </h1>
        <Link href="/" className={buttonVariants({ variant: "ghost", size: "sm", isIconOnly: true })} aria-label="Tutup">
          <X size={18} />
        </Link>
      </div>

      {!selected ? (
        <div className="flex flex-col gap-2">
          {plan.Categories.length === 0 ? (
            <p className="text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
              Belum ada kategori direncanakan hari ini.
            </p>
          ) : (
            plan.Categories.map((pc) => (
              <button
                key={pc.ID}
                onClick={() => {
                  setSelectedId(pc.ID);
                  setAmount(String(Math.round(pc.PlannedAmount)));
                }}
                className="flex items-center justify-between rounded-2xl border px-4 py-3 text-left"
                style={{ borderColor: "var(--orbit-border)" }}
              >
                <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
                  {categoryName(pc.CategoryID)}
                </span>
                <span className="text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
                  {formatRupiah(pc.PlannedAmount)}
                </span>
              </button>
            ))
          )}
        </div>
      ) : (
        <Card variant="secondary" className="!rounded-2xl">
          <Card.Content className="flex flex-col gap-3 py-5">
            <div className="flex items-center gap-2">
              <Coffee size={18} color="var(--orbit-accent)" />
              <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
                {categoryName(selected.CategoryID)}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span style={{ color: "var(--orbit-ink-muted)" }}>Rencana saat ini:</span>
              <span style={{ color: "var(--orbit-ink)" }}>{formatRupiah(selected.PlannedAmount)}</span>
            </div>
            <div className="flex flex-col gap-1.5">
              <span className="text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
                Jumlah baru:
              </span>
              <Input
                inputMode="numeric"
                fullWidth
                value={amount}
                onChange={(e) => setAmount(e.target.value.replace(/\D/g, ""))}
              />
            </div>
            <div className="flex justify-between text-sm">
              <span style={{ color: "var(--orbit-ink-muted)" }}>Selisih:</span>
              <span style={{ color: diff >= 0 ? "var(--orbit-sage)" : "var(--orbit-accent)" }}>
                {diff >= 0 ? "+" : ""}
                {formatRupiah(diff)}
              </span>
            </div>
          </Card.Content>
        </Card>
      )}

      <p className="text-center text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
        Rencana bisa berubah. Tidak apa-apa.
      </p>

      {error ? <p className="text-center text-sm" style={{ color: "#b3261e" }}>{error}</p> : null}

      {selected && (
        <>
          <Button
            variant="primary"
            size="lg"
            fullWidth
            isDisabled={saving}
            onPress={submit}
            className="!bg-[var(--orbit-accent)]"
          >
            {saving ? "Menyimpan..." : "Simpan Rencana Baru"}
          </Button>
          <Button variant="ghost" fullWidth onPress={() => setSelectedId(null)}>
            Batal
          </Button>
        </>
      )}
    </div>
  );
}

export default function AdjustPlanPage() {
  return (
    <Suspense fallback={null}>
      <AdjustPlanContent />
    </Suspense>
  );
}
