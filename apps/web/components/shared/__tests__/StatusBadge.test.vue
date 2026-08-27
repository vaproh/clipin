import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import StatusBadge from '../StatusBadge.vue'

describe('StatusBadge', () => {
  it('renders the status text', () => {
    const wrapper = mount(StatusBadge, { props: { status: 'active' } })
    expect(wrapper.text()).toContain('active')
  })

  it('applies high-contrast tone for "active"', () => {
    const wrapper = mount(StatusBadge, { props: { status: 'active' } })
    expect(wrapper.find('span').classes()).toContain('bg-white')
  })

  it('applies default tone for unknown status', () => {
    const wrapper = mount(StatusBadge, { props: { status: 'unknown_thing' } })
    const badge = wrapper.findAll('span')[0]
    expect(badge.classes()).toContain('bg-neutral-900')
  })

  it('renders a dot indicator', () => {
    const wrapper = mount(StatusBadge, { props: { status: 'pending' } })
    const dot = wrapper.find('.rounded-full')
    expect(dot.exists()).toBe(true)
  })

  it.each([
    ['approved', 'bg-neutral-800'],
    ['rejected', 'bg-neutral-950'],
    ['pending', 'bg-neutral-900'],
    ['draft', 'bg-neutral-950'],
    ['completed', 'bg-white'],
  ])('applies correct tone for "%s"', (status, expectedBg) => {
    const wrapper = mount(StatusBadge, { props: { status } })
    expect(wrapper.find('span').classes()).toContain(expectedBg)
  })
})
