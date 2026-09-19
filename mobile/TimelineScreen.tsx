import { useEffect, useMemo, useState } from 'react';
import { StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { MaterialIcons } from '@expo/vector-icons';
import { BADGE_TIERS, colors, radius, spacing, typography } from './theme';
import { Card, ErrorState, Icon, IconCircle } from './ui';

const API_BASE = 'http://localhost:8080/api/v1';

const MOOD_LABEL: Record<string, string> = {
  very_good: 'Bersemangat',
  good: 'Senang',
  neutral: 'Tenang & Santai',
  stressed: 'Lelah',
  sad: 'Cemas',
};
const MOOD_EMOJI: Record<string, string> = {
  very_good: '🤩', good: '😊', neutral: '🌿', stressed: '🥱', sad: '🌧️',
};

type TimelineEvent = {
  type: 'transaction' | 'plan_adjustment' | 'mood' | 'bank_ledger';
  occurred_at: string;
  transaction?: { CategoryID: string; CategoryName: string; ItemID: string | null; ItemName: string | null; Amount: number; PlanAmountSnapshot: number | null };
  plan_adjustment?: { CategoryID: string; CategoryName: string; OldAmount: number; NewAmount: number };
  mood?: { Mood: string };
  bank_ledger?: { DeltaAmount: number; Reason: string; RelatedTransactionID: string | null; BalanceAfter: number };
};

type Props = { sessionToken: string };
type FilterKey = 'semua' | 'transaction' | 'mood' | 'plan_adjustment' | 'bank_ledger';

function formatRupiah(amount: number): string {
  return 'Rp' + Math.round(amount).toLocaleString('id-ID');
}
function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
}
function formatDateHeader(iso: string): string {
  return new Date(iso).toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long' }).toUpperCase();
}
function tierFor(balance: number): string {
  return [...BADGE_TIERS].reverse().find((t) => balance >= t.threshold)?.name ?? '';
}

// Renders the merged Timeline (per TASK-010's design, server-sorted -
// no client-side re-sorting). Date grouping, filter chips, and the
// per-event context notes (TASK-015 restyle) are display-only: the
// "X menit setelah mood Y" and "tier naik ke Z" notes are computed
// from the real, already-fetched event list itself (nearest preceding
// mood, previous ledger balance), never invented.
export default function TimelineScreen({ sessionToken }: Props) {
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState<FilterKey>('semua');
  const [todayOnly, setTodayOnly] = useState(false);

  useEffect(() => {
    fetch(`${API_BASE}/transactions`, { headers: { Authorization: `Bearer ${sessionToken}` } })
      .then((res) => {
        if (!res.ok) throw new Error(`backend returned ${res.status}`);
        return res.json();
      })
      .then((data) => setEvents(data.events ?? []))
      .catch((err) => setError(`Failed to load timeline: ${err.message}`));
  }, []);

  const chronological = [...events].reverse(); // most recent first
  const today = new Date().toISOString().slice(0, 10);

  const filtered = chronological.filter((e) => {
    if (filter !== 'semua' && e.type !== filter) return false;
    if (todayOnly && e.occurred_at.slice(0, 10) !== today) return false;
    return true;
  });

  const grouped = useMemo(() => {
    const groups: { date: string; events: TimelineEvent[] }[] = [];
    for (const e of filtered) {
      const date = e.occurred_at.slice(0, 10);
      let g = groups.find((g) => g.date === date);
      if (!g) { g = { date, events: [] }; groups.push(g); }
      g.events.push(e);
    }
    return groups;
  }, [filtered]);

  // Nearest mood logged before a transaction, within 90 minutes -
  // computed from the same already-fetched (chronological) list.
  const moodBefore = (tx: TimelineEvent) => {
    const idx = chronological.indexOf(tx);
    for (let i = idx + 1; i < chronological.length; i++) {
      const e = chronological[i];
      if (e.type === 'mood') {
        const minutes = (new Date(tx.occurred_at).getTime() - new Date(e.occurred_at).getTime()) / 60000;
        if (minutes >= 0 && minutes <= 90) return { mood: e.mood!.Mood, minutes: Math.round(minutes) };
        return null;
      }
    }
    return null;
  };

  // Previous ledger balance right before this one, to detect a tier
  // change (real - both balances come from the server's own ledger).
  const previousLedgerBalance = (entry: TimelineEvent) => {
    const idx = chronological.indexOf(entry);
    for (let i = idx + 1; i < chronological.length; i++) {
      if (chronological[i].type === 'bank_ledger') return chronological[i].bank_ledger!.BalanceAfter;
    }
    return 0;
  };

  if (error) return <ErrorState message={error} />;

  return (
    <>
      <Card>
        <View style={styles.headerRow}>
          <Icon name="swap-horiz" size={20} color={colors.primaryContainer} />
          <Text style={typography.headlineSm}>Timeline Aliran Hari Ini</Text>
        </View>
        <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
          Ruang netral untuk meninjau ritme pengeluaran dan catatan rasa tanpa penghakiman.
        </Text>

        <View style={styles.filterRow}>
          <FilterChip label={`Semua ${chronological.length}`} active={filter === 'semua'} onPress={() => setFilter('semua')} />
          <FilterChip label="Transaksi" emoji="💰" active={filter === 'transaction'} onPress={() => setFilter('transaction')} />
          <FilterChip label="Mood" emoji="🌿" active={filter === 'mood'} onPress={() => setFilter('mood')} />
          <FilterChip label="Plan" emoji="⚖️" active={filter === 'plan_adjustment'} onPress={() => setFilter('plan_adjustment')} />
        </View>

        <TouchableOpacity style={styles.dateFilterChip} onPress={() => setTodayOnly((v) => !v)} activeOpacity={0.7}>
          <Icon name="event" size={14} color={colors.onSurfaceVariant} />
          <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>{todayOnly ? 'Hari Ini' : 'Semua Tanggal'}</Text>
          <Icon name="expand-more" size={16} color={colors.onSurfaceVariant} />
        </TouchableOpacity>
      </Card>

      {grouped.length === 0 && (
        <Card><Text style={[typography.bodyMd, { color: colors.onSurfaceVariant }]}>Belum ada aktivitas.</Text></Card>
      )}

      {grouped.map((group) => (
        <View key={group.date} style={{ width: '100%', gap: spacing.sm }}>
          <View style={styles.dateHeaderRow}>
            <View style={styles.dateDot} />
            <Text style={[typography.labelMd, { color: colors.onSurfaceVariant }]}>{formatDateHeader(group.date)}</Text>
          </View>

          {group.events.map((e, i) => (
            <TimelineCard
              key={i}
              event={e}
              moodBefore={e.type === 'transaction' ? moodBefore(e) : null}
              tierNote={
                e.type === 'bank_ledger' && e.bank_ledger!.DeltaAmount > 0
                  ? (() => {
                      const prevTier = tierFor(previousLedgerBalance(e));
                      const newTier = tierFor(e.bank_ledger!.BalanceAfter);
                      return newTier && newTier !== prevTier ? newTier : null;
                    })()
                  : null
              }
            />
          ))}
        </View>
      ))}

      {grouped.length > 0 && (
        <View style={styles.footerNote}>
          <Icon name="verified" size={16} color={colors.onSurfaceVariant} />
          <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>
            Semua rekaman tersimpan apa adanya untuk kesadaranmu.
          </Text>
        </View>
      )}
    </>
  );
}

function FilterChip({
  label,
  emoji,
  active,
  onPress,
}: {
  label: string;
  emoji?: string;
  active: boolean;
  onPress: () => void;
}) {
  return (
    <TouchableOpacity style={[styles.filterChip, active && styles.filterChipActive]} onPress={onPress} activeOpacity={0.8}>
      {emoji ? <Text style={{ fontSize: 12 }}>{emoji} </Text> : null}
      <Text style={[typography.labelSm, { color: active ? colors.onPrimary : colors.onSurfaceVariant }]}>{label}</Text>
    </TouchableOpacity>
  );
}

function TimelineCard({
  event,
  moodBefore,
  tierNote,
}: {
  event: TimelineEvent;
  moodBefore: { mood: string; minutes: number } | null;
  tierNote: string | null;
}) {
  if (event.type === 'mood') {
    const mood = event.mood!.Mood;
    return (
      <View style={styles.itemRow}>
        <IconCircle emoji={MOOD_EMOJI[mood] ?? '🌿'} bg={colors.secondaryFixed} />
        <Card style={styles.itemCard}>
          <View style={styles.itemTopRow}>
            <TagPill label="Mood Check-in" tone="secondary" />
            <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>{formatTime(event.occurred_at)}</Text>
          </View>
          <Text style={typography.labelLg}>{MOOD_EMOJI[mood] ?? '🌿'} {MOOD_LABEL[mood] ?? mood}</Text>
          <View style={styles.infoBox}>
            <Icon name="info-outline" size={14} color={colors.onSurfaceVariant} />
            <Text style={[typography.labelSm, { color: colors.onSurfaceVariant, flex: 1 }]}>
              Catatan temporal untuk melihat ritme harimu - bukan sebab-akibat.
            </Text>
          </View>
        </Card>
      </View>
    );
  }

  if (event.type === 'transaction') {
    const tx = event.transaction!;
    const title = tx.ItemName || tx.CategoryName;
    const over = tx.PlanAmountSnapshot != null && tx.Amount > tx.PlanAmountSnapshot;
    return (
      <View style={styles.itemRow}>
        <IconCircle icon="receipt-long" bg={colors.surfaceContainer} />
        <Card style={styles.itemCard}>
          <View style={styles.itemTopRow}>
            <TagPill label={tx.CategoryName} tone="primary" />
            <Text style={typography.currencyMd}>{formatRupiah(tx.Amount)}</Text>
          </View>
          <Text style={typography.labelLg} numberOfLines={1}>{title}</Text>
          {tx.ItemName ? (
            <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>{tx.CategoryName}</Text>
          ) : null}
          {tx.PlanAmountSnapshot != null && (
            <View style={styles.infoBox}>
              <Icon name={over ? 'trending-up' : 'check-circle-outline'} size={14} color={colors.onSurfaceVariant} />
              <Text style={[typography.labelSm, { color: colors.onSurfaceVariant, flex: 1 }]}>
                Baseline kategori {formatRupiah(tx.PlanAmountSnapshot)}
                {over ? ` - selisih +${formatRupiah(tx.Amount - tx.PlanAmountSnapshot)}` : ''}
              </Text>
            </View>
          )}
          {moodBefore && (
            <TagPill label={`Tercatat ${moodBefore.minutes} menit setelah mood ${MOOD_LABEL[moodBefore.mood] ?? moodBefore.mood}`} tone="neutral" icon="schedule" />
          )}
        </Card>
      </View>
    );
  }

  if (event.type === 'bank_ledger') {
    const l = event.bank_ledger!;
    const credited = l.DeltaAmount > 0;
    return (
      <View style={styles.itemRow}>
        <IconCircle icon={credited ? 'savings' : 'shield'} bg={credited ? colors.secondaryFixed : colors.tertiaryFixed} color={credited ? colors.onSecondaryFixedVariant : colors.tertiary} />
        <Card style={styles.itemCard}>
          <View style={styles.itemTopRow}>
            <TagPill label={credited ? 'Aliran Masuk Bank' : 'Consumption Bank'} tone={credited ? 'secondary' : 'tertiary'} />
            <Text style={[typography.currencyMd, { color: credited ? colors.secondary : colors.tertiary }]}>
              {credited ? '+' : ''}{formatRupiah(l.DeltaAmount)}
            </Text>
          </View>
          <Text style={typography.labelLg}>{credited ? 'Buffer Bertambah' : 'Buffer Diterapkan'}</Text>
          <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
            {credited
              ? 'Underspend hari ini dialirkan ke Consumption Bank secara otomatis.'
              : 'Dipakai menutup sebagian selisih pengeluaran (maks. 10% dari saldo).'}
          </Text>
          <View style={styles.infoBox}>
            <Icon name={credited ? 'eco' : 'account-balance-wallet'} size={14} color={colors.onSurfaceVariant} />
            <Text style={[typography.labelSm, { color: colors.onSurfaceVariant, flex: 1 }]}>
              Saldo Buffer Terkini {formatRupiah(l.BalanceAfter)}
            </Text>
          </View>
          {tierNote && (
            <TagPill label={`Pertumbuhan alami - Tier naik ke: ${tierNote}`} tone="secondary" icon="check" />
          )}
        </Card>
      </View>
    );
  }

  // plan_adjustment
  const pa = event.plan_adjustment!;
  return (
    <View style={styles.itemRow}>
      <IconCircle icon="tune" bg={colors.surfaceContainer} />
      <Card style={styles.itemCard}>
        <View style={styles.itemTopRow}>
          <TagPill label="Rencana Fleksibel" tone="secondary" />
          <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>{formatTime(event.occurred_at)}</Text>
        </View>
        <Text style={typography.labelLg}>Alokasi {pa.CategoryName}</Text>
        <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
          Mengubah alokasi dari {formatRupiah(pa.OldAmount)} menjadi {formatRupiah(pa.NewAmount)}.
        </Text>
      </Card>
    </View>
  );
}

function TagPill({
  label,
  tone,
  icon,
}: {
  label: string;
  tone: 'primary' | 'secondary' | 'tertiary' | 'neutral';
  icon?: keyof typeof MaterialIcons.glyphMap;
}) {
  const bg =
    tone === 'primary' ? colors.primaryFixed :
    tone === 'secondary' ? colors.secondaryContainer :
    tone === 'tertiary' ? colors.tertiaryFixed : colors.surfaceContainer;
  const fg =
    tone === 'primary' ? colors.onPrimaryFixedVariant :
    tone === 'secondary' ? colors.onSecondaryContainer :
    tone === 'tertiary' ? colors.tertiary : colors.onSurfaceVariant;
  return (
    <View style={[styles.tagPill, { backgroundColor: bg, flexDirection: 'row', alignItems: 'center', gap: 4 }]}>
      {icon ? <Icon name={icon} size={12} color={fg} /> : null}
      <Text style={[typography.labelSm, { color: fg }]}>{label}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  headerRow: { flexDirection: 'row', alignItems: 'center', gap: spacing.xs },
  filterRow: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.xs },
  filterChip: { backgroundColor: colors.surfaceContainerHigh, borderRadius: radius.full, paddingHorizontal: 12, paddingVertical: 6, flexDirection: 'row' },
  filterChipActive: { backgroundColor: colors.primaryContainer },
  dateFilterChip: {
    flexDirection: 'row', alignItems: 'center', gap: 4, alignSelf: 'flex-start',
    backgroundColor: colors.surfaceContainerLow, borderRadius: radius.full, paddingHorizontal: 10, paddingVertical: 6,
  },
  dateHeaderRow: { flexDirection: 'row', alignItems: 'center', gap: spacing.xs, paddingLeft: 4 },
  dateDot: { width: 6, height: 6, borderRadius: 3, backgroundColor: colors.primaryContainer },
  itemRow: { flexDirection: 'row', gap: spacing.sm, width: '100%' },
  itemCard: { flex: 1, gap: spacing.xs },
  itemTopRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  tagPill: { alignSelf: 'flex-start', borderRadius: radius.full, paddingHorizontal: 8, paddingVertical: 3 },
  infoBox: {
    flexDirection: 'row', alignItems: 'flex-start', gap: 6,
    backgroundColor: colors.surfaceContainerLow, borderRadius: radius.DEFAULT, padding: spacing.sm,
  },
  footerNote: { flexDirection: 'row', justifyContent: 'center', alignItems: 'center', gap: 6, paddingVertical: spacing.sm },
});
