<script setup lang="ts">
import { ExternalLink } from 'lucide-vue-next'
import type { LedgerEntry } from '~/composables/useApi'
import { formatPaise } from '~/lib/utils'

defineProps<{
  entry: LedgerEntry
}>()

const typeLabels: Record<string, string> = {
  earning: 'Earning',
  platform_fee: 'Platform fee',
  refund: 'Refund',
  escrow_lock: 'Escrow lock',
  escrow_release: 'Escrow release',
}

const typeTones: Record<string, string> = {
  earning: 'bg-white text-black border-white',
  platform_fee: 'bg-neutral-800 text-neutral-300 border-neutral-600',
  refund: 'bg-neutral-700 text-white border-neutral-500',
  escrow_lock: 'bg-neutral-900 text-neutral-400 border-neutral-700',
  escrow_release: 'bg-neutral-800 text-neutral-300 border-neutral-600',
}

function isCredit(entry: LedgerEntry): boolean {
  return entry.entry_type === 'earning' || entry.entry_type === 'refund' || entry.entry_type === 'escrow_release'
}

function relativeDate(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const absDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24))
  if (absDays === 0) return 'today'
  if (absDays === 1) return 'yesterday'
  return `${absDays}d ago`
}
</script>

<template>
  <div class="flex items-center justify-between gap-3 py-2.5 border-b border-neutral-900 last:border-0">
    <div class="space-y-1 min-w-0">
      <div class="flex items-center gap-2">
        <span
          class="inline-flex items-center rounded border px-1.5 py-0.5 text-[9px] font-mono uppercase tracking-wider"
          :class="typeTones[entry.entry_type] ?? typeTones.earning"
        >
          {{ typeLabels[entry.entry_type] ?? entry.entry_type }}
        </span>
        <span class="text-[11px] font-mono text-neutral-500">{{ relativeDate(entry.created_at) }}</span>
      </div>
      <p v-if="entry.description" class="text-xs text-neutral-400 truncate">{{ entry.description }}</p>
      <div class="flex items-center gap-2 text-[10px] font-mono text-neutral-600">
        <NuxtLink
          v-if="entry.campaign_id"
          :to="`/app/campaigns/${entry.campaign_id}`"
          class="inline-flex items-center gap-1 hover:text-neutral-400 transition-colors"
        >
          <ExternalLink class="w-2.5 h-2.5" />
          Campaign
        </NuxtLink>
        <span v-if="entry.submission_id" class="text-neutral-700">|</span>
        <span v-if="entry.submission_id" class="truncate">sub:{{ entry.submission_id.slice(0, 8) }}</span>
      </div>
    </div>
    <span
      class="shrink-0 text-sm font-mono font-semibold tabular-nums"
      :class="isCredit(entry) ? 'text-white' : 'text-neutral-500'"
    >
      {{ isCredit(entry) ? '' : '−' }}{{ formatPaise(Math.abs(entry.amount)) }}
    </span>
  </div>
</template>
