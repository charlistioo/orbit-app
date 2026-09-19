import { useEffect, useState } from 'react';
import { StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import {
  colors,
  currentTierFloor,
  nextBadgeTier,
  radius,
  spacing,
  typography,
} from './theme';
import { Card, DonutProgress, Icon, IconCircle, PrimaryButton, ProgressBar } from './ui';
import { badgeVisual } from './BadgeDisplay';
import { categoryEmoji } from './categoryEmoji';

const API_BASE = 'http://localhost:8080/api/v1';

const MOOD_LABEL: Record<string, string> = {
  very_good: 'Bersemangat',
  good: 'Senang',
  neutral: 'Santai & Tenang',
  stressed: 'Lelah',
  sad: 'Cemas',
};
const MOOD_EMOJI: Record<string, string> = {
  very_good: '🤩',
  good: '😊',
  neutral: '🌿',
  stressed: '🥱',
  sad: '🌧️',
};

type HomeData = {
  date: string;
  baseline: number;
  plan: number;
  actual: number;
  remaining: number;
  consumption_bank: { balance: number; badge: string };
  mood: string | null;
};
type BankData = { balance: number; ledger: { delta_amount: number; created_at: string; reason: string }[] };
type Category = { ID: string; Name: string };
type PlanCategory = { ID: string; CategoryID: string; PlannedAmount: number };
type TimelineEvent = {
  type: string;
  occurred_at: string;
  transaction?: { CategoryID: string; Amount: number };
  plan_adjustment?: { CategoryID: string; OldAmount: number; NewAmount: number };
  mood?: { Mood: string };
};

type Props = {
  sessionToken: string;
  onNavigateToCatat: () => void;
  onNavigateToTimeline: () => void;
};

function formatRupiah(amount: number): string {
  return 'Rp' + Math.round(amount).toLocaleString('id-ID');
}
function formatK(amount: number): string {
  return Math.round(amount / 1000) + 'k';
}
function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
}
function todayStr(): string {
  return new Date().toISOString().slice(0, 10);
}

// Home screen (TASK-015 restyle) - every number here still comes
// straight from GET /api/v1/home, /plans/{date}, /consumption-bank and
// /transactions (Timeline, filtered client-side to today) - this
// component only combines and formats what those endpoints already
// computed server-side, per TASK-009's "no client aggregation" design.
export default function HomeScreen({ sessionToken, onNavigateToCatat, onNavigateToTimeline }: Props) {
  const [home, setHome] = useState<HomeData | null>(null);
  const [bank, setBank] = useState<BankData | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [planCategories, setPlanCategories] = useState<PlanCategory[]>([]);
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [error, setError] = useState('');

  const authHeaders = { Authorization: `Bearer ${sessionToken}` };

  useEffect(() => {
    const today = todayStr();
    Promise.all([
      fetch(`${API_BASE}/home`, { headers: authHeaders }).then((r) => (r.ok ? r.json() : Promise.reject(r.status))),
      fetch(`${API_BASE}/consumption-bank`, { headers: authHeaders }).then((r) => (r.ok ? r.json() : Promise.reject(r.status))),
      fetch(`${API_BASE}/categories`, { headers: authHeaders }).then((r) => (r.ok ? r.json() : { categories: [] })),
      fetch(`${API_BASE}/plans/${today}`, { headers: authHeaders }).then((r) => (r.ok ? r.json() : { categories: [] })),
      fetch(`${API_BASE}/transactions`, { headers: authHeaders }).then((r) => (r.ok ? r.json() : { events: [] })),
    ])
      .then(([homeData, bankData, catData, planData, timelineData]) => {
        setHome(homeData);
        setBank(bankData);
        setCategories(catData.categories ?? []);
        setPlanCategories(planData.categories ?? []);
        setEvents(timelineData.events ?? []);
      })
      .catch((code) => setError(`Gagal memuat Home (kode ${code})`));
  }, []);

  if (error) return <Card><Text style={{ color: colors.error }}>{error}</Text></Card>;
  if (!home) return <Card><Text style={typography.bodyMd}>Memuat...</Text></Card>;

  const today = todayStr();
  const categoryName = (id: string) => categories.find((c) => c.ID === id)?.Name ?? 'Kategori';

  // Latest mood + when it was logged, from Timeline (Home's own
  // endpoint only reports the mood value, not its timestamp).
  const lastMoodEvent = [...events].reverse().find((e) => e.type === 'mood');

  // Per-category actual spend today, computed by filtering the
  // already-fetched Timeline transactions by date and category_id.
  const actualByCategory = (categoryId: string) =>
    events
      .filter((e) => e.type === 'transaction' && e.transaction?.CategoryID === categoryId && e.occurred_at.slice(0, 10) === today)
      .reduce((sum, e) => sum + (e.transaction?.Amount ?? 0), 0);

  // Today's plan-adjustment boosts per category (sum of new-old for
  // adjustments made today), for the small "+RpXXk" pill.
  const boostByCategory = (categoryId: string) =>
    events
      .filter((e) => e.type === 'plan_adjustment' && e.plan_adjustment?.CategoryID === categoryId && e.occurred_at.slice(0, 10) === today)
      .reduce((sum, e) => sum + ((e.plan_adjustment?.NewAmount ?? 0) - (e.plan_adjustment?.OldAmount ?? 0)), 0);

  const remainingProgress = home.plan > 0 ? Math.max(0, home.remaining) / home.plan : 0;
  const isHarmonious = home.remaining >= 0;

  const balance = bank?.balance ?? home.consumption_bank.balance;
  const badge = badgeVisual(home.consumption_bank.badge as any);
  const floor = currentTierFloor(balance);
  const next = nextBadgeTier(balance);
  const tierProgress = next ? (balance - floor) / (next.threshold - floor) : 1;

  // Balance ~7 days ago, from the real ledger history, for "Tumbuh
  // +RpX dari pekan lalu" - only shown when there's enough ledger
  // history to compute it honestly.
  const weekAgoISO = new Date(Date.now() - 7 * 86400000).toISOString();
  const ledgerBeforeWeekAgo = bank?.ledger.filter((l) => l.created_at < weekAgoISO) ?? [];
  const balanceWeekAgo = ledgerBeforeWeekAgo.length > 0 ? 0 + ledgerBeforeWeekAgo.reduce((s, l) => s + l.delta_amount, 0) : null;
  const weeklyDelta = balanceWeekAgo !== null ? balance - balanceWeekAgo : null;

  return (
    <>
      <View style={styles.topRow}>
        <View style={styles.dateChip}>
          <Icon name="calendar-today" size={14} color={colors.onSurfaceVariant} />
          <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
            Hari ini, {new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long' })}
          </Text>
        </View>
        {lastMoodEvent?.mood && (
          <View style={styles.moodChip}>
            <Text style={{ fontSize: 13 }}>{MOOD_EMOJI[lastMoodEvent.mood.Mood] ?? '🌿'}</Text>
            <Text style={[typography.labelSm, { color: colors.onSecondaryContainer }]}>
              {MOOD_LABEL[lastMoodEvent.mood.Mood] ?? lastMoodEvent.mood.Mood} {formatTime(lastMoodEvent.occurred_at)}
            </Text>
          </View>
        )}
      </View>

      <Card style={styles.heroCard}>
        <View style={styles.heroLabel}>
          <View style={styles.dotSmall} />
          <Text style={[typography.labelMd, { color: colors.onSurfaceVariant, textTransform: 'uppercase' }]}>
            {isHarmonious ? 'Ritme Keuangan Harmonis' : 'Ritme Keuangan Dinamis'}
          </Text>
        </View>

        <DonutProgress progress={remainingProgress}>
          <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>Sisa Hari Ini</Text>
          <Text style={[typography.headlineLg, { color: colors.onSurface }]}>{formatRupiah(home.remaining)}</Text>
          <Text style={[typography.labelSm, { color: colors.secondary }]}>
            {home.plan > 0 ? `${Math.round(remainingProgress * 100)}% Tersisa` : 'Belum ada rencana'}
          </Text>
        </DonutProgress>

        <Text style={[typography.bodySm, { color: colors.onSurfaceVariant, textAlign: 'center' }]}>
          {isHarmonious ? 'Pengeluaranmu mengalir wajar hari ini.' : 'Pengeluaran hari ini melampaui rencana - itu tetap informasi, bukan kegagalan.'}
        </Text>

        <View style={styles.statRow}>
          <StatChip label="Baseline" value={formatK(home.baseline)} caption="Patokan seimbang" />
          <StatChip label="Rencana" value={formatK(home.plan)} caption="Alokasi pos" />
          <StatChip label="Realisasi" value={formatK(home.actual)} caption="Tercatat teratur" />
        </View>
      </Card>

      <Card>
        <View style={styles.bankHeader}>
          <IconCircle icon="eco" bg={colors.secondaryFixed} color={colors.onSecondaryFixedVariant} />
          <View style={{ flex: 1 }}>
            <View style={styles.bankTitleRow}>
              <Text style={typography.headlineSm}>Consumption Bank</Text>
              <View style={[styles.badgePill, { backgroundColor: badge.bg }]}>
                <Text style={[typography.labelSm, { color: badge.fg }]}>{badge.label}</Text>
              </View>
            </View>
            <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>Bantalan kenyamanan belanja</Text>
          </View>
        </View>

        <View style={styles.bankAmountRow}>
          <View>
            <Text style={[typography.labelMd, { color: colors.onSurfaceVariant, textTransform: 'uppercase' }]}>
              Total Cadangan Underspend
            </Text>
            <Text style={[typography.currencyDisplay, { color: colors.onSurface }]}>{formatRupiah(balance)}</Text>
          </View>
          {weeklyDelta !== null && weeklyDelta !== 0 && (
            <Text style={[typography.labelMd, { color: weeklyDelta > 0 ? colors.secondary : colors.tertiary }]}>
              {weeklyDelta > 0 ? 'Tumbuh ' : 'Berkurang '}
              {weeklyDelta > 0 ? '+' : ''}
              {formatRupiah(weeklyDelta)}{'\n'}
              <Text style={{ color: colors.onSurfaceVariant }}>dari pekan lalu</Text>
            </Text>
          )}
        </View>

        <ProgressBar progress={tierProgress} color={colors.secondary} />
        <View style={styles.bankThresholdRow}>
          <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>Ambang: {formatRupiah(floor)}</Text>
          <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>
            {next ? `Menuju: ${next.name} (${formatRupiah(next.threshold)})` : 'Tier tertinggi tercapai'}
          </Text>
        </View>

        <View style={styles.infoBox}>
          <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
            Akumulasi underspend siap jadi bantalan pengeluaranmu (maks. 10% saldo per transaksi).
          </Text>
        </View>

        <TouchableOpacity style={styles.bufferLink} onPress={onNavigateToCatat} activeOpacity={0.7}>
          <Text style={[typography.labelMd, { color: colors.primaryContainer }]}>Gunakan buffer untuk transaksi berikutnya</Text>
          <Icon name="arrow-forward" size={16} color={colors.primaryContainer} />
        </TouchableOpacity>
      </Card>

      <View style={styles.sectionHeader}>
        <Text style={typography.headlineSm}>Rencana Kategori</Text>
        <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>{planCategories.length} Pos Aktif</Text>
      </View>

      {planCategories.length === 0 ? (
        <Card>
          <Text style={[typography.bodyMd, { color: colors.onSurfaceVariant }]}>Belum ada rencana kategori untuk hari ini.</Text>
        </Card>
      ) : (
        <View style={{ width: '100%', gap: spacing.sm }}>
          {planCategories.map((pc) => {
            const name = categoryName(pc.CategoryID);
            const actual = actualByCategory(pc.CategoryID);
            const boost = boostByCategory(pc.CategoryID);
            const over = actual > pc.PlannedAmount;
            const progress = pc.PlannedAmount > 0 ? actual / pc.PlannedAmount : 0;
            return (
              <Card key={pc.ID} style={styles.categoryRow}>
                <View style={styles.categoryTop}>
                  <View style={styles.categoryLeft}>
                    <IconCircle emoji={categoryEmoji(name)} size={36} />
                    <View>
                      <View style={{ flexDirection: 'row', alignItems: 'center', gap: 6 }}>
                        <Text style={typography.labelLg}>{name}</Text>
                        {boost > 0 && (
                          <View style={styles.boostPill}>
                            <Text style={[typography.labelSm, { color: colors.tertiary }]}>+{formatK(boost)}</Text>
                          </View>
                        )}
                      </View>
                      <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
                        Rencana: {formatRupiah(pc.PlannedAmount)}
                      </Text>
                    </View>
                  </View>
                  <View style={{ alignItems: 'flex-end' }}>
                    <Text style={typography.currencyMd}>{formatRupiah(actual)}</Text>
                    <Text style={[typography.labelSm, { color: actual === 0 ? colors.onSurfaceVariant : over ? colors.tertiary : colors.secondary }]}>
                      {actual === 0 ? 'Belum tercatat' : over ? 'Bisa ditukar buffer' : `Sisa ${formatRupiah(pc.PlannedAmount - actual)}`}
                    </Text>
                  </View>
                </View>
                <ProgressBar progress={progress} color={over ? colors.tertiaryFixedDim : colors.secondary} />
              </Card>
            );
          })}
        </View>
      )}

      <Card style={styles.encourageCard}>
        <IconCircle icon="spa" bg={colors.primaryFixed} color={colors.primaryContainer} />
        <View style={{ flex: 1 }}>
          <Text style={typography.labelLg}>Catat saat santai, nikmati prosesnya</Text>
          <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
            Menyadari pilihan hari ini adalah bentuk apresiasi diri.
          </Text>
        </View>
      </Card>

      <PrimaryButton title="Catat Pengeluaran Baru" onPress={onNavigateToCatat} icon="add" />
      <TouchableOpacity onPress={onNavigateToTimeline} style={{ alignSelf: 'center' }}>
        <Text style={[typography.labelMd, { color: colors.primaryContainer }]}>Lihat Rincian Timeline Hari Ini</Text>
      </TouchableOpacity>
    </>
  );
}

function StatChip({ label, value, caption }: { label: string; value: string; caption: string }) {
  return (
    <View style={styles.statChip}>
      <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>{label}</Text>
      <Text style={typography.headlineSm}>{value}</Text>
      <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>{caption}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  topRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', width: '100%', flexWrap: 'wrap', gap: spacing.xs },
  dateChip: { flexDirection: 'row', alignItems: 'center', gap: 6 },
  moodChip: {
    flexDirection: 'row', alignItems: 'center', gap: 6,
    backgroundColor: colors.secondaryContainer, borderRadius: radius.full, paddingHorizontal: 10, paddingVertical: 5,
  },
  heroCard: { alignItems: 'center', gap: spacing.md },
  heroLabel: { flexDirection: 'row', alignItems: 'center', gap: 6, alignSelf: 'flex-start' },
  dotSmall: { width: 6, height: 6, borderRadius: 3, backgroundColor: colors.secondary },
  statRow: { flexDirection: 'row', width: '100%', gap: spacing.xs },
  statChip: { flex: 1, backgroundColor: colors.surfaceContainerLow, borderRadius: radius.DEFAULT, padding: spacing.sm, alignItems: 'center', gap: 2 },
  bankHeader: { flexDirection: 'row', alignItems: 'flex-start', gap: spacing.sm },
  bankTitleRow: { flexDirection: 'row', alignItems: 'center', gap: spacing.xs, flexWrap: 'wrap' },
  badgePill: { borderRadius: radius.full, paddingHorizontal: 10, paddingVertical: 3 },
  bankAmountRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'flex-start' },
  bankThresholdRow: { flexDirection: 'row', justifyContent: 'space-between' },
  infoBox: { backgroundColor: colors.surfaceContainerLow, borderRadius: radius.DEFAULT, padding: spacing.sm },
  bufferLink: { flexDirection: 'row', alignItems: 'center', gap: 6, alignSelf: 'flex-start' },
  sectionHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', width: '100%' },
  categoryRow: { gap: spacing.xs },
  categoryTop: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  categoryLeft: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm, flexShrink: 1 },
  boostPill: { backgroundColor: colors.tertiaryFixed, borderRadius: radius.full, paddingHorizontal: 6, paddingVertical: 1 },
  encourageCard: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm },
});
