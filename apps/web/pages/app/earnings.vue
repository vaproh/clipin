<script setup lang="ts">
import { Wallet, Eye, CheckCircle, TrendingUp, ArrowRight, BarChart3 } from 'lucide-vue-next'
import { formatPaise, platformLabel } from '~/lib/utils'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const { data: analytics, isLoading, isError } = useMyAnalytics()

const hasData = computed(() => {
  if (!analytics.value) return false
  return (
    analytics.value.summary.total_earnings > 0 ||
    analytics.value.earnings_by_day.length > 0 ||
    analytics.value.earnings_by_campaign.length > 0
  )
})

function formatNumber(n: number): string {
  return n.toLocaleString('en-IN')
}

function formatViews(n: number): string {
  if (n >= 100_000) return (n / 100_000).toFixed(1) + 'L'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K'
  return formatNumber(n)
}

// Bar chart: normalize earnings to max 100% height
const chartBars = computed(() => {
  if (!analytics.value) return []
  const days = analytics.value.earnings_by_day
  const maxTotal = Math.max(...days.map((d) => d.total), 1)
  return days.map((d) => ({
    ...d,
    height: Math.max((d.total / maxTotal) * 100, 2),
  }))
})

const hoveredBar = ref<{ day: string; total: number } | null>(null)
</script>

<template>
  <div class="space-y-6">
    <SharedPageHeader title="Analytics" description="Your earnings trends, campaign performance, and clip stats" />

    <SharedLoadingSpinner v-if="isLoading" />

    <div v-else-if="isError" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
      <p class="text-sm text-red-400 font-mono">Failed to load analytics. Please try again.</p>
    </div>

    <template v-else-if="hasData && analytics">
      <!-- Summary stats row -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <SharedStatCard
          label="Total earnings"
          :value="formatPaise(analytics.summary.total_earnings)"
          :icon="Wallet"
        />
        <SharedStatCard
          label="Total views"
          :value="formatNumber(analytics.summary.total_views)"
          :icon="Eye"
        />
        <SharedStatCard
          label="Approval rate"
          :value="analytics.summary.approval_rate.toFixed(1) + '%'"
          :icon="CheckCircle"
        />
        <SharedStatCard
          label="Avg per clip"
          :value="formatPaise(analytics.summary.avg_earnings_per_clip)"
          :icon="TrendingUp"
        />
      </div>

      <!-- Earnings by day chart -->
      <div v-if="chartBars.length > 0" class="rounded bg-neutral-950 border border-neutral-800 p-5 space-y-4">
        <div class="flex items-center gap-2">
          <BarChart3 class="w-4 h-4 text-neutral-500" />
          <h3 class="text-sm font-semibold text-white">Earnings (last 30 days)</h3>
        </div>
        <div class="flex items-end gap-px h-24">
          <div
            v-for="bar in chartBars"
            :key="bar.day"
            class="flex-1 bg-white rounded-t-sm cursor-default relative group"
            :style="{ height: bar.height + '%' }"
            @mouseenter="hoveredBar = { day: bar.day, total: bar.total }"
            @mouseleave="hoveredBar = null"
          >
            <div
              v-if="hoveredBar?.day === bar.day"
              class="absolute -top-10 left-1/2 -translate-x-1/2 bg-neutral-800 border border-neutral-700 rounded px-2 py-1 text-[10px] font-mono text-white whitespace-nowrap z-10"
            >
              {{ formatPaise(bar.total) }}
              <span class="text-neutral-400 ml-1">{{ bar.day }}</span>
            </div>
          </div>
        </div>
        <div v-if="chartBars.length > 1" class="flex justify-between text-[10px] font-mono text-neutral-500">
          <span>{{ chartBars[0].day }}</span>
          <span>{{ chartBars[chartBars.length - 1].day }}</span>
        </div>
      </div>

      <!-- Campaign performance table -->
      <div v-if="analytics.earnings_by_campaign.length > 0" class="space-y-3">
        <h3 class="text-sm font-semibold text-white">Campaign performance</h3>
        <div class="rounded bg-neutral-950 border border-neutral-800 overflow-x-auto">
          <table class="w-full text-left">
            <thead>
              <tr class="border-b border-neutral-800">
                <th class="px-4 py-3 text-[10px] font-mono uppercase tracking-wider text-neutral-500 font-medium">Campaign</th>
                <th class="px-4 py-3 text-[10px] font-mono uppercase tracking-wider text-neutral-500 font-medium">Platform</th>
                <th class="px-4 py-3 text-[10px] font-mono uppercase tracking-wider text-neutral-500 font-medium text-right">CPM</th>
                <th class="px-4 py-3 text-[10px] font-mono uppercase tracking-wider text-neutral-500 font-medium text-right">Subs</th>
                <th class="px-4 py-3 text-[10px] font-mono uppercase tracking-wider text-neutral-500 font-medium text-right">Approved</th>
                <th class="px-4 py-3 text-[10px] font-mono uppercase tracking-wider text-neutral-500 font-medium text-right">Views</th>
                <th class="px-4 py-3 text-[10px] font-mono uppercase tracking-wider text-neutral-500 font-medium text-right">Earnings</th>
              </tr>
            </thead>
            <tbody>
              <NuxtLink
                v-for="c in analytics.earnings_by_campaign"
                :key="c.campaign_id"
                :to="`/app/campaigns/${c.campaign_id}`"
                custom
                v-slot="{ navigate }"
              >
                <tr
                  class="border-b border-neutral-900 last:border-0 hover:bg-neutral-900/50 cursor-pointer transition-colors"
                  @click="navigate"
                >
                  <td class="px-4 py-3 text-sm text-white font-medium truncate max-w-[200px]">{{ c.title }}</td>
                  <td class="px-4 py-3 text-xs font-mono text-neutral-400">{{ platformLabel(c.platform) }}</td>
                  <td class="px-4 py-3 text-sm font-mono text-neutral-300 text-right">{{ formatPaise(c.cpm_rate) }}</td>
                  <td class="px-4 py-3 text-sm font-mono text-neutral-300 text-right">{{ formatNumber(c.total_submissions) }}</td>
                  <td class="px-4 py-3 text-sm font-mono text-neutral-300 text-right">{{ formatNumber(c.approved_submissions) }}</td>
                  <td class="px-4 py-3 text-sm font-mono text-neutral-300 text-right">{{ formatViews(c.total_views) }}</td>
                  <td class="px-4 py-3 text-sm font-mono font-semibold text-white text-right">{{ formatPaise(c.total_earnings) }}</td>
                </tr>
              </NuxtLink>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Recent submissions -->
      <div v-if="analytics.recent_submissions.length > 0" class="space-y-3">
        <h3 class="text-sm font-semibold text-white">Recent submissions</h3>
        <div class="space-y-1">
          <a
            v-for="s in analytics.recent_submissions"
            :key="s.id"
            :href="s.post_url"
            target="_blank"
            rel="noopener noreferrer"
            class="flex items-center justify-between gap-3 rounded bg-neutral-950 border border-neutral-800 px-4 py-3 hover:border-neutral-700 transition-colors"
          >
            <div class="min-w-0 space-y-0.5">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium text-white truncate">{{ s.campaign_title }}</span>
                <SharedStatusBadge :status="s.status" />
              </div>
              <div class="text-[11px] font-mono text-neutral-500 truncate">{{ s.post_url }}</div>
            </div>
            <div class="flex items-center gap-3 shrink-0">
              <span class="text-xs font-mono text-neutral-400">{{ formatViews(s.latest_views) }} views</span>
              <ArrowRight class="w-3.5 h-3.5 text-neutral-600" />
            </div>
          </a>
        </div>
      </div>
    </template>

    <!-- Empty state -->
    <SharedEmptyState
      v-else
      :icon="Wallet"
      title="No analytics yet"
      description="Your analytics dashboard will populate once your clips get verified views and earn revenue."
    />
  </div>
</template>
