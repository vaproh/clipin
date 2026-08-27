<script setup lang="ts">
import { Search, X } from 'lucide-vue-next'
import { useDebounceFn } from '@vueuse/core'
import type { CampaignFilters } from '~/composables/useApi'

const filters = defineModel<CampaignFilters>({ required: true })

const searchInput = ref(filters.value.q ?? '')

const debouncedSearch = useDebounceFn((value: string) => {
  filters.value = { ...filters.value, q: value || undefined, page: 1 }
}, 300)

watch(searchInput, (v) => debouncedSearch(v))

function clearSearch() {
  searchInput.value = ''
}

// Platform select needs string, not undefined. Use '' for "all".
const platformValue = computed({
  get: () => filters.value.platform ?? '',
  set: (v: string) => {
    filters.value = { ...filters.value, platform: v || undefined, page: 1 }
  },
})

// Inputs need string values
const cpmValue = computed({
  get: () => filters.value.max_cpm?.toString() ?? '',
  set: (v: string) => {
    filters.value = { ...filters.value, max_cpm: v ? Number(v) : undefined, page: 1 }
  },
})

const budgetValue = computed({
  get: () => filters.value.min_budget?.toString() ?? '',
  set: (v: string) => {
    filters.value = { ...filters.value, min_budget: v ? Number(v) : undefined, page: 1 }
  },
})

function clearFilters() {
  searchInput.value = ''
  filters.value = { page: 1, page_size: 20 }
}
</script>

<template>
  <div class="space-y-3">
    <!-- Search -->
    <div class="relative">
      <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-neutral-500 pointer-events-none" />
      <input
        v-model="searchInput"
        type="text"
        placeholder="Search campaigns..."
        class="h-8 w-full rounded-lg border border-neutral-800 bg-neutral-950 pl-8 pr-8 py-1 text-base md:text-sm text-neutral-200 placeholder:text-neutral-500 outline-none focus-visible:border-neutral-600 transition-colors"
      />
      <button
        v-if="searchInput"
        type="button"
        class="absolute right-2 top-1/2 -translate-y-1/2 text-neutral-500 hover:text-neutral-300 transition-colors"
        @click="clearSearch"
      >
        <X class="h-4 w-4" />
      </button>
    </div>

    <!-- Filters row -->
    <div class="flex flex-wrap items-center gap-3">
      <!-- Platform -->
      <UiSelect v-model="platformValue">
        <UiSelectTrigger class="w-full sm:w-40">
          <UiSelectValue placeholder="All platforms" />
        </UiSelectTrigger>
        <UiSelectContent>
          <UiSelectItem value="">All platforms</UiSelectItem>
          <UiSelectItem value="youtube">YouTube</UiSelectItem>
          <UiSelectItem value="instagram">Instagram</UiSelectItem>
          <UiSelectItem value="tiktok">TikTok</UiSelectItem>
          <UiSelectItem value="multi">Multi</UiSelectItem>
        </UiSelectContent>
      </UiSelect>

      <!-- Max CPM -->
      <div class="flex items-center gap-1.5">
        <span class="text-[11px] font-mono text-neutral-500">Max CPM</span>
        <UiInput
          v-model="cpmValue"
          type="number"
          placeholder="₹"
          class="w-24"
        />
      </div>

      <!-- Min budget -->
      <div class="flex items-center gap-1.5">
        <span class="text-[11px] font-mono text-neutral-500">Min budget</span>
        <UiInput
          v-model="budgetValue"
          type="number"
          placeholder="₹"
          class="w-28"
        />
      </div>

      <!-- Clear -->
      <UiButton variant="ghost" size="sm" class="text-neutral-400" @click="clearFilters">
        Clear
      </UiButton>
    </div>
  </div>
</template>
