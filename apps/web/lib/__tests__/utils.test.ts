import { describe, it, expect } from 'vitest'
import { formatPaise, cn } from '../utils'

describe('formatPaise', () => {
  it('formats zero paise', () => {
    expect(formatPaise(0)).toBe('₹0')
  })

  it('formats whole rupees', () => {
    expect(formatPaise(100)).toBe('₹1')
    expect(formatPaise(5000)).toBe('₹50')
  })

  it('formats with Indian comma separators (lakhs)', () => {
    expect(formatPaise(100000)).toBe('₹1,000')
  })

  it('formats with Indian comma separators (crores)', () => {
    // 10000000 paise = ₹1,00,000 (1 lakh)
    expect(formatPaise(10000000)).toBe('₹1,00,000')
  })

  it('rounds fractional paise down to whole rupees', () => {
    // 150 paise = ₹1.50, maximumFractionDigits: 0 rounds to ₹2
    expect(formatPaise(150)).toBe('₹2')
  })

  it('handles small amounts', () => {
    expect(formatPaise(50)).toBe('₹1')
    expect(formatPaise(1)).toBe('₹0')
  })
})

describe('cn', () => {
  it('merges class names', () => {
    const result = cn('text-white', 'text-black')
    expect(result).toBe('text-black')
  })

  it('handles conditional classes', () => {
    const result = cn('base', false && 'hidden', 'extra')
    expect(result).toContain('base')
    expect(result).toContain('extra')
    expect(result).not.toContain('hidden')
  })

  it('returns empty string for no inputs', () => {
    expect(cn()).toBe('')
  })
})
