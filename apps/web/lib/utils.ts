import type { ClassValue } from "clsx"
import { clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Convert paise to ₹ display with Indian comma separators. */
export function formatPaise(amount: number): string {
  const rupees = amount / 100
  return '₹' + rupees.toLocaleString('en-IN', { maximumFractionDigits: 0 })
}
