// Categories are free-text, user-created with no icon field in the
// schema (TASK-004's "no normalization" requirement) - this is a
// best-effort decorative guess from the name, never a claim about the
// category's real classification.
const CATEGORY_EMOJI: [RegExp, string][] = [
  [/kopi|nongkrong|cafe|kafe/i, '☕'],
  [/makan|nasi|siang|malam/i, '🍜'],
  [/transport|ojol|bensin|motor/i, '🚌'],
  [/jajan|cemilan|camilan|snack/i, '🍿'],
  [/buku|reward|hobi/i, '📚'],
];

export function categoryEmoji(name: string): string {
  return CATEGORY_EMOJI.find(([re]) => re.test(name))?.[1] ?? '🛍️';
}
