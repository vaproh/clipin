import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import LoadingSpinner from '../LoadingSpinner.vue'

describe('LoadingSpinner', () => {
  it('renders a spinner element', () => {
    const wrapper = mount(LoadingSpinner)
    expect(wrapper.find('[role="status"]').exists()).toBe(true)
  })

  it('has aria-label for accessibility', () => {
    const wrapper = mount(LoadingSpinner)
    expect(wrapper.find('[aria-label="Loading"]').exists()).toBe(true)
  })

  it('contains the spinning animation element', () => {
    const wrapper = mount(LoadingSpinner)
    expect(wrapper.find('.animate-spin').exists()).toBe(true)
  })
})
