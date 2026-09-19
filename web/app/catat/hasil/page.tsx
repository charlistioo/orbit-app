"use client";

import { Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { CheckCircle2 } from "lucide-react";
import { Card, Button } from "@heroui/react";
import { formatRupiah } from "@/lib/api";

// Screen 5 (Plan vs Actual result) - a neutral confirmation shown right
// after saving a transaction, never framed as pass/fail: underspending
// gets an encouraging note, overspending is shown as plain information
// (per the non-punitive design principle - overspend is never blocked
// or shamed).
//
// "Aktual" is the category's CUMULATIVE spend for today (not just this
// one transaction), and the +/- badge is the REAL Consumption Bank
// balance change this save caused (read back from the backend before/
// after, not recomputed as planned-actual here) - a second purchase in
// the same category must never look like it doubled the credit.
function ResultContent() {
  const params = useSearchParams();
  const router = useRouter();

  const category = params.get("category") ?? "Pengeluaran";
  const actual = Number(params.get("actual") ?? 0);
  const plannedRaw = params.get("planned");
  const planned = plannedRaw ? Number(plannedRaw) : null;
  const delta = Number(params.get("delta") ?? 0);

  return (
    <div className="flex min-h-[calc(100vh-3rem)] flex-col items-center justify-center gap-6 text-center">
      <div
        className="flex h-16 w-16 items-center justify-center rounded-full"
        style={{ backgroundColor: "var(--orbit-sage-soft)" }}
      >
        <CheckCircle2 size={32} color="var(--orbit-sage)" strokeWidth={1.8} />
      </div>

      <Card variant="secondary" className="w-full !rounded-2xl">
        <Card.Content className="flex flex-col gap-4 py-6">
          <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
            {category}
          </span>
          <div className="flex items-center justify-around">
            <div className="flex flex-col gap-0.5">
              <span className="text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
                Direncanakan
              </span>
              <span className="text-lg font-semibold" style={{ color: "var(--orbit-ink)" }}>
                {planned != null ? formatRupiah(planned) : "-"}
              </span>
            </div>
            <div className="h-8 w-px" style={{ backgroundColor: "var(--orbit-border)" }} />
            <div className="flex flex-col gap-0.5">
              <span className="text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
                Aktual hari ini
              </span>
              <span className="text-lg font-semibold" style={{ color: "var(--orbit-ink)" }}>
                {formatRupiah(actual)}
              </span>
            </div>
          </div>

          {delta !== 0 && (
            <div
              className="self-center rounded-full px-4 py-1.5 text-sm font-medium"
              style={{
                backgroundColor: delta > 0 ? "var(--orbit-sage-soft)" : "var(--orbit-accent-soft)",
                color: delta > 0 ? "var(--orbit-sage)" : "var(--orbit-accent)",
              }}
            >
              {delta > 0 ? "+" : ""}
              {formatRupiah(delta)} Bank Konsumsi
            </div>
          )}
        </Card.Content>
      </Card>

      <p className="max-w-xs text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
        {planned == null
          ? "Tercatat. Tidak ada rencana untuk kategori ini hari ini."
          : actual <= planned
            ? `Keren! Realisasi ${category.toLowerCase()} hari ini masih ${formatRupiah(planned - actual)} di bawah rencana.`
            : `Tercatat. Realisasi ${category.toLowerCase()} hari ini melebihi rencana - itu tetap informasi, bukan masalah.`}
      </p>

      <Button
        variant="primary"
        size="lg"
        fullWidth
        onPress={() => router.push("/")}
        className="!bg-[var(--orbit-accent)]"
      >
        Kembali ke Hari Ini
      </Button>
    </div>
  );
}

export default function CatatHasilPage() {
  return (
    <Suspense fallback={null}>
      <ResultContent />
    </Suspense>
  );
}
