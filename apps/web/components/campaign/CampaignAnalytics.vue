<script setup lang="ts">
import { Users, Eye, ThumbsUp, MessageSquare, Share2, BarChart3, Clock, Wallet } from 'lucide-vue-next'
import { formatPaise } from '~/lib/utils'
import type { CampaignAnalytics } from '~/composables/useApi'

const props = defineProps<{
  analytics: CampaignAnalytics
  totalBudget: number
}>()

const approvalRate = computed(() => {
  const { total, approved } = props.analytics.submissions
  if (total === 0) return 0
  return Math.round((approved / total) * 100)
})

const budgetConsumed = computed(() => Math.round(props.analytics.progress.budget_consumed_pct))
</script>

<template>
  <div class="space-y-3">
    <!-- Submissions -->
    <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
      <SharedStatCard
        label="Total submissions"
        :value="analytics.submissions.total"
        :icon="BarChart3"
      />
      <SharedStatCard
        label="Approved"
        :value="analytics.submissions.approved"
        :hint="`${approvalRate}% approval rate`"
      />
      <SharedStatCard
        label="Unique clippers"
        :value="analytics.submissions.unique_clippers"
        :icon="Users"
      />
    </div>

    <!-- Views -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <SharedStatCard
        label="Total views"
        :value="analytics.views.total_views.toLocaleString('en-IN')"
        :icon="Eye"
      />
      <SharedStatCard
        label="Likes"
        :value="analytics.views.total_likes.toLocaleString('en-IN')"
        :icon="ThumbsUp"
      />
      <SharedStatCard
        label="Comments"
        :value="analytics.views.total_comments.toLocaleString('en-IN')"
        :icon="MessageSquare"
      />
      <SharedStatCard
        label="Shares"
        :value="analytics.views.total_shares.toLocaleString('en-IN')"
        :icon="Share2"
      />
    </div>

    <!-- Financial -->
    <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
      <SharedStatCard
        label="Paid to clippers"
        :value="formatPaise(analytics.financial.total_earnings)"
        :icon="Wallet"
      />
      <SharedStatCard
        label="Platform fee"
        :value="formatPaise(analytics.financial.total_fees)"
      />
      <SharedStatCard
        label="Refunds"
        :value="formatPaise(analytics.financial.total_refunds)"
      />
    </div>

    <!-- Budget progress -->
    <div class="rounded bg-neutral-950 border border-neutral-800 p-4 space-y-3">
      <div class="flex items-center justify-between">
        <span class="text-[11px] font-mono uppercase tracking-wider text-neutral-500">Budget</span>
        <span class="text-[11px] font-mono text-neutral-400">{{ budgetConsumed }}% consumed</span>
      </div>
      <div class="h-2 w-full rounded-full bg-neutral-800 overflow-hidden">
        <div class="h-full rounded-full bg-white" :style="{ width: `${budgetConsumed}%` }" />
      </div>
      <div class="flex items-center justify-between text-xs font-mono text-neutral-400">
        <span>{{ formatPaise(totalBudget) }} total</span>
        <span v-if="analytics.progress.time_remaining" class="flex items-center gap-1.5">
          <Clock class="w-3.5 h-3.5" />
          {{ analytics.progress.time_remaining }}
        </span>
      </div>
    </div>
  </div>
</template>
