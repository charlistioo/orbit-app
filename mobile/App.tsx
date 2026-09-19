import { useState } from 'react';
import { StatusBar } from 'expo-status-bar';
import { ScrollView, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import LoginScreen from './LoginScreen';
import RecordScreen from './RecordScreen';
import HomeScreen from './HomeScreen';
import TimelineScreen from './TimelineScreen';
import CalendarScreen from './CalendarScreen';
import InsightsScreen from './InsightsScreen';
import { colors, radius, spacing, typography, fabShadow } from './theme';
import { Icon } from './ui';
import { useAppFonts } from './fonts';

type SignedInUser = {
  id: string;
  email: string;
  display_name: string;
};

type TabKey = 'home' | 'timeline' | 'catat' | 'kalender' | 'insight';

const TAB_LABELS: Record<TabKey, string> = {
  home: 'Home',
  timeline: 'Timeline',
  catat: 'Catat',
  kalender: 'Kalender',
  insight: 'Insight',
};

function initials(name: string): string {
  const parts = name.trim().split(/\s+/);
  return (parts[0]?.[0] ?? '?').toUpperCase() + (parts[1]?.[0] ?? '').toUpperCase();
}

export default function App() {
  const [fontsLoaded] = useAppFonts();
  const [sessionToken, setSessionToken] = useState<string | null>(null);
  const [user, setUser] = useState<SignedInUser | null>(null);
  const [activeTab, setActiveTab] = useState<TabKey>('home');
  const [refreshKey, setRefreshKey] = useState(0);

  if (!fontsLoaded) {
    return <View style={styles.app} />;
  }

  if (!sessionToken || !user) {
    return (
      <LoginScreen
        onSignedIn={(token, signedInUser) => {
          setSessionToken(token);
          setUser(signedInUser);
        }}
      />
    );
  }

  return (
    <View style={styles.app}>
      <View style={styles.header}>
        <View style={styles.brandBlock}>
          <View style={styles.logoCircle}>
            <Text style={styles.logoEmoji}>🪐</Text>
          </View>
          <View style={{ minWidth: 0 }}>
            <Text style={[typography.labelLg, { color: colors.onSurface }]} numberOfLines={1}>ORBIT</Text>
            <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]} numberOfLines={1}>Spend freely, stay aware</Text>
          </View>
        </View>

        <View style={styles.greetingBlock}>
          <View style={styles.greetingRow}>
            <Text style={[typography.labelLg, { color: colors.onSurface }]} numberOfLines={1}>
              Halo, {user.display_name.split(' ')[0]} ✨
            </Text>
            <View style={styles.statusDot} />
          </View>
          <Text style={[typography.labelSm, { color: colors.onSurfaceVariant }]}>{TAB_LABELS[activeTab]}</Text>
        </View>

        <View style={styles.headerRight}>
          <TouchableOpacity style={styles.iconButton} activeOpacity={0.7}>
            <Icon name="notifications" size={22} color={colors.onSurfaceVariant} />
          </TouchableOpacity>
          <View style={styles.avatar}>
            <Text style={styles.avatarText}>{initials(user.display_name)}</Text>
          </View>
        </View>
      </View>

      <ScrollView contentContainerStyle={styles.content} showsVerticalScrollIndicator={false}>
        {activeTab === 'home' && (
          <HomeScreen
            sessionToken={sessionToken}
            key={`home-${refreshKey}`}
            onNavigateToCatat={() => setActiveTab('catat')}
            onNavigateToTimeline={() => setActiveTab('timeline')}
          />
        )}
        {activeTab === 'timeline' && <TimelineScreen sessionToken={sessionToken} key={`timeline-${refreshKey}`} />}
        {activeTab === 'catat' && (
          <RecordScreen
            sessionToken={sessionToken}
            onSaved={() => setRefreshKey((k) => k + 1)}
          />
        )}
        {activeTab === 'kalender' && <CalendarScreen sessionToken={sessionToken} key={`kalender-${refreshKey}`} />}
        {activeTab === 'insight' && <InsightsScreen sessionToken={sessionToken} key={`insight-${refreshKey}`} />}
      </ScrollView>

      <View style={styles.navBar}>
        <NavItem tabKey="home" active={activeTab === 'home'} icon="all-inclusive" onPress={setActiveTab} />
        <NavItem tabKey="timeline" active={activeTab === 'timeline'} icon="schedule" onPress={setActiveTab} />
        <TouchableOpacity style={styles.fab} activeOpacity={0.85} onPress={() => setActiveTab('catat')}>
          <Icon name="edit" size={24} color={colors.onPrimary} />
        </TouchableOpacity>
        <NavItem tabKey="kalender" active={activeTab === 'kalender'} icon="calendar-today" onPress={setActiveTab} />
        <NavItem tabKey="insight" active={activeTab === 'insight'} icon="lightbulb" onPress={setActiveTab} />
      </View>
      <StatusBar style="dark" />
    </View>
  );
}

function NavItem({
  tabKey,
  active,
  icon,
  onPress,
}: {
  tabKey: TabKey;
  active: boolean;
  icon: React.ComponentProps<typeof Icon>['name'];
  onPress: (key: TabKey) => void;
}) {
  return (
    <TouchableOpacity style={styles.navItem} onPress={() => onPress(tabKey)} activeOpacity={0.7}>
      <Icon name={icon} size={24} color={active ? colors.primaryContainer : colors.onSurfaceVariant} />
      <Text style={[typography.labelSm, { color: active ? colors.primaryContainer : colors.onSurfaceVariant }]}>
        {TAB_LABELS[tabKey]}
      </Text>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  app: {
    flex: 1,
    backgroundColor: colors.surface,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: spacing.sm,
    paddingTop: 44,
    paddingBottom: spacing.md,
    paddingHorizontal: spacing.margin,
    backgroundColor: colors.surface,
    borderBottomWidth: 1,
    borderBottomColor: colors.surfaceContainerHigh,
  },
  brandBlock: { flexDirection: 'row', alignItems: 'center', gap: spacing.xs, flexShrink: 1 },
  logoCircle: {
    width: 30,
    height: 30,
    borderRadius: 15,
    backgroundColor: colors.primaryFixed,
    alignItems: 'center',
    justifyContent: 'center',
  },
  logoEmoji: { fontSize: 15 },
  greetingBlock: { flexShrink: 1, alignItems: 'flex-end' },
  greetingRow: { flexDirection: 'row', alignItems: 'center', gap: 6 },
  statusDot: { width: 7, height: 7, borderRadius: 4, backgroundColor: colors.secondary },
  headerRight: { flexDirection: 'row', alignItems: 'center', gap: spacing.xs },
  iconButton: { width: 40, height: 40, borderRadius: 20, alignItems: 'center', justifyContent: 'center' },
  avatar: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: colors.surfaceContainer,
    alignItems: 'center',
    justifyContent: 'center',
  },
  avatarText: { ...typography.labelMd, color: colors.primaryContainer },
  content: {
    alignItems: 'center',
    gap: spacing.lg,
    paddingTop: spacing.lg,
    paddingBottom: spacing.xl,
  },
  navBar: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-around',
    backgroundColor: colors.surfaceContainerLowest,
    borderTopWidth: 1,
    borderTopColor: colors.surfaceContainerHigh,
    paddingTop: spacing.sm,
    paddingBottom: 28,
  },
  navItem: { alignItems: 'center', gap: 2, minWidth: 56 },
  fab: {
    width: 56,
    height: 56,
    borderRadius: radius.full,
    backgroundColor: colors.primaryContainer,
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: -24,
    ...fabShadow,
  },
});
