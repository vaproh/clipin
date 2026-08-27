<script setup lang="ts">
import { Motion } from 'motion-v'
import { Trophy, ArrowRight } from 'lucide-vue-next'
import { formatPaise } from '~/lib/utils'

const sort = ref('earnings')
const { data: entries } = useLeaderboard(sort, 3)

const topThree = computed(() => entries.value ?? [])
</script>

<template>
  <section class="py-16 border-t border-neutral-900 bg-black text-white">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 space-y-10">
      <!-- Header -->
      <Motion
        :initial="{ opacity: 0, y: 12 }"
        :animate="{ opacity: 1, y: 0 }"
        :transition="{ duration: 0.4 }"
        class="text-center space-y-2"
      >
        <div class="text-xs font-mono uppercase tracking-wider text-neutral-400">Leaderboard</div>
        <h2 class="text-2xl sm:text-3xl font-semibold tracking-tight text-white">
          Top clippers this month
        </h2>
      </Motion>

      <!-- Cards -->
      <div v-if="topThree.length > 0" class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Motion
          v-for="(entry, idx) in topThree"
          :key="entry.user_id"
          :initial="{ opacity: 0, y: 16 }"
          :animate="{ opacity: 1, y: 0 }"
          :transition="{ duration: 0.4, delay: 0.08 * (idx + 1) }"
          class="rounded border p-5 space-y-3"
          :class="entry.rank === 1
            ? 'bg-neutral-900 border-white/20'
            : entry.rank === 2
              ? 'bg-neutral-950 border-neutral-700'
              : 'bg-neutral-950 border-neutral-800'"
        >
          <div class="flex items-center justify-between">
            <span
              class="text-3xl font-bold font-mono"
              :class="entry.rank === 1 ? 'text-white' : entry.rank === 2 ? 'text-neutral-300' : 'text-neutral-500'"
            >
              {{ entry.rank }}
            </span>
            <Trophy
              class="w-5 h-5"
              :class="entry.rank === 1 ? 'text-white' : entry.rank === 2 ? 'text-neutral-400' : 'text-neutral-600'"
            />
          </div>

          <div class="text-sm font-medium text-white">{{ entry.display_name ?? 'Unnamed clipper' }}</div>

          <div class="text-xs font-mono text-neutral-400">
            {{ formatPaise(entry.total_earnings) }} earned
          </div>
        </Motion>
      </div>

      <!-- Empty state fallback -->
      <div v-else class="text-center py-8 text-sm text-neutral-500">
        Leaderboard rankings will appear here once clippers start earning.
      </div>

      <!-- CTA -->
      <Motion
        :initial="{ opacity: 0 }"
        :animate="{ opacity: 1 }"
        :transition="{ duration: 0.4, delay: 0.3 }"
        class="text-center"
      >
        <NuxtLink
          to="/app/leaderboard"
          class="inline-flex items-center gap-2 h-10 px-5 rounded bg-white text-black text-xs font-medium hover:bg-neutral-200 transition-colors"
        >
          View full leaderboard
          <ArrowRight class="w-3.5 h-3.5" />
        </NuxtLink>
      </Motion>
    </div>
  </section>
</template>
