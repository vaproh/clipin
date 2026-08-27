import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import EmptyState from '../EmptyState.vue'

const MockIcon = { template: '<span class="mock-icon">icon</span>' }

// Stub NuxtLink and UiButton since they won't resolve in test context
const stubs = {
  NuxtLink: { template: '<a><slot /></a>' },
  UiButton: { template: '<button><slot /></button>', props: ['variant', 'size'] },
}

describe('EmptyState', () => {
  it('renders title and description', () => {
    const wrapper = mount(EmptyState, {
      props: {
        icon: MockIcon,
        title: 'No campaigns',
        description: 'Create your first campaign to get started.',
      },
      global: { stubs },
    })
    expect(wrapper.text()).toContain('No campaigns')
    expect(wrapper.text()).toContain('Create your first campaign to get started.')
  })

  it('renders the icon', () => {
    const wrapper = mount(EmptyState, {
      props: {
        icon: MockIcon,
        title: 'Empty',
        description: 'Nothing here.',
      },
      global: { stubs },
    })
    expect(wrapper.find('.mock-icon').exists()).toBe(true)
  })

  it('does not render action link when actionLabel is omitted', () => {
    const wrapper = mount(EmptyState, {
      props: {
        icon: MockIcon,
        title: 'Empty',
        description: 'Nothing here.',
      },
      global: { stubs },
    })
    expect(wrapper.find('a').exists()).toBe(false)
  })

  it('renders action link when both actionLabel and actionTo are provided', () => {
    const wrapper = mount(EmptyState, {
      props: {
        icon: MockIcon,
        title: 'Empty',
        description: 'Nothing here.',
        actionLabel: 'Create campaign',
        actionTo: '/app/campaigns/new',
      },
      global: { stubs },
    })
    expect(wrapper.text()).toContain('Create campaign')
    expect(wrapper.find('a').exists()).toBe(true)
  })
})
