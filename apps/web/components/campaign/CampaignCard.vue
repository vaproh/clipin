<script setup lang="ts">
import { Motion } from 'motion-v'
import type { Campaign } from '~/composables/useApi'
import { formatPaise, platformLabel } from '~/lib/utils'

const props = defineProps<{
  campaign: Campaign
}>()

const progress = computed(() => {
  const { total_budget, remaining_budget } = props.campaign
  if (total_budget === 0) return 0
  return ((total_budget - remaining_budget) / total_budget) * 100
})

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
</script>

<template>
  <NuxtLink :to="`/app/campaigns/${campaign.id}`" class="block">
    <Motion
      :initial="{ opacity: 0, y: 8 }"
      :animate="{ opacity: 1, y: 0 }"
      :hover="{ y: -2 }"
      :transition="{ duration: 0.2 }"
    >
      <div class="rounded-lg bg-neutral-950 border border-neutral-800 p-4 space-y-3 cursor-pointer hover:border-neutral-700 transition-colors">
        <!-- Header: title + badges -->
        <div class="flex items-start justify-between gap-2">
          <h3 class="text-sm font-semibold text-white truncate">{{ campaign.title }}</h3>
          <div class="flex items-center gap-1.5 shrink-0">
            <span class="text-[9px] font-mono px-1.5 py-0.5 rounded bg-neutral-900 text-neutral-400 border border-neutral-800">
              {{ platformLabel(campaign.platform) }}
            </span>
            <SharedStatusBadge :status="campaign.status" />
          </div>
        </div>

        <!-- Description -->
        <p v-if="campaign.description" class="text-xs text-neutral-400 line-clamp-2 leading-relaxed">
          {{ campaign.description }}
        </p>

        <!-- Budget progress -->
        <div class="space-y-1.5">
          <div class="h-1 w-full rounded-full bg-neutral-800 overflow-hidden">
            <div class="h-full rounded-full bg-white" :style="{ width: `${progress}%` }" />
          </div>
          <div class="flex items-center justify-between text-[11px] font-mono text-neutral-500">
            <span>{{ formatPaise(campaign.remaining_budget) }} remaining</span>
            <span>{{ formatPaise(campaign.total_budget) }} total</span>
          </div>
        </div>

        <!-- Stats row -->
        <div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] font-mono text-neutral-500">
          <span>CPM {{ formatPaise(campaign.cpm_rate) }}</span>
          <span class="text-neutral-800">|</span>
          <span>Min {{ campaign.min_views_per_clip.toLocaleString('en-IN') }} views</span>
          <span class="text-neutral-800">|</span>
          <span>{{ campaign.max_clips_per_clipper }} clips/clipper</span>
        </div>

        <!-- End date -->
        <div class="text-[11px] font-mono text-neutral-500">
          {{ relativeEndDate(campaign.ends_at) }}
        </div>
      </div>
    </Motion>
  </NuxtLink>
</template>
