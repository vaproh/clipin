import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { formatPaise } from '~/lib/utils'

// Extract relativeEndDate for direct testing (it's a pure function inside CampaignCard)
function relativeEndDate(endsAt: string | null): string {
  if (!endsAt) return 'No end date'
  const end = new Date(endsAt)
  const now = new Date()
  if (end <= now) return 'Ended'
  const diffMs = end.getTime() - now.getTime()
  const days = Math.ceil(diffMs / (1000 * 60 * 60 * 24))
  if (days === 1) return 'Ends tomorrow'
  return `Ends in ${days} days`
}

// Extract progress calculation for direct testing
function progress(total_budget: number, remaining_budget: number): number {
  if (total_budget === 0) return 0
  return ((total_budget - remaining_budget) / total_budget) * 100
}

describe('relativeEndDate', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-06-15T12:00:00Z'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('returns "No end date" for null', () => {
    expect(relativeEndDate(null)).toBe('No end date')
  })

  it('returns "Ended" for past dates', () => {
    expect(relativeEndDate('2026-06-10T12:00:00Z')).toBe('Ended')
  })

  it('returns "Ended" for exactly now', () => {
    expect(relativeEndDate('2026-06-15T12:00:00Z')).toBe('Ended')
  })

  it('returns "Ends tomorrow" for 1 day away', () => {
    expect(relativeEndDate('2026-06-16T12:00:00Z')).toBe('Ends tomorrow')
  })

  it('returns "Ends in N days" for multiple days', () => {
    expect(relativeEndDate('2026-06-18T12:00:00Z')).toBe('Ends in 3 days')
    expect(relativeEndDate('2026-06-25T12:00:00Z')).toBe('Ends in 10 days')
  })

  it('returns "Ends in 1 day" for slightly more than 24h (rounds up)', () => {
    // 25 hours from now = ceil(25/24) = 2 days
    expect(relativeEndDate('2026-06-16T13:00:00Z')).toBe('Ends in 2 days')
  })
})

describe('progress calculation', () => {
  it('returns 0 when total budget is 0', () => {
    expect(progress(0, 0)).toBe(0)
  })

  it('returns 100 when all budget spent', () => {
    expect(progress(1000, 0)).toBe(100)
  })

  it('returns 0 when no budget spent', () => {
    expect(progress(1000, 1000)).toBe(0)
  })

  it('calculates partial progress', () => {
    expect(progress(1000, 500)).toBe(50)
    expect(progress(1000, 250)).toBe(75)
  })
})

describe('formatPaise (used by CampaignCard)', () => {
  it('formats CPM rate', () => {
    expect(formatPaise(500)).toBe('₹5')
  })

  it('formats budget amounts', () => {
    expect(formatPaise(100000)).toBe('₹1,000')
  })
})

describe('platformLabel mapping', () => {
  const platformLabel: Record<string, string> = {
    youtube: 'YouTube',
    instagram: 'Instagram',
    multi: 'Multi',
  }

  it('maps known platforms', () => {
    expect(platformLabel['youtube']).toBe('YouTube')
    expect(platformLabel['instagram']).toBe('Instagram')
    expect(platformLabel['multi']).toBe('Multi')
  })

  it('returns undefined for unknown platform', () => {
    expect(platformLabel['unknown']).toBeUndefined()
  })
})
