// Tag colors are palette keys stored by the API (see TagColors in
// app/services/wealth/tags_activities.go). Class names stay literal so
// Tailwind can find them.
const STYLES = {
  gray: { badge: 'bg-slate-500/15 text-slate-700 ring-slate-500/30 dark:text-slate-300', dot: 'bg-slate-500' },
  red: { badge: 'bg-red-500/15 text-red-700 ring-red-500/30 dark:text-red-300', dot: 'bg-red-500' },
  orange: { badge: 'bg-orange-500/15 text-orange-700 ring-orange-500/30 dark:text-orange-300', dot: 'bg-orange-500' },
  amber: { badge: 'bg-amber-500/15 text-amber-700 ring-amber-500/30 dark:text-amber-300', dot: 'bg-amber-500' },
  green: { badge: 'bg-green-500/15 text-green-700 ring-green-500/30 dark:text-green-300', dot: 'bg-green-500' },
  teal: { badge: 'bg-teal-500/15 text-teal-700 ring-teal-500/30 dark:text-teal-300', dot: 'bg-teal-500' },
  blue: { badge: 'bg-blue-500/15 text-blue-700 ring-blue-500/30 dark:text-blue-300', dot: 'bg-blue-500' },
  violet: { badge: 'bg-violet-500/15 text-violet-700 ring-violet-500/30 dark:text-violet-300', dot: 'bg-violet-500' },
  pink: { badge: 'bg-pink-500/15 text-pink-700 ring-pink-500/30 dark:text-pink-300', dot: 'bg-pink-500' },
}

export const TAG_COLORS = Object.keys(STYLES)

export function tagStyle(color) {
  return STYLES[color] || STYLES.gray
}
