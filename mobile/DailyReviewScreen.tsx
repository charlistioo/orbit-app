import { useEffect, useState } from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { colors, spacing, typography } from './theme';
import { Card, ErrorState, LoadingState, PrimaryButton, SectionTitle, StatusPill } from './ui';

const API_BASE = 'http://localhost:8080/api/v1';

type ReminderResponse = {
  date: string;
  pending_categories: string[];
  message: string;
};

type Props = {
  sessionToken: string;
};

// Shows the server-computed reminder (neutral wording, per TASK-013's
// design) and lets the user finalize the day - completing it, or
// skipping without any penalty. This component does no wording or
// validation logic of its own.
export default function DailyReviewScreen({ sessionToken }: Props) {
  const [reminder, setReminder] = useState<ReminderResponse | null>(null);
  const [status, setStatus] = useState<'skipped' | 'completed' | null>(null);
  const [error, setError] = useState('');

  const load = async () => {
    try {
      const res = await fetch(`${API_BASE}/reminders/today`, {
        headers: { Authorization: `Bearer ${sessionToken}` },
      });
      if (!res.ok) throw new Error(`backend returned ${res.status}`);
      setReminder(await res.json());
      setError('');
    } catch (err: any) {
      setError(`Failed to load reminder: ${err.message}`);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const finalize = async (action: 'complete' | 'skip') => {
    if (!reminder) return;
    try {
      const res = await fetch(`${API_BASE}/daily-review`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${sessionToken}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ date: reminder.date, action }),
      });
      if (!res.ok) throw new Error(`backend returned ${res.status}`);
      const data = await res.json();
      setStatus(data.status);
    } catch (err: any) {
      setError(`Failed to finalize: ${err.message}`);
    }
  };

  if (error) return <ErrorState message={error} />;
  if (!reminder) return <LoadingState />;

  return (
    <Card>
      <View style={styles.headerRow}>
        <SectionTitle icon="event-available">Peninjauan Harian</SectionTitle>
        {status ? (
          <StatusPill label={status === 'skipped' ? 'Dilewati' : 'Selesai'} tone="secondary" icon="check" />
        ) : null}
      </View>

      {reminder.pending_categories.length > 0 ? (
        <View style={styles.pendingList}>
          {reminder.pending_categories.map((name) => (
            <Text key={name} style={[typography.bodyMd, { color: colors.tertiary }]}>
              • {name}
            </Text>
          ))}
        </View>
      ) : null}
      <Text style={[typography.bodyMd, { color: colors.onSurfaceVariant }]}>
        {reminder.message || 'Semua kategori sudah tercatat.'}
      </Text>
      <View style={styles.buttons}>
        <View style={{ flex: 1 }}>
          <PrimaryButton title="Selesai" onPress={() => finalize('complete')} icon="check" />
        </View>
        <View style={{ flex: 1 }}>
          <PrimaryButton title="Lewati" onPress={() => finalize('skip')} variant="secondary" />
        </View>
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  headerRow: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  pendingList: { gap: 2 },
  buttons: { flexDirection: 'row', gap: spacing.sm },
});
