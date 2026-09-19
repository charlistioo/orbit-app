// Shared presentational components implementing the Stitch design
// system (TASK-015 visual design pass) - Material 3-style cards, pill
// buttons/chips, and status indicators. Icon() wraps @expo/vector-icons'
// MaterialIcons, which uses the same icon names as the mockups'
// "material-symbols-outlined" font, so icon names carry over 1:1.
import { ReactNode } from 'react';
import {
  ActivityIndicator,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
  ViewStyle,
} from 'react-native';
import { MaterialIcons } from '@expo/vector-icons';
import Svg, { Circle } from 'react-native-svg';
import { colors, radius, spacing, typography, cardShadow } from './theme';

export function Icon({
  name,
  size = 20,
  color = colors.onSurfaceVariant,
}: {
  name: keyof typeof MaterialIcons.glyphMap;
  size?: number;
  color?: string;
}) {
  return <MaterialIcons name={name} size={size} color={color} />;
}

export function ScreenContainer({ children }: { children: ReactNode }) {
  return <View style={styles.screen}>{children}</View>;
}

export function Card({ children, style }: { children: ReactNode; style?: ViewStyle }) {
  return <View style={[styles.card, style]}>{children}</View>;
}

export function SectionTitle({ children, icon }: { children: ReactNode; icon?: keyof typeof MaterialIcons.glyphMap }) {
  return (
    <View style={styles.sectionTitleRow}>
      {icon ? <Icon name={icon} size={20} color={colors.primaryContainer} /> : null}
      <Text style={[typography.headlineSm, { color: colors.onSurface }]}>{children}</Text>
    </View>
  );
}

export function LoadingState({ label = 'Memuat...' }: { label?: string }) {
  return (
    <View style={styles.centered}>
      <ActivityIndicator color={colors.primaryContainer} />
      <Text style={[typography.bodyMd, { color: colors.onSurfaceVariant, marginTop: spacing.sm }]}>{label}</Text>
    </View>
  );
}

export function ErrorState({ message }: { message: string }) {
  return (
    <View style={styles.errorBox}>
      <Text style={[typography.bodyMd, { color: colors.onErrorContainer }]}>{message}</Text>
    </View>
  );
}

export function PrimaryButton({
  title,
  onPress,
  disabled,
  variant = 'primary',
  icon,
}: {
  title: string;
  onPress: () => void;
  disabled?: boolean;
  variant?: 'primary' | 'secondary';
  icon?: keyof typeof MaterialIcons.glyphMap;
}) {
  return (
    <TouchableOpacity
      onPress={onPress}
      disabled={disabled}
      activeOpacity={0.85}
      style={[
        styles.button,
        variant === 'secondary' && styles.buttonSecondary,
        disabled && styles.buttonDisabled,
      ]}
    >
      {icon ? (
        <Icon name={icon} size={18} color={variant === 'secondary' ? colors.primaryContainer : colors.onPrimary} />
      ) : null}
      <Text style={variant === 'secondary' ? styles.buttonSecondaryText : styles.buttonText}>{title}</Text>
    </TouchableOpacity>
  );
}

export function Chip({
  label,
  selected,
  onPress,
  emoji,
}: {
  label: string;
  selected?: boolean;
  onPress?: () => void;
  emoji?: string;
}) {
  const Wrapper = onPress ? TouchableOpacity : View;
  return (
    <Wrapper style={[styles.chip, selected && styles.chipSelected]} onPress={onPress} activeOpacity={0.8}>
      {emoji ? <Text style={styles.chipEmoji}>{emoji}</Text> : null}
      <Text style={selected ? styles.chipTextSelected : styles.chipText}>{label}</Text>
    </Wrapper>
  );
}

export function StatusPill({
  label,
  tone = 'secondary',
  icon,
}: {
  label: string;
  tone?: 'secondary' | 'tertiary' | 'neutral';
  icon?: keyof typeof MaterialIcons.glyphMap;
}) {
  const toneStyle =
    tone === 'secondary' ? styles.pillSecondary : tone === 'tertiary' ? styles.pillTertiary : styles.pillNeutral;
  const textColor = tone === 'secondary' ? colors.onSecondaryContainer : tone === 'tertiary' ? colors.tertiary : colors.onSurfaceVariant;
  return (
    <View style={[styles.pill, toneStyle]}>
      {icon ? <Icon name={icon} size={13} color={textColor} /> : null}
      <Text style={[typography.labelSm, { color: textColor, fontFamily: typography.labelMd.fontFamily }]}>{label}</Text>
    </View>
  );
}

export function IconCircle({
  icon,
  emoji,
  size = 40,
  bg = colors.surfaceContainer,
  color = colors.primaryContainer,
}: {
  icon?: keyof typeof MaterialIcons.glyphMap;
  emoji?: string;
  size?: number;
  bg?: string;
  color?: string;
}) {
  return (
    <View style={[styles.iconCircle, { width: size, height: size, borderRadius: size / 2, backgroundColor: bg }]}>
      {emoji ? <Text style={{ fontSize: size * 0.45 }}>{emoji}</Text> : icon ? <Icon name={icon} size={size * 0.5} color={color} /> : null}
    </View>
  );
}

export function SegmentedControl<T extends string>({
  options,
  value,
  onChange,
}: {
  options: { key: T; label: string }[];
  value: T;
  onChange: (key: T) => void;
}) {
  return (
    <View style={styles.segmented}>
      {options.map((opt) => {
        const active = opt.key === value;
        return (
          <TouchableOpacity
            key={opt.key}
            style={[styles.segmentedItem, active && styles.segmentedItemActive]}
            onPress={() => onChange(opt.key)}
            activeOpacity={0.85}
          >
            <Text style={[typography.labelMd, { color: active ? colors.primaryContainer : colors.onSurfaceVariant, textAlign: 'center' }]}>
              {opt.label}
            </Text>
          </TouchableOpacity>
        );
      })}
    </View>
  );
}

// DonutProgress draws a ring gauge (used on Home for "Sisa Hari Ini") -
// the fraction itself always comes from real server-reported numbers,
// this component only draws it.
export function DonutProgress({
  progress,
  size = 168,
  strokeWidth = 14,
  color = colors.primaryContainer,
  trackColor = colors.surfaceContainerHigh,
  children,
}: {
  progress: number;
  size?: number;
  strokeWidth?: number;
  color?: string;
  trackColor?: string;
  children?: ReactNode;
}) {
  const clamped = Math.max(0, Math.min(1, progress));
  const radiusPx = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radiusPx;
  const dashOffset = circumference * (1 - clamped);
  return (
    <View style={{ width: size, height: size, alignItems: 'center', justifyContent: 'center' }}>
      <Svg width={size} height={size}>
        <Circle
          cx={size / 2}
          cy={size / 2}
          r={radiusPx}
          stroke={trackColor}
          strokeWidth={strokeWidth}
          fill="none"
        />
        <Circle
          cx={size / 2}
          cy={size / 2}
          r={radiusPx}
          stroke={color}
          strokeWidth={strokeWidth}
          fill="none"
          strokeDasharray={`${circumference} ${circumference}`}
          strokeDashoffset={dashOffset}
          strokeLinecap="round"
          rotation="-90"
          origin={`${size / 2}, ${size / 2}`}
        />
      </Svg>
      <View style={{ position: 'absolute', alignItems: 'center' }}>{children}</View>
    </View>
  );
}

export function ProgressBar({ progress, color = colors.primaryContainer }: { progress: number; color?: string }) {
  const pct = Math.max(0, Math.min(1, progress));
  return (
    <View style={styles.progressTrack}>
      <View style={[styles.progressFill, { width: `${pct * 100}%`, backgroundColor: color }]} />
    </View>
  );
}

const styles = StyleSheet.create({
  screen: {
    width: '100%',
    maxWidth: 480,
    alignSelf: 'center',
    paddingHorizontal: spacing.margin,
    gap: spacing.lg,
  },
  card: {
    backgroundColor: colors.surfaceContainerLowest,
    borderRadius: radius.lg,
    padding: spacing.lg,
    gap: spacing.md,
    width: '100%',
    ...cardShadow,
  },
  sectionTitleRow: { flexDirection: 'row', alignItems: 'center', gap: spacing.xs },
  centered: { alignItems: 'center', justifyContent: 'center', paddingVertical: spacing.xl },
  errorBox: {
    backgroundColor: colors.errorContainer,
    borderRadius: radius.DEFAULT,
    padding: spacing.md,
  },
  button: {
    flexDirection: 'row',
    backgroundColor: colors.primaryContainer,
    borderRadius: radius.full,
    paddingVertical: 12,
    paddingHorizontal: spacing.lg,
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.xs,
  },
  buttonSecondary: {
    backgroundColor: colors.surfaceContainer,
  },
  buttonDisabled: {
    opacity: 0.5,
  },
  buttonText: { color: colors.onPrimary, fontFamily: typography.labelLg.fontFamily, fontSize: 15 },
  buttonSecondaryText: { color: colors.primaryContainer, fontFamily: typography.labelLg.fontFamily, fontSize: 15 },
  chip: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    backgroundColor: colors.surfaceContainerHigh,
    borderRadius: radius.full,
    paddingHorizontal: 14,
    paddingVertical: 8,
    marginRight: spacing.xs,
  },
  chipSelected: {
    backgroundColor: colors.primaryContainer,
  },
  chipEmoji: { fontSize: 14 },
  chipText: { ...typography.labelMd, color: colors.onSurfaceVariant },
  chipTextSelected: { ...typography.labelMd, color: colors.onPrimary },
  pill: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    alignSelf: 'flex-start',
    borderRadius: radius.full,
    paddingHorizontal: 10,
    paddingVertical: 4,
  },
  pillSecondary: { backgroundColor: colors.secondaryContainer },
  pillTertiary: { backgroundColor: colors.tertiaryFixed },
  pillNeutral: { backgroundColor: colors.surfaceContainer },
  iconCircle: { alignItems: 'center', justifyContent: 'center' },
  segmented: {
    flexDirection: 'row',
    padding: 4,
    backgroundColor: colors.surfaceContainerHigh,
    borderRadius: radius.full,
    width: '100%',
    maxWidth: 320,
    alignSelf: 'center',
  },
  segmentedItem: {
    flex: 1,
    paddingVertical: 8,
    paddingHorizontal: 10,
    borderRadius: radius.full,
  },
  segmentedItemActive: {
    backgroundColor: colors.surfaceContainerLowest,
  },
  progressTrack: {
    width: '100%',
    height: 10,
    backgroundColor: colors.surfaceContainer,
    borderRadius: radius.full,
    overflow: 'hidden',
  },
  progressFill: {
    height: '100%',
    borderRadius: radius.full,
  },
});
