import { useState } from 'react';
import { StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { colors, radius, spacing, typography } from './theme';
import { Card, Icon, SectionTitle } from './ui';

const API_BASE = 'http://localhost:8080/api/v1';

const MOODS: { value: string; emoji: string; label: string }[] = [
  { value: 'very_good', emoji: '🤩', label: 'Bersemangat' },
  { value: 'good', emoji: '😊', label: 'Senang' },
  { value: 'neutral', emoji: '🌿', label: 'Tenang' },
  { value: 'stressed', emoji: '🥱', label: 'Lelah' },
  { value: 'sad', emoji: '🌧️', label: 'Cemas' },
];

type Props = {
  sessionToken: string;
  showCorrelationNote?: boolean;
};

// Mood is entirely optional and can be logged as many times a day as
// the user wants (each tap is its own timestamped entry - never
// overwriting a previous one).
export default function MoodSelector({ sessionToken, showCorrelationNote }: Props) {
  const [status, setStatus] = useState('');
  const [lastLogged, setLastLogged] = useState<string | null>(null);

  const logMood = async (value: string) => {
    setStatus('Menyimpan...');
    try {
      const res = await fetch(`${API_BASE}/mood`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${sessionToken}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ mood: value }),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error ?? `backend returned ${res.status}`);
      }
      setLastLogged(value);
      setStatus('Mood tersimpan');
    } catch (err: any) {
      setStatus(`Gagal: ${err.message}`);
    }
  };

  return (
    <Card>
      <SectionTitle icon="spa">Bagaimana perasaanmu sekarang?</SectionTitle>
      <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>
        Satu sentuhan untuk menandai suasana hati saat berbelanja.
      </Text>
      <View style={styles.row}>
        {MOODS.map((m) => {
          const selected = lastLogged === m.value;
          return (
            <TouchableOpacity
              key={m.value}
              style={[styles.moodButton, selected && styles.moodButtonSelected]}
              onPress={() => logMood(m.value)}
              activeOpacity={0.8}
            >
              <Text style={styles.emoji}>{m.emoji}</Text>
              <Text style={[typography.labelSm, { color: selected ? colors.primary : colors.onSurface }]}>
                {m.label}
              </Text>
            </TouchableOpacity>
          );
        })}
      </View>
      {status ? <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>{status}</Text> : null}
      {showCorrelationNote && (
        <View style={styles.noteBox}>
          <Icon name="favorite" size={16} color={colors.primaryContainer} />
          <Text style={[typography.bodySm, { color: colors.onSurfaceVariant, flex: 1 }]}>
            Mood dicatat sebagai korelasi waktu di timeline, tanpa dihakimi. Pola ini membantumu memahami ritme diri.
          </Text>
        </View>
      )}
    </Card>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', gap: spacing.xs, justifyContent: 'space-between' },
  moodButton: {
    alignItems: 'center',
    gap: 4,
    paddingVertical: spacing.sm,
    paddingHorizontal: 4,
    borderRadius: radius.DEFAULT,
    backgroundColor: colors.surfaceContainerLow,
    flex: 1,
  },
  moodButtonSelected: { backgroundColor: colors.surfaceContainerHighest },
  emoji: { fontSize: 22 },
  noteBox: {
    flexDirection: 'row', alignItems: 'flex-start', gap: spacing.sm,
    backgroundColor: colors.primaryFixed, borderRadius: radius.DEFAULT, padding: spacing.sm,
  },
});
