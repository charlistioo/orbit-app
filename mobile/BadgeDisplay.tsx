import { useEffect, useState } from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { colors, spacing, typography } from './theme';
import { Card, ErrorState, IconCircle, LoadingState } from './ui';

const API_BASE = 'http://localhost:8080/api/v1';

// Mirrors the 5 tiers computed server-side in badgeTierFor (TASK-007):
// Frugal >= 100k, Thrifty >= 200k, Economical >= 300k, Collector >= 400k,
// Master >= 500k. "" means the balance hasn't reached the first tier yet.
type BadgeTier = 'Frugal' | 'Thrifty' | 'Economical' | 'Collector' | 'Master' | '';

type BankData = { balance: number; badge: BadgeTier };
type Props = { sessionToken: string };

function formatRupiah(amount: number): string {
  return 'Rp' + Math.round(amount).toLocaleString('id-ID');
}

// badgeVisual maps every possible tier to its display - an exhaustive
// switch so TypeScript itself flags a missing tier at compile time if a
// 6th one is ever added.
export function badgeVisual(badge: BadgeTier): { emoji: string; label: string; bg: string; fg: string } {
  switch (badge) {
    case 'Frugal':
      return { emoji: '🌱', label: 'Frugal', bg: colors.secondaryFixed, fg: colors.onSecondaryFixedVariant };
    case 'Thrifty':
      return { emoji: '🌿', label: 'Thrifty', bg: colors.secondaryContainer, fg: colors.onSecondaryContainer };
    case 'Economical':
      return { emoji: '🍀', label: 'Economical', bg: colors.primaryFixed, fg: colors.onPrimaryFixedVariant };
    case 'Collector':
      return { emoji: '🌳', label: 'Collector', bg: colors.primaryFixedDim, fg: colors.onPrimaryFixed };
    case 'Master':
      return { emoji: '🏆', label: 'Master', bg: colors.tertiaryFixed, fg: colors.onTertiaryFixedVariant };
    case '':
      return { emoji: '🪙', label: 'Belum ada badge', bg: colors.surfaceContainer, fg: colors.onSurfaceVariant };
  }
}

// Dedicated Consumption Bank badge display (per TASK-014) - fetches its
// own data and renders whichever of the 5 tiers (or none yet) the
// backend reports, doing no threshold math of its own.
export default function BadgeDisplay({ sessionToken }: Props) {
  const [data, setData] = useState<BankData | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    fetch(`${API_BASE}/consumption-bank`, { headers: { Authorization: `Bearer ${sessionToken}` } })
      .then((res) => {
        if (!res.ok) throw new Error(`backend returned ${res.status}`);
        return res.json();
      })
      .then(setData)
      .catch((err) => setError(`Failed to load badge: ${err.message}`));
  }, []);

  if (error) return <ErrorState message={error} />;
  if (!data) return <LoadingState />;

  const visual = badgeVisual(data.badge);

  return (
    <Card style={styles.card}>
      <IconCircle emoji={visual.emoji} size={56} bg={visual.bg} />
      <View style={{ flex: 1 }}>
        <Text style={[typography.labelMd, { color: colors.onSurfaceVariant, textTransform: 'uppercase' }]}>
          Consumption Bank
        </Text>
        <Text style={[typography.headlineSm, { color: visual.fg }]}>{visual.label}</Text>
        <Text style={[typography.bodySm, { color: colors.onSurfaceVariant }]}>Saldo {formatRupiah(data.balance)}</Text>
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  card: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
});
