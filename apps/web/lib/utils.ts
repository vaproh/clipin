import type { ClassValue } from "clsx"
import { clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Convert paise to ₹ display with Indian comma separators. */
export function formatPaise(amount: number | undefined | null): string {
  const val = amount ?? 0
  if (val < 0) return '-' + formatPaise(-val)
  const rupees = val / 100
  return '₹' + rupees.toLocaleString('en-IN', { maximumFractionDigits: 0 })
}

export function platformLabel(platform: string): string {
  const labels: Record<string, string> = {
    youtube: 'YouTube',
    instagram: 'Instagram',
    tiktok: 'TikTok',
    multi: 'Multi-platform',
  }
  return labels[platform] || platform
}
