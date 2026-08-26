<script setup lang="ts">
import { Eye, CheckCircle, Clock, BarChart3 } from 'lucide-vue-next'
import type { VerificationStatus } from '~/composables/useApi'

const props = defineProps<{
  status: VerificationStatus
}>()

function formatViews(n: number): string {
  return n.toLocaleString('en-IN')
}
</script>

<template>
  <!-- Verified: eligible views > 0 -->
  <span
    v-if="status.eligible_views > 0"
    class="inline-flex items-center gap-1 text-[10px] font-mono text-emerald-400"
  >
    <CheckCircle class="w-3 h-3" />
    Verified: {{ formatViews(status.eligible_views) }} eligible views
  </span>

  <!-- Pending: no snapshots yet -->
  <span
    v-else-if="!status.has_snapshots"
    class="inline-flex items-center gap-1 text-[10px] font-mono text-neutral-500"
  >
    <Clock class="w-3 h-3" />
    Pending verification
  </span>

  <!-- Tracking but below floor -->
  <span
    v-else
    class="inline-flex items-center gap-1 text-[10px] font-mono text-neutral-400"
  >
    <BarChart3 class="w-3 h-3" />
    {{ formatViews(status.current_views) }} views tracked
  </span>
</template>
