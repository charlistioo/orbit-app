"use client";

import { useEffect, useRef, useState } from "react";
import { User, LogOut, Camera, Pencil, Check } from "lucide-react";
import { Card, Button, Input } from "@heroui/react";
import { useAuth } from "@/lib/auth";
import { api, ApiError, BADGE_TIERS, Profile, formatRupiah } from "@/lib/api";

const AVATAR_MAX_DIM = 256;

function badgeMeta(name: string): { color: string } {
  switch (name) {
    case "Frugal":
      return { color: "#7C9A6E" };
    case "Thrifty":
      return { color: "#5E8B6E" };
    case "Economical":
      return { color: "#4A7C7C" };
    case "Collector":
      return { color: "#3D6B8A" };
    case "Master":
      return { color: "var(--orbit-accent)" };
    default:
      return { color: "var(--orbit-ink-muted)" };
  }
}

// Resizes/compresses an image client-side before it's ever sent to the
// backend - avatar_data is stored as base64 directly on the users row
// (no file-storage service, per techstack-002's risk mitigation), so
// keeping the payload small here is what keeps that safe.
function resizeImage(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error("Gagal membaca file"));
    reader.onload = () => {
      const img = new Image();
      img.onerror = () => reject(new Error("File bukan gambar yang valid"));
      img.onload = () => {
        const scale = Math.min(1, AVATAR_MAX_DIM / Math.max(img.width, img.height));
        const w = Math.round(img.width * scale);
        const h = Math.round(img.height * scale);
        const canvas = document.createElement("canvas");
        canvas.width = w;
        canvas.height = h;
        const ctx = canvas.getContext("2d");
        if (!ctx) return reject(new Error("Canvas tidak didukung"));
        ctx.drawImage(img, 0, 0, w, h);
        resolve(canvas.toDataURL("image/jpeg", 0.85));
      };
      img.src = reader.result as string;
    };
    reader.readAsDataURL(file);
  });
}

export default function ProfilPage() {
  const { token, signOut } = useAuth();
  const [profile, setProfile] = useState<Profile | null>(null);
  const [error, setError] = useState("");
  const [uploading, setUploading] = useState(false);
  const [editingBudget, setEditingBudget] = useState(false);
  const [budgetText, setBudgetText] = useState("");
  const [savingBudget, setSavingBudget] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const load = () => {
    if (!token) return;
    api.get<Profile>("/me/profile", token).then(setProfile).catch(() => setError("Gagal memuat profil"));
  };

  useEffect(load, [token]);

  const onPickAvatar = async (file: File | undefined) => {
    if (!file) return;
    setError("");
    setUploading(true);
    try {
      const dataUrl = await resizeImage(file);
      const updated = await api.patch<Profile>("/me/profile", token, { avatar_data: dataUrl });
      setProfile(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal mengunggah foto");
    } finally {
      setUploading(false);
    }
  };

  const startEditBudget = () => {
    setBudgetText(profile?.default_baseline_amount ? String(Math.round(profile.default_baseline_amount)) : "");
    setEditingBudget(true);
  };

  const saveBudget = async () => {
    const value = parseInt(budgetText, 10);
    if (!value || value <= 0) return setError("Isi budget harian yang valid");
    setError("");
    setSavingBudget(true);
    try {
      const updated = await api.patch<Profile>("/me/profile", token, { default_baseline_amount: value });
      setProfile(updated);
      setEditingBudget(false);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal menyimpan budget");
    } finally {
      setSavingBudget(false);
    }
  };

  if (error && !profile) return <p className="pt-2 text-sm" style={{ color: "#b3261e" }}>{error}</p>;
  if (!profile) return null;

  return (
    <div className="flex flex-col gap-5 pt-2">
      <div className="flex items-center gap-2">
        <User size={20} color="var(--orbit-accent)" />
        <h1 className="text-lg font-semibold" style={{ color: "var(--orbit-ink)" }}>
          Profil
        </h1>
      </div>

      <Card variant="secondary" className="!rounded-2xl">
        <Card.Content className="flex flex-col items-center gap-3 py-6">
          <button
            onClick={() => fileInputRef.current?.click()}
            className="relative flex h-24 w-24 items-center justify-center overflow-hidden rounded-full"
            style={{ backgroundColor: "var(--orbit-surface)", border: "2px solid var(--orbit-border)" }}
            aria-label="Ganti foto profil"
          >
            {profile.avatar_data ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={profile.avatar_data} alt="" className="h-full w-full object-cover" />
            ) : (
              <User size={36} color="var(--orbit-ink-muted)" />
            )}
            <div
              className="absolute bottom-0 flex w-full items-center justify-center py-1"
              style={{ backgroundColor: "rgba(0,0,0,0.45)" }}
            >
              <Camera size={14} color="#fff" />
            </div>
          </button>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            className="hidden"
            onChange={(e) => onPickAvatar(e.target.files?.[0])}
          />
          {uploading ? (
            <span className="text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
              Mengunggah...
            </span>
          ) : null}

          <div className="flex flex-col items-center gap-0.5">
            <span className="text-base font-medium" style={{ color: "var(--orbit-ink)" }}>
              {profile.display_name}
            </span>
            <span className="text-sm" style={{ color: "var(--orbit-ink-muted)" }}>
              {profile.email}
            </span>
            {profile.age ? (
              <span className="text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
                {profile.age} tahun
              </span>
            ) : null}
          </div>
        </Card.Content>
      </Card>

      <div className="flex flex-col gap-2">
        <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
          Badge Consumption Bank
        </span>
        <div className="grid grid-cols-5 gap-2">
          {BADGE_TIERS.map((t) => {
            const active = profile.badge === t.name;
            const meta = badgeMeta(t.name);
            return (
              <div key={t.name} className="flex flex-col items-center gap-1">
                <div
                  className="flex h-11 w-11 items-center justify-center rounded-full"
                  style={{
                    backgroundColor: active ? meta.color : "var(--orbit-accent-soft)",
                    opacity: active ? 1 : 0.5,
                  }}
                >
                  <span className="text-[10px] font-bold" style={{ color: active ? "#fff" : "var(--orbit-ink-muted)" }}>
                    {t.name[0]}
                  </span>
                </div>
                <span
                  className="text-center text-[10px] leading-tight"
                  style={{ color: active ? "var(--orbit-ink)" : "var(--orbit-ink-muted)", fontWeight: active ? 600 : 400 }}
                >
                  {t.name}
                </span>
              </div>
            );
          })}
        </div>
        {!profile.badge && (
          <p className="text-center text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
            Belum mencapai tier Frugal (Rp100.000) - terus kumpulkan sisa rencana.
          </p>
        )}
      </div>

      <Card variant="default" className="!rounded-2xl">
        <Card.Content className="flex flex-col gap-2 py-4">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium" style={{ color: "var(--orbit-ink)" }}>
              Budget Harian Default
            </span>
            {!editingBudget && (
              <button onClick={startEditBudget} aria-label="Ubah budget harian">
                <Pencil size={15} color="var(--orbit-accent)" />
              </button>
            )}
          </div>
          {editingBudget ? (
            <div className="flex items-center gap-2">
              <Input
                inputMode="numeric"
                fullWidth
                value={budgetText}
                onChange={(e) => setBudgetText(e.target.value.replace(/\D/g, ""))}
              />
              <Button variant="secondary" isIconOnly isDisabled={savingBudget} onPress={saveBudget} aria-label="Simpan">
                <Check size={16} />
              </Button>
            </div>
          ) : (
            <span className="text-lg font-semibold" style={{ color: "var(--orbit-ink)" }}>
              {profile.default_baseline_amount ? formatRupiah(profile.default_baseline_amount) : "Belum diatur"}
            </span>
          )}
          <p className="text-xs" style={{ color: "var(--orbit-ink-muted)" }}>
            Dipakai sebagai saran awal saat merencanakan hari baru - tidak mengunci pengeluaranmu.
          </p>
        </Card.Content>
      </Card>

      {error ? <p className="text-sm" style={{ color: "#b3261e" }}>{error}</p> : null}

      <Button variant="outline" fullWidth onPress={signOut} className="flex items-center gap-2">
        <LogOut size={16} />
        Keluar
      </Button>
    </div>
  );
}
