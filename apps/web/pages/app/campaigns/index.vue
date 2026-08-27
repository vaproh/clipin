<script setup lang="ts">
import { Megaphone } from 'lucide-vue-next'
import type { CampaignFilters } from '~/composables/useApi'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const filters = ref<CampaignFilters>({
  page: 1,
  page_size: 20,
})

const { data, isLoading, isError } = useCampaigns(filters)

const totalPages = computed(() => {
  if (!data.value) return 0
  return Math.ceil(data.value.total / data.value.page_size)
})

function setPage(page: number) {
  filters.value = { ...filters.value, page }
}
</script>

<template>
  <div class="space-y-5">
    <SharedPageHeader title="Campaigns" description="Browse & launch performance clipping campaigns" />

    <CampaignCampaignFilters v-model="filters" />

    <!-- Search result count -->
    <p v-if="data && !isLoading" class="text-xs font-mono text-neutral-500">
      {{ filters.q ? `${data.total} campaign${data.total === 1 ? '' : 's'} found` : '' }}
    </p>

    <!-- Loading -->
    <SharedLoadingSpinner v-if="isLoading" />

    <!-- Error -->
    <div v-else-if="isError" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
      <p class="text-sm text-neutral-400">Failed to load campaigns. Try again later.</p>
    </div>

    <!-- Empty -->
    <SharedEmptyState
      v-else-if="!data?.campaigns.length"
      :icon="Megaphone"
      title="No campaigns found"
      description="No campaigns match your filters. Try adjusting your search criteria."
    />

    <!-- Campaign grid -->
    <template v-else>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        <CampaignCampaignCard
          v-for="campaign in data.campaigns"
          :key="campaign.id"
          :campaign="campaign"
        />
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex items-center justify-center gap-1 pt-4">
        <UiButton
          v-for="page in totalPages"
          :key="page"
          :variant="page === filters.page ? 'default' : 'ghost'"
          size="sm"
          class="w-8 h-8 p-0 font-mono text-xs"
          @click="setPage(page)"
        >
          {{ page }}
        </UiButton>
      </div>
    </template>
  </div>
</template>
