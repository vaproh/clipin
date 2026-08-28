<script setup lang="ts">
import { Wallet, Megaphone, ArrowRight } from 'lucide-vue-next'
import { formatPaise } from '~/lib/utils'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const { data: earnings, isLoading, isError } = useMyEarnings()
const { data: payoutsRes } = useMyPayouts()

const pendingPayoutTotal = computed(() =>
  (payoutsRes.value?.payout_requests ?? [])
    .filter((p) => p.status === 'pending' || p.status === 'processing')
    .reduce((sum, p) => sum + p.amount, 0),
)

const availableBalance = computed(() => (earnings.value?.total_earnings ?? 0) - pendingPayoutTotal.value)
const canWithdraw = computed(() => availableBalance.value >= 50_000)

const hasEarnings = computed(() => {
  if (!earnings.value) return false
  return earnings.value.total_earnings > 0 || earnings.value.recent_entries.length > 0
})

const campaignCount = computed(() => earnings.value?.campaign_earnings.length ?? 0)

const totalSubmissions = computed(() => {
  if (!earnings.value) return 0
  return earnings.value.recent_entries.filter((e) => e.entry_type === 'earning').length
})
</script>

<template>
  <div class="space-y-5">
    <SharedPageHeader title="Earnings" description="Verified view calculations and earned income" />

    <SharedLoadingSpinner v-if="isLoading" />

    <div v-else-if="isError" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
      <p class="text-sm text-red-400 font-mono">Failed to load. Please try again.</p>
    </div>

    <template v-else-if="hasEarnings && earnings">
      <!-- Stats row -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <SharedStatCard
          label="Total earnings"
          :value="formatPaise(earnings.total_earnings ?? 0)"
          :icon="Wallet"
        />
        <SharedStatCard
          label="Campaigns"
          :value="campaignCount"
          :icon="Megaphone"
        />
        <SharedStatCard
          label="Earning entries"
          :value="totalSubmissions"
          :icon="Wallet"
        />
      </div>

      <!-- Withdraw CTA -->
      <NuxtLink
        v-if="canWithdraw"
        to="/app/payouts"
        class="inline-flex items-center gap-2 rounded border border-neutral-700 bg-white px-4 py-2 text-sm font-medium text-black hover:bg-neutral-200 transition-colors"
      >
        Withdraw {{ formatPaise(availableBalance) }}
        <ArrowRight class="w-3.5 h-3.5" />
      </NuxtLink>

      <!-- Campaign breakdown -->
      <div v-if="earnings.campaign_earnings.length > 0" class="space-y-3">
        <h3 class="text-sm font-semibold text-white">Campaign breakdown</h3>
        <div class="space-y-2">
          <NuxtLink
            v-for="ce in earnings.campaign_earnings"
            :key="ce.campaign_id"
            :to="`/app/campaigns/${ce.campaign_id}`"
            class="block"
          >
            <div class="rounded bg-neutral-950 border border-neutral-800 p-4 flex items-center justify-between gap-3 hover:border-neutral-700 transition-colors">
              <div class="space-y-0.5 min-w-0">
                <div class="text-sm font-medium text-white truncate">{{ ce.campaign_title }}</div>
                <div class="text-[11px] font-mono text-neutral-500">campaign</div>
              </div>
              <div class="flex items-center gap-2 shrink-0">
                <span class="text-sm font-mono font-semibold text-white">{{ formatPaise(ce.amount) }}</span>
                <ArrowRight class="w-3.5 h-3.5 text-neutral-500" />
              </div>
            </div>
          </NuxtLink>
        </div>
      </div>

      <!-- Recent ledger entries -->
      <div v-if="earnings.recent_entries.length > 0" class="space-y-3">
        <h3 class="text-sm font-semibold text-white">Recent entries</h3>
        <div class="rounded bg-neutral-950 border border-neutral-800 px-4 divide-y-0">
          <LedgerEntryRow
            v-for="entry in earnings.recent_entries"
            :key="entry.id"
            :entry="entry"
          />
        </div>
      </div>
    </template>

    <!-- Empty state -->
    <SharedEmptyState
      v-else
      :icon="Wallet"
      title="No earnings yet"
      description="Your earnings ledger is empty. Verified views from published clips are converted to earnings here, recorded in an auditable ledger."
    />
  </div>
</template>
