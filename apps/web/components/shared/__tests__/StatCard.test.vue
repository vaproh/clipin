import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import StatCard from '../StatCard.vue'

describe('StatCard', () => {
  it('renders label and value', () => {
    const wrapper = mount(StatCard, {
      props: { label: 'Revenue', value: '₹1,000' },
    })
    expect(wrapper.text()).toContain('Revenue')
    expect(wrapper.text()).toContain('₹1,000')
  })

  it('renders hint when provided', () => {
    const wrapper = mount(StatCard, {
      props: { label: 'Revenue', value: '₹1,000', hint: 'per month' },
    })
    expect(wrapper.text()).toContain('per month')
  })

  it('does not render hint when omitted', () => {
    const wrapper = mount(StatCard, {
      props: { label: 'Revenue', value: '₹1,000' },
    })
    expect(wrapper.text()).not.toContain('per month')
  })

  it('renders icon when provided', () => {
    const Icon = { template: '<span class="icon">I</span>' }
    const wrapper = mount(StatCard, {
      props: { label: 'Revenue', value: '₹1,000', icon: Icon },
    })
    expect(wrapper.find('.icon').exists()).toBe(true)
  })

  it('does not render icon when omitted', () => {
    const wrapper = mount(StatCard, {
      props: { label: 'Revenue', value: '₹1,000' },
    })
    expect(wrapper.find('svg, .icon, [class*="icon"]').exists()).toBe(false)
  })

  it('renders numeric value', () => {
    const wrapper = mount(StatCard, {
      props: { label: 'Count', value: 42 },
    })
    expect(wrapper.text()).toContain('42')
  })
})
