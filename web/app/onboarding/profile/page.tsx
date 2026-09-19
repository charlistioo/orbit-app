"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { UserRound, Cake, Wallet } from "lucide-react";
import { Input, Button } from "@heroui/react";
import { useAuth } from "@/lib/auth";
import { api, Profile } from "@/lib/api";

// Screen 2a (Nama/Umur/Budget) - shown exactly once, right after first
// sign-in. Calls PATCH /me/profile, which also marks onboarding
// complete server-side (see backend/internal/store/users.go).
export default function OnboardingProfilePage() {
  const { token, refreshMe } = useAuth();
  const router = useRouter();
  const [name, setName] = useState("");
  const [age, setAge] = useState("");
  const [baseline, setBaseline] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  const submit = async () => {
    setError("");
    const parsedAge = parseInt(age, 10);
    const parsedBaseline = parseInt(baseline.replace(/\D/g, ""), 10);

    if (!name.trim()) return setError("Nama tidak boleh kosong");
    if (!parsedAge || parsedAge < 1 || parsedAge > 130) return setError("Umur tidak valid");
    if (!parsedBaseline || parsedBaseline <= 0) return setError("Budget harian tidak valid");

    setSaving(true);
    try {
      await api.patch<Profile>("/me/profile", token, {
        display_name: name.trim(),
        age: parsedAge,
        default_baseline_amount: parsedBaseline,
      });
      await refreshMe();
      router.replace("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal menyimpan profil");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex min-h-[calc(100vh-3rem)] flex-col justify-between gap-8 py-4">
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-1 text-center">
          <h1 className="text-xl font-semibold" style={{ color: "var(--orbit-ink)" }}>
            Lengkapi profilmu
          </h1>
          <p className="text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
            Sedikit info supaya ORBIT terasa lebih personal.
          </p>
        </div>

        <Field icon={UserRound} label="Nama">
          <Input
            placeholder="Nama panggilan"
            value={name}
            onChange={(e) => setName(e.target.value)}
            fullWidth
          />
        </Field>

        <Field icon={Cake} label="Umur">
          <Input
            type="number"
            placeholder="Contoh: 21"
            value={age}
            onChange={(e) => setAge(e.target.value)}
            fullWidth
          />
        </Field>

        <Field icon={Wallet} label="Budget batasan harian">
          <Input
            inputMode="numeric"
            placeholder="Contoh: 90000"
            value={baseline}
            onChange={(e) => setBaseline(e.target.value.replace(/\D/g, ""))}
            fullWidth
          />
          <p className="mt-1 text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
            Ini jadi patokan awal - bisa disesuaikan tiap hari, tidak mengunci pengeluaranmu.
          </p>
        </Field>

        {error ? <p className="text-sm" style={{ color: "#b3261e" }}>{error}</p> : null}
      </div>

      <Button
        variant="primary"
        size="lg"
        fullWidth
        isDisabled={saving}
        onPress={submit}
        className="!bg-[var(--orbit-accent)]"
      >
        {saving ? "Menyimpan..." : "Simpan & Lanjutkan"}
      </Button>
    </div>
  );
}

function Field({
  icon: Icon,
  label,
  children,
}: {
  icon: typeof UserRound;
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <span className="flex items-center gap-1.5 text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
        <Icon size={16} color="var(--orbit-sage)" />
        {label}
      </span>
      {children}
    </div>
  );
}
