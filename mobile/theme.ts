// Design tokens ported directly from the Google Stitch mockups
// (Kalender + Catat screens, TASK-015 visual design pass) - a
// Material 3-style token set: indigo primary, mint-green secondary
// (for "under baseline" / positive states), warm amber-brown tertiary
// (for "over baseline" - deliberately not red/alarming, per the
// non-punitive design principle), on a soft lavender-white surface.
// Values are copied verbatim from the mockups' Tailwind config so the
// native app matches the approved design exactly.

export const colors = {
  primary: '#4143D5',
  primaryContainer: '#5B5FEF',
  onPrimary: '#FFFFFF',
  onPrimaryContainer: '#F9F6FF',
  primaryFixed: '#E1E0FF',
  primaryFixedDim: '#C0C1FF',
  onPrimaryFixed: '#05006C',
  onPrimaryFixedVariant: '#2C2CC3',

  secondary: '#006C49',
  secondaryContainer: '#6CF8BB',
  onSecondary: '#FFFFFF',
  onSecondaryContainer: '#00714D',
  secondaryFixed: '#6FFBBE',
  secondaryFixedDim: '#4EDEA3',
  onSecondaryFixed: '#002113',
  onSecondaryFixedVariant: '#005236',

  tertiary: '#8C4600',
  tertiaryContainer: '#B15A00',
  onTertiary: '#FFFFFF',
  onTertiaryContainer: '#FFF6F2',
  tertiaryFixed: '#FFDCC5',
  tertiaryFixedDim: '#FFB783',
  onTertiaryFixed: '#301400',
  onTertiaryFixedVariant: '#713700',

  error: '#BA1A1A',
  errorContainer: '#FFDAD6',
  onError: '#FFFFFF',
  onErrorContainer: '#93000A',

  background: '#FCF8FF',
  onBackground: '#181445',
  surface: '#FCF8FF',
  surfaceDim: '#DAD6FF',
  surfaceBright: '#FCF8FF',
  surfaceContainerLowest: '#FFFFFF',
  surfaceContainerLow: '#F6F2FF',
  surfaceContainer: '#EFEBFF',
  surfaceContainerHigh: '#E9E5FF',
  surfaceContainerHighest: '#E3DFFF',
  surfaceVariant: '#E3DFFF',
  onSurface: '#181445',
  onSurfaceVariant: '#464555',
  inverseSurface: '#2D2A5B',
  inverseOnSurface: '#F3EEFF',
  inversePrimary: '#C0C1FF',
  outline: '#767586',
  outlineVariant: '#C6C5D7',
};

export const spacing = {
  xs: 4,
  sm: 8,
  md: 16, // "gutter" / "space-md" in the mockup
  margin: 20,
  lg: 24,
  xl: 32,
};

export const radius = {
  DEFAULT: 16,
  lg: 32,
  xl: 48,
  full: 999,
};

// Plus Jakarta Sans (the mockup's only typeface, used across every
// role) loaded via @expo-google-fonts - see fonts.ts for the useFonts
// call. React Native ignores fontWeight on a custom-loaded font, so
// each role picks its weight via fontFamily instead.
export const fonts = {
  regular: 'PlusJakartaSans_400Regular',
  medium: 'PlusJakartaSans_500Medium',
  semibold: 'PlusJakartaSans_600SemiBold',
  bold: 'PlusJakartaSans_700Bold',
};

// Font sizes/weights/line-heights ported from the mockup's Tailwind
// fontSize scale. React Native has no letter-spacing-by-em, so tracking
// values are converted to approximate px at each size.
export const typography = {
  headlineXl: { fontFamily: fonts.bold, fontSize: 32, lineHeight: 40, letterSpacing: -0.6 },
  headlineLg: { fontFamily: fonts.bold, fontSize: 26, lineHeight: 34, letterSpacing: -0.4 },
  headlineMd: { fontFamily: fonts.semibold, fontSize: 20, lineHeight: 28, letterSpacing: -0.2 },
  headlineSm: { fontFamily: fonts.semibold, fontSize: 18, lineHeight: 24 },
  currencyDisplay: { fontFamily: fonts.bold, fontSize: 28, lineHeight: 36, letterSpacing: -0.5 },
  currencyMd: { fontFamily: fonts.semibold, fontSize: 16, lineHeight: 22, letterSpacing: -0.15 },
  bodyLg: { fontFamily: fonts.regular, fontSize: 16, lineHeight: 24 },
  bodyMd: { fontFamily: fonts.regular, fontSize: 14, lineHeight: 22 },
  bodySm: { fontFamily: fonts.regular, fontSize: 13, lineHeight: 18 },
  labelLg: { fontFamily: fonts.semibold, fontSize: 14, lineHeight: 20 },
  labelMd: { fontFamily: fonts.semibold, fontSize: 12, lineHeight: 16, letterSpacing: 0.1 },
  labelSm: { fontFamily: fonts.medium, fontSize: 11, lineHeight: 14, letterSpacing: 0.2 },
};

// Soft, color-tinted elevation used on every Card - matches the
// mockup's `shadow-[0_8px_24px_-4px_rgba(91,95,239,0.06)...]`.
export const cardShadow = {
  shadowColor: colors.primaryContainer,
  shadowOffset: { width: 0, height: 8 },
  shadowOpacity: 0.08,
  shadowRadius: 20,
  elevation: 3,
};

// Consumption Bank badge thresholds (mirrors backend badgeTierFor,
// TASK-007) - used client-side only to compute "how far to the next
// tier", never to decide the badge itself (that always comes from the
// server's own `badge` field).
export const BADGE_TIERS: { name: string; threshold: number }[] = [
  { name: 'Frugal', threshold: 100000 },
  { name: 'Thrifty', threshold: 200000 },
  { name: 'Economical', threshold: 300000 },
  { name: 'Collector', threshold: 400000 },
  { name: 'Master', threshold: 500000 },
];

export function nextBadgeTier(balance: number): { name: string; threshold: number } | null {
  return BADGE_TIERS.find((t) => balance < t.threshold) ?? null;
}

export function currentTierFloor(balance: number): number {
  const reached = [...BADGE_TIERS].reverse().find((t) => balance >= t.threshold);
  return reached ? reached.threshold : 0;
}

export const fabShadow = {
  shadowColor: colors.primaryContainer,
  shadowOffset: { width: 0, height: 10 },
  shadowOpacity: 0.35,
  shadowRadius: 18,
  elevation: 8,
};
