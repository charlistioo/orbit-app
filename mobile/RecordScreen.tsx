import { useEffect, useMemo, useState } from 'react';
import { FlatList, StyleSheet, Text, TextInput, View } from 'react-native';
import { colors, radius, spacing, typography } from './theme';
import { Card, Chip, ErrorState, Icon, IconCircle, PrimaryButton, StatusPill } from './ui';
import MoodSelector from './MoodSelector';
import { categoryEmoji } from './categoryEmoji';

const API_BASE = 'http://localhost:8080/api/v1';
const PRESETS = [5000, 10000, 20000, 50000];

type Category = { ID: string; Name: string };
type Item = { ID: string; CategoryID: string; Name: string };
type ReminderResponse = { date: string; pending_categories: string[]; message: string };
type TimelineEvent = { type: string; occurred_at: string };

type Props = {
  sessionToken: string;
  onSaved: () => void;
};

function formatRupiah(amount: number): string {
  return amount.toLocaleString('id-ID');
}
function todayStr(): string {
  return new Date().toISOString().slice(0, 10);
}
function todayLabel(): string {
  return new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'short' });
}

type ReviewAnswer = 'sudah' | 'belum' | 'tidak_ada';

// The "Catat" tab (TASK-015 restyle, ported from the Stitch mockups) -
// records a transaction against a quick-selected category/item, shows
// mood check-in, and a daily review card. Per-category Sudah/Belum/
// Tidak-ada answers are visual-only local state (the backend's
// daily_reviews table only stores one status for the whole day, not a
// per-category answer) - the two final actions ("Lewati Hari Ini" /
// "Selesai Tinjau") are what actually call POST /daily-review.
export default function RecordScreen({ sessionToken, onSaved }: Props) {
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(null);
  const [items, setItems] = useState<Item[]>([]);
  const [amountText, setAmountText] = useState('');
  const [itemName, setItemName] = useState('');
  const [newCategoryName, setNewCategoryName] = useState('');
  const [addingCategory, setAddingCategory] = useState(false);
  const [error, setError] = useState('');
  const [status, setStatus] = useState('');

  const [reminder, setReminder] = useState<ReminderResponse | null>(null);
  const [reviewAnswers, setReviewAnswers] = useState<Record<string, ReviewAnswer>>({});
  const [reviewStatus, setReviewStatus] = useState<'skipped' | 'completed' | null>(null);
  const [weeklyCount, setWeeklyCount] = useState<number | null>(null);

  const authHeaders = {
    Authorization: `Bearer ${sessionToken}`,
    'Content-Type': 'application/json',
  };

  const loadCategories = async () => {
    try {
      const res = await fetch(`${API_BASE}/categories`, { headers: authHeaders });
      const data = await res.json();
      const list: Category[] = data.categories ?? [];
      setCategories(list);
      if (!selectedCategory && list.length > 0) selectCategory(list[0]);
    } catch {
      setError('Gagal memuat kategori');
    }
  };

  const loadItems = async (category: Category) => {
    try {
      const res = await fetch(`${API_BASE}/categories/${category.ID}/items`, { headers: authHeaders });
      const data = await res.json();
      setItems(data.items ?? []);
    } catch {
      setError('Gagal memuat item');
    }
  };

  const loadReminder = async () => {
    try {
      const res = await fetch(`${API_BASE}/reminders/today`, { headers: authHeaders });
      setReminder(await res.json());
    } catch {
      // reminder is non-critical for the entry form itself
    }
  };

  const loadWeeklyCount = async () => {
    try {
      const res = await fetch(`${API_BASE}/transactions`, { headers: authHeaders });
      const data = await res.json();
      const events: TimelineEvent[] = data.events ?? [];
      const weekAgo = new Date(Date.now() - 7 * 86400000).toISOString().slice(0, 10);
      const days = new Set(
        events
          .filter((e) => e.type === 'transaction' && e.occurred_at.slice(0, 10) >= weekAgo)
          .map((e) => e.occurred_at.slice(0, 10))
      );
      setWeeklyCount(days.size);
    } catch {
      setWeeklyCount(null);
    }
  };

  useEffect(() => {
    loadCategories();
    loadReminder();
    loadWeeklyCount();
  }, []);

  const selectCategory = (category: Category) => {
    setSelectedCategory(category);
    setItems([]);
    loadItems(category);
  };

  const createCategory = async () => {
    if (!newCategoryName) return;
    try {
      const res = await fetch(`${API_BASE}/categories`, {
        method: 'POST',
        headers: authHeaders,
        body: JSON.stringify({ name: newCategoryName }),
      });
      const created: Category = await res.json();
      setNewCategoryName('');
      setAddingCategory(false);
      await loadCategories();
      selectCategory(created);
    } catch {
      setError('Gagal membuat kategori');
    }
  };

  const addPreset = (value: number) => {
    const current = parseInt(amountText.replace(/\D/g, ''), 10) || 0;
    setAmountText(String(current + value));
  };

  const save = async () => {
    const amount = parseInt(amountText.replace(/\D/g, ''), 10);
    if (!selectedCategory || !amount) {
      setError('Pilih kategori dan isi jumlah terlebih dahulu');
      return;
    }
    setError('');
    setStatus('Menyimpan...');
    try {
      let itemID: string | undefined;
      if (itemName) {
        const existing = items.find((i) => i.Name === itemName);
        if (existing) {
          itemID = existing.ID;
        } else {
          const res = await fetch(`${API_BASE}/categories/${selectedCategory.ID}/items`, {
            method: 'POST',
            headers: authHeaders,
            body: JSON.stringify({ name: itemName }),
          });
          const created: Item = await res.json();
          itemID = created.ID;
        }
      }

      const res = await fetch(`${API_BASE}/transactions`, {
        method: 'POST',
        headers: authHeaders,
        body: JSON.stringify({ category_id: selectedCategory.ID, item_id: itemID ?? null, amount }),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error ?? `backend returned ${res.status}`);
      }

      setStatus(`Tersimpan: Rp${formatRupiah(amount)}`);
      setAmountText('');
      setItemName('');
      await loadItems(selectedCategory);
      await loadReminder();
      await loadWeeklyCount();
      onSaved();
    } catch (err: any) {
      setStatus('');
      setError(`Gagal menyimpan: ${err.message}`);
    }
  };

  const finalizeReview = async (action: 'complete' | 'skip') => {
    if (!reminder) return;
    try {
      const res = await fetch(`${API_BASE}/daily-review`, {
        method: 'POST',
        headers: authHeaders,
        body: JSON.stringify({ date: reminder.date, action }),
      });
      if (!res.ok) throw new Error(`backend returned ${res.status}`);
      const data = await res.json();
      setReviewStatus(data.status);
    } catch (err: any) {
      setError(`Gagal menyimpan peninjauan: ${err.message}`);
    }
  };

  return (
    <>
      <Card style={styles.introCard}>
        <View style={styles.introBlob} />
        <IconCircle emoji="🌱" bg={colors.secondaryFixed} />
        <View style={{ flex: 1 }}>
          <Text style={typography.headlineSm}>Catat Pengeluaran</Text>
          <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
            Satu catatan jujur membawa ketenangan.
          </Text>
        </View>
      </Card>

      <Card>
        <Text style={[typography.labelMd, { color: colors.onSurfaceVariant, textTransform: 'uppercase' }]}>
          Jumlah Transaksi
        </Text>
        <View style={styles.amountRow}>
          <Text style={[typography.currencyDisplay, { color: colors.primaryContainer }]}>Rp</Text>
          <TextInput
            style={styles.amountInput}
            keyboardType="numeric"
            placeholder="0"
            placeholderTextColor={colors.outlineVariant}
            value={amountText ? formatRupiah(parseInt(amountText, 10) || 0) : ''}
            onChangeText={(t) => setAmountText(t.replace(/\D/g, ''))}
          />
        </View>
        <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
          Bisa disesuaikan kapan saja tanpa batas minimum.
        </Text>
        <FlatList
          horizontal
          showsHorizontalScrollIndicator={false}
          data={PRESETS}
          keyExtractor={(p) => String(p)}
          renderItem={({ item }) => <Chip label={`+Rp${item / 1000}rb`} onPress={() => addPreset(item)} />}
        />

        <View style={styles.divider} />

        <View style={styles.rowBetween}>
          <Text style={typography.labelLg}>Kategori</Text>
          <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>Sesuai aslinya</Text>
        </View>
        <FlatList
          horizontal
          showsHorizontalScrollIndicator={false}
          data={categories}
          keyExtractor={(c) => c.ID}
          renderItem={({ item }) => (
            <Chip label={item.Name} selected={selectedCategory?.ID === item.ID} onPress={() => selectCategory(item)} />
          )}
          ListFooterComponent={<Chip label="Kategori Baru" onPress={() => setAddingCategory((v) => !v)} />}
        />
        {addingCategory && (
          <View style={styles.row}>
            <TextInput
              style={styles.input}
              placeholder="Nama kategori baru"
              placeholderTextColor={colors.outlineVariant}
              value={newCategoryName}
              onChangeText={setNewCategoryName}
            />
            <PrimaryButton title="Tambah" onPress={createCategory} variant="secondary" />
          </View>
        )}

        <Text style={typography.labelLg}>Nama Item / Catatan</Text>
        <View style={styles.itemInputWrap}>
          <TextInput
            style={styles.itemInput}
            placeholder="Untuk apa hari ini?"
            placeholderTextColor={colors.outlineVariant}
            value={itemName}
            onChangeText={setItemName}
          />
          {itemName ? <Icon name="close" size={18} color={colors.onSurfaceVariant} /> : null}
        </View>
        {items.length > 0 && (
          <FlatList
            horizontal
            showsHorizontalScrollIndicator={false}
            data={items}
            keyExtractor={(i) => i.ID}
            renderItem={({ item }) => <Chip label={item.Name} onPress={() => setItemName(item.Name)} />}
          />
        )}

        {error ? <ErrorState message={error} /> : null}
      </Card>

      <MoodSelector sessionToken={sessionToken} showCorrelationNote />

      <PrimaryButton title="Simpan Pengeluaran" onPress={save} icon="check-circle" />
      {status ? <Text style={[typography.bodySm, { color: colors.secondary, textAlign: 'center' }]}>{status}</Text> : null}

      {reminder && (
        <Card>
          <View style={styles.rowBetween}>
            <View style={styles.sectionTitleRow}>
              <Icon name="check-circle" size={18} color={colors.secondary} />
              <Text style={typography.headlineSm}>Peninjauan Hari Ini</Text>
            </View>
            <StatusPill label={todayLabel()} tone="neutral" />
          </View>
          <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
            Cek alokasi kategori hari ini agar catatanmu utuh:
          </Text>

          {reminder.pending_categories.length === 0 ? (
            <Text style={[typography.bodyMd, { color: colors.onSurfaceVariant }]}>Semua kategori sudah tercatat.</Text>
          ) : (
            reminder.pending_categories.map((name) => (
              <ReviewRow
                key={name}
                name={name}
                answer={reviewAnswers[name]}
                onAnswer={(a) => setReviewAnswers((prev) => ({ ...prev, [name]: a }))}
              />
            ))
          )}

          <View style={styles.buttons}>
            <View style={{ flex: 1 }}>
              <PrimaryButton title="Lewati Hari Ini" onPress={() => finalizeReview('skip')} variant="secondary" icon="redo" />
            </View>
            <View style={{ flex: 1 }}>
              <PrimaryButton title="Selesai Tinjau" onPress={() => finalizeReview('complete')} icon="check" />
            </View>
          </View>
          <Text style={[typography.labelSm, { color: colors.onSurfaceVariant, textAlign: 'center' }]}>
            Tetap valid, lanjut tanpa penilaian jika terburu-buru.
          </Text>
          {reviewStatus && (
            <StatusPill label={reviewStatus === 'skipped' ? 'Hari ini: dilewati' : 'Hari ini: selesai'} tone="secondary" icon="check" />
          )}
        </Card>
      )}

      {weeklyCount !== null && weeklyCount > 0 && (
        <Card style={styles.streakCard}>
          <IconCircle emoji="🌾" bg={colors.secondaryFixed} size={36} />
          <Text style={[typography.bodySm, { color: colors.onSecondaryFixedVariant, flex: 1 }]}>
            Kamu sudah mencatat {weeklyCount} hari dalam 7 hari terakhir. Pola belanjamu semakin teratur dan penuh kesadaran.
          </Text>
        </Card>
      )}
    </>
  );
}

function ReviewRow({
  name,
  answer,
  onAnswer,
}: {
  name: string;
  answer?: ReviewAnswer;
  onAnswer: (a: ReviewAnswer) => void;
}) {
  return (
    <View style={styles.reviewRow}>
      <View style={styles.rowBetween}>
        <View style={{ flexDirection: 'row', alignItems: 'center', gap: 6 }}>
          <Text style={{ fontSize: 16 }}>{categoryEmoji(name)}</Text>
          <Text style={typography.labelLg}>{name}</Text>
        </View>
      </View>
      <View style={styles.reviewButtons}>
        <ReviewButton label="Sudah" active={answer === 'sudah'} onPress={() => onAnswer('sudah')} />
        <ReviewButton label="Belum" active={answer === 'belum'} onPress={() => onAnswer('belum')} />
        <ReviewButton label="Tidak ada" active={answer === 'tidak_ada'} tone="secondary" onPress={() => onAnswer('tidak_ada')} />
      </View>
    </View>
  );
}

function ReviewButton({
  label,
  active,
  onPress,
  tone = 'primary',
}: {
  label: string;
  active?: boolean;
  onPress: () => void;
  tone?: 'primary' | 'secondary';
}) {
  return (
    <Chip
      label={active ? `${label} ✓` : label}
      selected={false}
      onPress={onPress}
    />
  );
}

const styles = StyleSheet.create({
  introCard: { flexDirection: 'row', alignItems: 'center', gap: spacing.md, overflow: 'hidden' },
  introBlob: {
    position: 'absolute', right: -20, top: -20, width: 90, height: 90, borderRadius: 45,
    backgroundColor: colors.secondaryFixed, opacity: 0.25,
  },
  amountRow: { flexDirection: 'row', alignItems: 'baseline', gap: spacing.xs },
  amountInput: { flex: 1, ...typography.headlineXl, color: colors.onSurface, padding: 0 },
  divider: { height: 1, backgroundColor: colors.surfaceContainerHigh, marginVertical: spacing.xs },
  rowBetween: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  row: { flexDirection: 'row', gap: spacing.sm, alignItems: 'center' },
  sectionTitleRow: { flexDirection: 'row', alignItems: 'center', gap: spacing.xs },
  input: {
    flex: 1,
    backgroundColor: colors.surfaceContainerLow,
    borderRadius: radius.DEFAULT,
    paddingHorizontal: spacing.md,
    paddingVertical: 10,
    color: colors.onSurface,
    ...typography.bodyMd,
  },
  itemInputWrap: {
    flexDirection: 'row', alignItems: 'center',
    backgroundColor: colors.surfaceContainerLow, borderRadius: radius.DEFAULT,
    paddingHorizontal: spacing.md, paddingVertical: 10,
  },
  itemInput: { flex: 1, color: colors.onSurface, ...typography.bodyMd, padding: 0 },
  buttons: { flexDirection: 'row', gap: spacing.sm },
  reviewRow: {
    backgroundColor: colors.surfaceContainerLow, borderRadius: radius.DEFAULT,
    padding: spacing.sm, gap: spacing.sm,
  },
  reviewButtons: { flexDirection: 'row', gap: spacing.xs },
  streakCard: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm, backgroundColor: colors.secondaryContainer },
});
