import { useEffect, useState } from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { colors, spacing, typography } from './theme';
import { Card, ErrorState, IconCircle, LoadingState, SectionTitle } from './ui';

const API_BASE = 'http://localhost:8080/api/v1';

type InsightResult = { period: string; date: string; insights: string[] };
type Props = { sessionToken: string };

// Shows the template-based insight text generated server-side (per
// TASK-012's design) - this component does no rule logic of its own,
// only fetches and lists whatever the backend already decided to say.
export default function InsightsScreen({ sessionToken }: Props) {
  const [daily, setDaily] = useState<InsightResult | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    fetch(`${API_BASE}/insights/daily`, { headers: { Authorization: `Bearer ${sessionToken}` } })
      .then((res) => {
        if (!res.ok) throw new Error(`backend returned ${res.status}`);
        return res.json();
      })
      .then(setDaily)
      .catch((err) => setError(`Failed to load insights: ${err.message}`));
  }, []);

  if (error) return <ErrorState message={error} />;
  if (!daily) return <LoadingState />;

  return (
    <Card>
      <SectionTitle icon="lightbulb">Insight Hari Ini</SectionTitle>
      {daily.insights.length === 0 ? (
        <Text style={[typography.bodyMd, { color: colors.onSurfaceVariant }]}>Belum ada insight.</Text>
      ) : (
        <View style={styles.list}>
          {daily.insights.map((text, i) => (
            <View key={i} style={styles.item}>
              <IconCircle emoji="💡" size={32} bg={colors.tertiaryFixed} />
              <Text style={[typography.bodyMd, { flex: 1 }]}>{text}</Text>
            </View>
          ))}
        </View>
      )}
    </Card>
  );
}

const styles = StyleSheet.create({
  list: { gap: spacing.sm },
  item: { flexDirection: 'row', gap: spacing.sm, alignItems: 'center' },
});
