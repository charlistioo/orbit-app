// Thin fetch wrapper for the ORBIT backend (Go, unchanged by the web
// pivot - every endpoint here already existed before TASK-016). Every
// call attaches the session token as a bearer header; callers get back
// parsed JSON or throw ApiError with the backend's own message.
const API_BASE = process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080/api/v1";

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(
  path: string,
  token: string | null,
  options: RequestInit = {}
): Promise<T> {
  const headers: Record<string, string> = {
    ...(options.body ? { "Content-Type": "application/json" } : {}),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(res.status, body.error ?? `Request failed (${res.status})`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  get: <T>(path: string, token: string | null) => request<T>(path, token),
  post: <T>(path: string, token: string | null, body?: unknown) =>
    request<T>(path, token, { method: "POST", body: body ? JSON.stringify(body) : undefined }),
  patch: <T>(path: string, token: string | null, body?: unknown) =>
    request<T>(path, token, { method: "PATCH", body: body ? JSON.stringify(body) : undefined }),
};

export type SignedInUser = {
  id: string;
  email: string;
  display_name: string;
};

export type MeResponse = {
  user_id: string;
  onboarding_completed: boolean;
};

export type Profile = {
  id: string;
  email: string;
  display_name: string;
  age: number | null;
  avatar_data: string | null;
  default_baseline_amount: number | null;
  badge: string;
  onboarding_completed: boolean;
};

export type Category = { ID: string; Name: string };
export type Item = { ID: string; CategoryID: string; Name: string };

export type PlanCategory = { ID: string; CategoryID: string; PlannedAmount: number };
export type DailyPlan = {
  ID: string;
  PlanDate: string;
  BaselineAmount: number;
  Categories: PlanCategory[];
};

export type HomeData = {
  date: string;
  baseline: number;
  plan: number;
  actual: number;
  remaining: number;
  consumption_bank: { balance: number; badge: string };
  mood: string | null;
};

export type Transaction = {
  ID: string;
  CategoryID: string;
  ItemID: string | null;
  PlanCategoryID: string | null;
  Amount: number;
  PlanAmountSnapshot: number | null;
  OccurredAt: string;
};

export type TimelineEvent = {
  type: "transaction" | "plan_adjustment" | "mood" | "bank_ledger";
  occurred_at: string;
  transaction?: {
    CategoryID: string;
    CategoryName: string;
    ItemID: string | null;
    ItemName: string | null;
    Amount: number;
    PlanAmountSnapshot: number | null;
  };
  plan_adjustment?: {
    CategoryID: string;
    CategoryName: string;
    OldAmount: number;
    NewAmount: number;
  };
  mood?: { Mood: string };
  bank_ledger?: {
    DeltaAmount: number;
    Reason: string;
    RelatedTransactionID: string | null;
    BalanceAfter: number;
  };
};

export type MoodEntry = { ID: string; Mood: string; RecordedAt: string };

export type LedgerEntry = {
  ID: string;
  DeltaAmount: number;
  Reason: string;
  RelatedTransactionID: string | null;
  BalanceAfter: number;
  CreatedAt: string;
};

export type BankData = { balance: number; badge: string; ledger: LedgerEntry[] };

export type InsightResult = { period: string; date: string; insights: string[] };

export function todayStr(): string {
  return new Date().toISOString().slice(0, 10);
}

export function formatRupiah(amount: number): string {
  return "Rp" + Math.round(amount).toLocaleString("id-ID");
}

// Mirrors backend badgeTierFor (TASK-007) - client-side only used to
// compute "how far to the next tier"; the badge itself always comes
// from the server's own `badge` field, never recomputed here.
export const BADGE_TIERS = [
  { name: "Frugal", threshold: 100000 },
  { name: "Thrifty", threshold: 200000 },
  { name: "Economical", threshold: 300000 },
  { name: "Collector", threshold: 400000 },
  { name: "Master", threshold: 500000 },
];

export function nextBadgeTier(balance: number) {
  return BADGE_TIERS.find((t) => balance < t.threshold) ?? null;
}

export function currentTierFloor(balance: number): number {
  const reached = [...BADGE_TIERS].reverse().find((t) => balance >= t.threshold);
  return reached ? reached.threshold : 0;
}
