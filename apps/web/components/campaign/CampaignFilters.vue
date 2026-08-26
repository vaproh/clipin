<script setup lang="ts">
import type { CampaignFilters } from '~/composables/useApi'

const filters = defineModel<CampaignFilters>({ required: true })

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
  filters.value = { page: 1, page_size: 20 }
}
</script>

<template>
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
</template>
