"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { X, Smile } from "lucide-react";
import { Button, buttonVariants } from "@heroui/react";
import { useAuth } from "@/lib/auth";
import { api, ApiError } from "@/lib/api";
import Link from "next/link";

const MOODS: { value: string; label: string }[] = [
  { value: "very_good", label: "Sangat Baik" },
  { value: "good", label: "Baik" },
  { value: "neutral", label: "Biasa Saja" },
  { value: "stressed", label: "Stres" },
  { value: "sad", label: "Sedih" },
];

// Screen 6 (Update Mood) - entirely optional, each save is its own
// timestamped entry (never overwrites a previous one, TASK-008's
// standing rule). No causal language anywhere - mood is only ever
// shown elsewhere as a temporal note, never as a cause of spending.
export default function MoodPage() {
  const { token } = useAuth();
  const router = useRouter();
  const [selected, setSelected] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const save = async () => {
    if (!selected) return;
    setSaving(true);
    setError("");
    try {
      await api.post("/mood", token, { mood: selected });
      router.push("/");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal menyimpan mood");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col gap-6 pt-2">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold" style={{ color: "var(--orbit-ink)" }}>
          Bagaimana perasaanmu saat ini?
        </h1>
        <Link href="/" className={buttonVariants({ variant: "ghost", size: "sm", isIconOnly: true })} aria-label="Tutup">
          <X size={18} />
        </Link>
      </div>

      <div className="grid grid-cols-2 gap-3">
        {MOODS.map((m) => (
          <button
            key={m.value}
            onClick={() => setSelected(m.value)}
            className="flex flex-col items-center gap-2 rounded-2xl border py-5 transition-colors"
            style={{
              borderColor: selected === m.value ? "var(--orbit-accent)" : "var(--orbit-border)",
              backgroundColor: selected === m.value ? "var(--orbit-accent-soft)" : "var(--orbit-surface)",
            }}
          >
            <Smile size={26} color={selected === m.value ? "var(--orbit-accent)" : "var(--orbit-ink-muted)"} />
            <span
              className="text-sm font-medium"
              style={{ color: selected === m.value ? "var(--orbit-accent)" : "var(--orbit-ink)" }}
            >
              {m.label}
            </span>
          </button>
        ))}
      </div>

      <p className="text-center text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
        Opsional. Perbarui kapan pun kamu merasa ada perubahan.
      </p>

      {error ? <p className="text-center text-sm" style={{ color: "#b3261e" }}>{error}</p> : null}

      <Button
        variant="primary"
        size="lg"
        fullWidth
        isDisabled={!selected || saving}
        onPress={save}
        className="!bg-[var(--orbit-accent)]"
      >
        {saving ? "Menyimpan..." : "Simpan Mood"}
      </Button>
    </div>
  );
}
