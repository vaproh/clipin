import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import PageHeader from '../PageHeader.vue'

describe('PageHeader', () => {
  it('renders title', () => {
    const wrapper = mount(PageHeader, {
      props: { title: 'Campaign Dashboard' },
    })
    expect(wrapper.text()).toContain('Campaign Dashboard')
  })

  it('renders description when provided', () => {
    const wrapper = mount(PageHeader, {
      props: { title: 'Dashboard', description: 'Overview of your campaigns' },
    })
    expect(wrapper.text()).toContain('Overview of your campaigns')
  })

  it('does not render description when omitted', () => {
    const wrapper = mount(PageHeader, {
      props: { title: 'Dashboard' },
    })
    expect(wrapper.find('p').exists()).toBe(false)
  })
})
