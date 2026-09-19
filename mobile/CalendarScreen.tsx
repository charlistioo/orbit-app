import { useEffect, useMemo, useState } from 'react';
import { StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { colors, radius, spacing, typography } from './theme';
import { Card, ErrorState, Icon, IconCircle, LoadingState, ProgressBar, SectionTitle, SegmentedControl } from './ui';

const API_BASE = 'http://localhost:8080/api/v1';
const WEEKDAYS = ['Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab', 'Min'];
const MONTH_NAMES = [
  'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
  'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember',
];

type CalendarDay = { date: string; baseline: number | null; actual: number };
type CalendarMonth = {
  month: string;
  days_in_month: number;
  days: CalendarDay[];
  cumulative_baseline: number;
  cumulative_actual: number;
};
type TimelineEvent = {
  type: string;
  occurred_at: string;
  transaction?: { CategoryID: string; CategoryName: string; ItemID: string | null; Amount: number };
};

type Props = { sessionToken: string };

function formatRupiah(amount: number): string {
  return 'Rp' + Math.round(amount).toLocaleString('id-ID');
}

function todayStr(): string {
  return new Date().toISOString().slice(0, 10);
}

// Grid/list calendar restyled per the Stitch mockup (TASK-015) - all
// numbers still come straight from GET /api/v1/calendar and GET
// /api/v1/consumption-bank (per TASK-011's server-side-only design);
// this component only adds day-of-week grid placement and client-side
// filtering of the already-fetched Timeline for the selected day's
// transaction list.
export default function CalendarScreen({ sessionToken }: Props) {
  const now = new Date();
  const [year, setYear] = useState(now.getFullYear());
  const [month, setMonth] = useState(now.getMonth() + 1); // 1-12
  const [data, setData] = useState<CalendarMonth | null>(null);
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [bankBalance, setBankBalance] = useState<number | null>(null);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [selectedDate, setSelectedDate] = useState<string>(todayStr());
  const [affirmed, setAffirmed] = useState<Record<string, boolean>>({});
  const [error, setError] = useState('');

  const authHeaders = { Authorization: `Bearer ${sessionToken}` };

  const loadCalendar = async () => {
    try {
      const monthParam = `${year}-${String(month).padStart(2, '0')}`;
      const res = await fetch(`${API_BASE}/calendar?month=${monthParam}`, { headers: authHeaders });
      if (!res.ok) throw new Error(`backend returned ${res.status}`);
      setData(await res.json());
      setError('');
    } catch (err: any) {
      setError(`Failed to load calendar: ${err.message}`);
    }
  };

  useEffect(() => {
    loadCalendar();
  }, [year, month]);

  useEffect(() => {
    fetch(`${API_BASE}/transactions`, { headers: authHeaders })
      .then((r) => r.json())
      .then((d) => setEvents(d.events ?? []))
      .catch(() => {});
    fetch(`${API_BASE}/consumption-bank`, { headers: authHeaders })
      .then((r) => r.json())
      .then((d) => setBankBalance(d.balance ?? null))
      .catch(() => {});
  }, []);

  const goMonth = (delta: number) => {
    let m = month + delta;
    let y = year;
    if (m < 1) { m = 12; y -= 1; }
    if (m > 12) { m = 1; y += 1; }
    setMonth(m);
    setYear(y);
  };

  const selectedDay = data?.days.find((d) => d.date === selectedDate) ?? null;
  const dayTransactions = useMemo(
    () =>
      events.filter(
        (e) => e.type === 'transaction' && e.transaction && e.occurred_at.slice(0, 10) === selectedDate
      ),
    [events, selectedDate]
  );

  if (error) return <ErrorState message={error} />;
  if (!data) return <LoadingState />;

  const cumulativeProgress = data.cumulative_baseline > 0 ? data.cumulative_actual / data.cumulative_baseline : 0;
  const isCurrentMonth = year === now.getFullYear() && month === now.getMonth() + 1;
  const dayOfMonthLabel = isCurrentMonth ? `Hari ke-${now.getDate()} / ${data.days_in_month}` : `${data.days_in_month} hari`;

  const firstWeekday = (new Date(year, month - 1, 1).getDay() + 6) % 7; // 0=Senin
  const leadingBlanks = Array.from({ length: firstWeekday });

  return (
    <>
      <SegmentedControl
        options={[
          { key: 'grid', label: 'Tampilan Kalender' },
          { key: 'list', label: 'Daftar Riwayat' },
        ]}
        value={viewMode}
        onChange={setViewMode}
      />

      <View style={styles.monthNav}>
        <View style={styles.monthNavLeft}>
          <Text style={typography.headlineMd}>{MONTH_NAMES[month - 1]} {year}</Text>
          {isCurrentMonth ? (
            <View style={styles.activePill}>
              <Text style={[typography.labelSm, { color: colors.onSecondaryContainer }]}>Aktif</Text>
            </View>
          ) : null}
        </View>
        <View style={styles.monthNavRight}>
          <TouchableOpacity style={styles.navBtn} onPress={() => goMonth(-1)}>
            <Icon name="chevron-left" size={20} />
          </TouchableOpacity>
          <TouchableOpacity style={styles.navBtn} onPress={() => goMonth(1)}>
            <Icon name="chevron-right" size={20} />
          </TouchableOpacity>
        </View>
      </View>

      <Card style={styles.heroCard}>
        <View style={styles.heroTopRow}>
          <Text style={[typography.labelMd, { color: colors.onSurfaceVariant, textTransform: 'uppercase' }]}>
            Total Pengeluaran Bulan Ini
          </Text>
          <View style={styles.dayCounter}>
            <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>{dayOfMonthLabel}</Text>
          </View>
        </View>
        <View style={styles.heroAmountRow}>
          <Text style={[typography.currencyDisplay, { color: colors.primary }]}>{formatRupiah(data.cumulative_actual)}</Text>
          <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>/ {formatRupiah(data.cumulative_baseline)}</Text>
        </View>

        <ProgressBar progress={cumulativeProgress} />
        <View style={styles.heroTopRow}>
          <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>
            Realisasi Akumulatif ({Math.round(cumulativeProgress * 100)}%)
          </Text>
          <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>Target Pagu: {formatRupiah(data.cumulative_baseline)}</Text>
        </View>

        {bankBalance !== null && (
          <View style={styles.bufferCallout}>
            <IconCircle icon="eco" bg={colors.secondaryContainer} color={colors.secondary} size={32} />
            <View style={{ flex: 1 }}>
              <Text style={[typography.labelMd, { color: colors.secondary }]}>{formatRupiah(bankBalance)} tersimpan</Text>
              <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>Saldo Consumption Bank saat ini.</Text>
            </View>
          </View>
        )}
      </Card>

      <View style={styles.legend}>
        <LegendItem color={colors.secondary} label="Terkendali" />
        <LegendItem color={colors.tertiaryFixedDim} label="Dinamis" />
        <LegendItem color={colors.outlineVariant} label="Belum tiba" />
      </View>

      {viewMode === 'grid' ? (
        <Card>
          <View style={styles.weekdayRow}>
            {WEEKDAYS.map((w) => (
              <Text key={w} style={[typography.labelMd, styles.weekdayCell, { color: colors.onSurfaceVariant }]}>
                {w}
              </Text>
            ))}
          </View>
          <View style={styles.grid}>
            {leadingBlanks.map((_, i) => (
              <View key={`b${i}`} style={styles.dayCell} />
            ))}
            {data.days.map((d) => {
              const dayNum = parseInt(d.date.slice(-2), 10);
              const hasPlan = d.baseline !== null;
              const isFuture = d.date > todayStr();
              const over = hasPlan && d.actual > (d.baseline as number);
              const dotColor = !hasPlan || isFuture ? colors.outlineVariant : over ? colors.tertiaryFixedDim : colors.secondary;
              const selected = d.date === selectedDate;
              return (
                <TouchableOpacity
                  key={d.date}
                  style={[styles.dayCell, selected && styles.dayCellSelected]}
                  onPress={() => setSelectedDate(d.date)}
                  activeOpacity={0.8}
                >
                  <Text style={[typography.labelMd, { color: selected ? colors.onPrimary : colors.onSurface }]}>{dayNum}</Text>
                  <View style={[styles.dot, { backgroundColor: selected ? colors.secondaryFixed : dotColor }]} />
                </TouchableOpacity>
              );
            })}
          </View>
        </Card>
      ) : (
        <View style={{ width: '100%', gap: spacing.sm }}>
          {[...data.days].reverse().map((d) => {
            const hasPlan = d.baseline !== null;
            const over = hasPlan && d.actual > (d.baseline as number);
            const diff = hasPlan ? (d.baseline as number) - d.actual : 0;
            return (
              <Card key={d.date} style={styles.listRow}>
                <View>
                  <Text style={typography.labelMd}>{d.date}</Text>
                  {hasPlan ? (
                    <Text style={[typography.bodySm, { color: over ? colors.tertiary : colors.secondary }]}>
                      Realisasi: {formatRupiah(d.actual)} ({diff >= 0 ? 'Hemat' : 'Lebih'} {formatRupiah(Math.abs(diff))})
                    </Text>
                  ) : (
                    <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
                      Realisasi: {formatRupiah(d.actual)} (tanpa rencana)
                    </Text>
                  )}
                </View>
                <View style={[styles.dot, { backgroundColor: !hasPlan ? colors.outlineVariant : over ? colors.tertiaryFixedDim : colors.secondary }]} />
              </Card>
            );
          })}
        </View>
      )}

      <Card>
        <View style={styles.detailHeader}>
          <View style={{ flex: 1 }}>
            <SectionTitle icon="event-available">Detail {selectedDate}</SectionTitle>
            <Text style={[typography.bodySm, { color: colors.onSurfaceVariant, marginTop: 2 }]}>
              {selectedDay?.baseline != null
                ? `Realisasi ${formatRupiah(selectedDay.actual)} dari Baseline ${formatRupiah(selectedDay.baseline)}`
                : `Realisasi ${formatRupiah(selectedDay?.actual ?? 0)} - belum ada rencana untuk tanggal ini`}
            </Text>
          </View>
        </View>

        {selectedDay?.baseline != null && (
          <View style={styles.progressWell}>
            <View style={styles.heroTopRow}>
              <Text
                style={[
                  typography.labelMd,
                  { color: selectedDay.actual <= selectedDay.baseline ? colors.secondary : colors.tertiary },
                ]}
              >
                {selectedDay.actual <= selectedDay.baseline
                  ? `+${formatRupiah(selectedDay.baseline - selectedDay.actual)} Belum Terpakai`
                  : `-${formatRupiah(selectedDay.actual - selectedDay.baseline)} Melebihi Rencana`}
              </Text>
              <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>
                {Math.round((selectedDay.actual / selectedDay.baseline) * 100)}% Terpakai
              </Text>
            </View>
            <ProgressBar
              progress={selectedDay.actual / selectedDay.baseline}
              color={selectedDay.actual <= selectedDay.baseline ? colors.secondary : colors.tertiaryFixedDim}
            />
          </View>
        )}

        {dayTransactions.length > 0 && (
          <View style={{ gap: spacing.xs }}>
            <Text style={[typography.labelMd, { color: colors.onSurfaceVariant, textTransform: 'uppercase' }]}>
              Aktivitas Pengeluaran
            </Text>
            {dayTransactions.map((e, i) => (
              <View key={i} style={styles.txRow}>
                <View style={styles.txLeft}>
                  <IconCircle icon="payments" size={36} />
                  <View>
                    <Text style={typography.labelLg}>{e.transaction!.CategoryName}</Text>
                    <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
                      {new Date(e.occurred_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}
                    </Text>
                  </View>
                </View>
                <Text style={typography.currencyMd}>{formatRupiah(e.transaction!.Amount)}</Text>
              </View>
            ))}
          </View>
        )}

        {dayTransactions.length > 0 && (
          <TouchableOpacity
            style={styles.affirmBtn}
            onPress={() => setAffirmed((prev) => ({ ...prev, [selectedDate]: true }))}
            activeOpacity={0.8}
          >
            <Icon name={affirmed[selectedDate] ? 'check-circle' : 'verified'} size={16} color={colors.primaryContainer} />
            <Text style={[typography.labelLg, { color: colors.primaryContainer }]}>
              {affirmed[selectedDate] ? 'Catatan Refleksi Sudah Sesuai' : 'Tandai Catatan Sudah Sesuai'}
            </Text>
          </TouchableOpacity>
        )}
      </Card>
    </>
  );
}

function LegendItem({ color, label }: { color: string; label: string }) {
  return (
    <View style={styles.legendItem}>
      <View style={[styles.legendDot, { backgroundColor: color }]} />
      <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>{label}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  monthNav: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', width: '100%' },
  monthNavLeft: { flexDirection: 'row', alignItems: 'center', gap: spacing.xs },
  monthNavRight: { flexDirection: 'row', gap: spacing.xs },
  activePill: { backgroundColor: colors.secondaryContainer, borderRadius: radius.full, paddingHorizontal: 8, paddingVertical: 2 },
  navBtn: {
    width: 34, height: 34, borderRadius: 17,
    backgroundColor: colors.surfaceContainer,
    alignItems: 'center', justifyContent: 'center',
  },
  heroCard: { gap: spacing.sm },
  heroTopRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  dayCounter: { backgroundColor: colors.surfaceContainer, borderRadius: radius.full, paddingHorizontal: 8, paddingVertical: 2 },
  heroAmountRow: { flexDirection: 'row', alignItems: 'baseline', gap: 6 },
  bufferCallout: {
    flexDirection: 'row', alignItems: 'center', gap: spacing.sm,
    backgroundColor: colors.surfaceContainerLow, borderRadius: radius.DEFAULT, padding: spacing.sm,
  },
  legend: { flexDirection: 'row', justifyContent: 'space-between', width: '100%', paddingHorizontal: spacing.xs },
  legendItem: { flexDirection: 'row', alignItems: 'center', gap: 6 },
  legendDot: { width: 8, height: 8, borderRadius: 4 },
  weekdayRow: { flexDirection: 'row' },
  weekdayCell: { flex: 1, textAlign: 'center' },
  grid: { flexDirection: 'row', flexWrap: 'wrap' },
  dayCell: {
    width: `${100 / 7}%`,
    height: 44,
    alignItems: 'center',
    justifyContent: 'center',
    borderRadius: radius.DEFAULT,
    marginVertical: 2,
  },
  dayCellSelected: { backgroundColor: colors.primaryContainer },
  dot: { width: 5, height: 5, borderRadius: 3, marginTop: 3 },
  listRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  detailHeader: { flexDirection: 'row', alignItems: 'flex-start' },
  progressWell: { backgroundColor: colors.surfaceContainerLow, borderRadius: radius.DEFAULT, padding: spacing.sm, gap: spacing.xs },
  txRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', paddingVertical: 4 },
  txLeft: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm },
  affirmBtn: {
    flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 6,
    backgroundColor: colors.surfaceContainer, borderRadius: radius.full, paddingVertical: 10,
  },
});
