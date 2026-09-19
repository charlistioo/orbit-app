"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { X } from "lucide-react";
import { Card, Button, Input, buttonVariants } from "@heroui/react";
import { useAuth } from "@/lib/auth";
import { api, ApiError, Category, Item, Transaction, TimelineEvent, todayStr } from "@/lib/api";
import Link from "next/link";

// Screen 4 (Record Consumption) - quick-select categories/items exactly
// as typed (TASK-004's standing no-fuzzy-matching rule), no amount
// minimum. On save, routes to the result screen (screen 5) carrying:
// - the category's CUMULATIVE actual for today (not just this one
//   transaction - a second purchase in the same category must show the
//   running total, matching what "Direncanakan" is compared against)
// - the REAL Consumption Bank delta this transaction caused, read back
//   from the balance before/after the save rather than recomputed here
//   (planned - actual would double count when a category already had
//   spend earlier today, which is exactly the bug this fixes)
export default function CatatPage() {
  const { token } = useAuth();
  const router = useRouter();
  const [categories, setCategories] = useState<Category[]>([]);
  const [selected, setSelected] = useState<Category | null>(null);
  const [items, setItems] = useState<Item[]>([]);
  const [itemName, setItemName] = useState("");
  const [amount, setAmount] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!token) return;
    api.get<{ categories: Category[] }>("/categories", token).then((d) => {
      setCategories(d.categories);
      if (d.categories.length > 0) selectCategory(d.categories[0]);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  const selectCategory = async (c: Category) => {
    setSelected(c);
    setItems([]);
    const d = await api.get<{ items: Item[] }>(`/categories/${c.ID}/items`, token);
    setItems(d.items);
  };

  const save = async () => {
    const value = parseInt(amount, 10);
    if (!selected || !value) return setError("Pilih kategori dan isi harga");
    setError("");
    setSaving(true);
    try {
      const balanceBefore = (await api.get<{ balance: number }>("/consumption-bank", token)).balance;

      let itemId: string | undefined;
      if (itemName.trim()) {
        const existing = items.find((i) => i.Name === itemName.trim());
        itemId = existing
          ? existing.ID
          : (await api.post<Item>(`/categories/${selected.ID}/items`, token, { name: itemName.trim() })).ID;
      }
      const tx = await api.post<Transaction>("/transactions", token, {
        category_id: selected.ID,
        item_id: itemId ?? null,
        amount: value,
      });

      const [timeline, bankAfter] = await Promise.all([
        api.get<{ events: TimelineEvent[] }>("/transactions", token),
        api.get<{ balance: number }>("/consumption-bank", token),
      ]);
      const today = todayStr();
      const cumulativeActual = timeline.events
        .filter((e) => e.type === "transaction" && e.transaction?.CategoryID === selected.ID && e.occurred_at.slice(0, 10) === today)
        .reduce((sum, e) => sum + (e.transaction?.Amount ?? 0), 0);
      const bankDelta = bankAfter.balance - balanceBefore;

      const params = new URLSearchParams({
        category: selected.Name,
        actual: String(cumulativeActual),
        planned: tx.PlanAmountSnapshot != null ? String(tx.PlanAmountSnapshot) : "",
        delta: String(bankDelta),
      });
      router.push(`/catat/hasil?${params.toString()}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal menyimpan");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col gap-5 pt-2">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold" style={{ color: "var(--orbit-ink)" }}>
          Apa yang kamu beli?
        </h1>
        <Link href="/" className={buttonVariants({ variant: "ghost", size: "sm", isIconOnly: true })} aria-label="Tutup">
          <X size={18} />
        </Link>
      </div>

      <Card variant="default" className="!rounded-2xl">
        <Card.Content className="flex flex-col gap-5 py-5">
          <div className="flex flex-col gap-2">
            <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
              Kategori
            </span>
            <div className="flex flex-wrap gap-2">
              {categories.map((c) => (
                <button
                  key={c.ID}
                  onClick={() => selectCategory(c)}
                  className="rounded-full px-3.5 py-1.5 text-sm transition-colors"
                  style={{
                    backgroundColor: selected?.ID === c.ID ? "var(--orbit-accent)" : "var(--orbit-accent-soft)",
                    color: selected?.ID === c.ID ? "#fff" : "var(--orbit-ink)",
                  }}
                >
                  {c.Name}
                </button>
              ))}
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
              Nama Item
            </span>
            <Input fullWidth placeholder="Contoh: Es Teh Manis" value={itemName} onChange={(e) => setItemName(e.target.value)} />
          </div>

          <div className="flex flex-col gap-1.5">
            <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
              Harga
            </span>
            <Input
              inputMode="numeric"
              fullWidth
              placeholder="0"
              value={amount}
              onChange={(e) => setAmount(e.target.value.replace(/\D/g, ""))}
            />
          </div>

          {items.length > 0 && (
            <div className="flex flex-col gap-1.5">
              <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
                Item Terbaru
              </span>
              <div className="flex flex-wrap gap-2">
                {items.map((i) => (
                  <button
                    key={i.ID}
                    onClick={() => setItemName(i.Name)}
                    className="rounded-full px-3 py-1 text-xs"
                    style={{ backgroundColor: "var(--orbit-sage-soft)", color: "var(--orbit-ink)" }}
                  >
                    {i.Name}
                  </button>
                ))}
              </div>
            </div>
          )}
        </Card.Content>
      </Card>

      {error ? <p className="text-sm" style={{ color: "#b3261e" }}>{error}</p> : null}

      <Button
        variant="primary"
        size="lg"
        fullWidth
        isDisabled={saving}
        onPress={save}
        className="!bg-[var(--orbit-accent)]"
      >
        {saving ? "Menyimpan..." : "Simpan Pengeluaran"}
      </Button>
    </div>
  );
}
