<script setup lang="ts">
import { Trophy, ArrowRight } from 'lucide-vue-next'
import { formatPaise } from '~/lib/utils'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const sort = ref('earnings')
const { data: entries, isLoading, isError } = useLeaderboard(sort, 20)

const topThree = computed(() => (entries.value ?? []).filter((e) => e.rank <= 3))
const rest = computed(() => (entries.value ?? []).filter((e) => e.rank > 3))

const hasData = computed(() => (entries.value?.length ?? 0) > 0)
</script>

<template>
  <div class="space-y-5">
    <SharedPageHeader title="Leaderboard" description="Top performing clippers ranked by verified results" />

    <template v-if="isLoading">
      <SharedLoadingSpinner />
    </template>

    <div v-else-if="isError" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
      <p class="text-sm text-red-400 font-mono">Failed to load. Please try again.</p>
    </div>

    <template v-else-if="hasData">
      <!-- Sort toggle -->
      <div class="flex gap-1.5">
        <button
          class="px-3 py-1.5 rounded text-xs font-medium transition-colors"
          :class="sort === 'earnings'
            ? 'bg-white text-black'
            : 'bg-neutral-950 text-neutral-400 border border-neutral-800 hover:text-white'"
          @click="sort = 'earnings'"
        >
          By Earnings
        </button>
        <button
          class="px-3 py-1.5 rounded text-xs font-medium transition-colors"
          :class="sort === 'submissions'
            ? 'bg-white text-black'
            : 'bg-neutral-950 text-neutral-400 border border-neutral-800 hover:text-white'"
          @click="sort = 'submissions'"
        >
          By Submissions
        </button>
      </div>

      <!-- Top 3 podium -->
      <div v-if="topThree.length > 0" class="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <NuxtLink
          v-for="entry in topThree"
          :key="entry.user_id"
          :to="`/app/clippers/${entry.user_id}`"
          class="rounded border p-5 space-y-3 transition-colors"
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

          <div class="flex items-center gap-2.5">
            <div
              class="w-8 h-8 rounded bg-neutral-800 border border-neutral-700 flex items-center justify-center text-xs font-bold font-mono text-white shrink-0"
            >
              {{ (entry.display_name ?? 'U').charAt(0).toUpperCase() }}
            </div>
            <span class="text-sm font-medium text-white truncate">{{ entry.display_name ?? 'Unnamed clipper' }}</span>
          </div>

          <div class="grid grid-cols-2 gap-2 text-xs font-mono">
            <div>
              <div class="text-neutral-500">Earnings</div>
              <div class="text-white font-semibold">{{ formatPaise(entry.total_earnings) }}</div>
            </div>
            <div>
              <div class="text-neutral-500">Submissions</div>
              <div class="text-white font-semibold">{{ entry.total_submissions }}</div>
            </div>
          </div>
        </NuxtLink>
      </div>

      <!-- Rest of rankings table -->
      <div v-if="rest.length > 0" class="rounded border border-neutral-800 overflow-hidden">
        <table class="w-full text-xs">
          <thead>
            <tr class="border-b border-neutral-800 text-left">
              <th class="px-4 py-2.5 font-medium text-neutral-500 w-16">Rank</th>
              <th class="px-4 py-2.5 font-medium text-neutral-500">Clipper</th>
              <th class="px-4 py-2.5 font-medium text-neutral-500 text-right">Submissions</th>
              <th class="px-4 py-2.5 font-medium text-neutral-500 text-right">Earnings</th>
              <th class="px-4 py-2.5 font-medium text-neutral-500 text-right">Campaigns</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-neutral-900">
            <NuxtLink
              v-for="entry in rest"
              :key="entry.user_id"
              :to="`/app/clippers/${entry.user_id}`"
              class="contents"
              custom
              v-slot="{ navigate }"
            >
              <tr
                class="hover:bg-neutral-950 transition-colors cursor-pointer"
                @click="navigate"
              >
                <td class="px-4 py-2.5 font-mono font-semibold text-neutral-400">{{ entry.rank }}</td>
                <td class="px-4 py-2.5 text-white">{{ entry.display_name ?? 'Unnamed clipper' }}</td>
                <td class="px-4 py-2.5 font-mono text-right text-neutral-300">{{ entry.total_submissions }}</td>
                <td class="px-4 py-2.5 font-mono text-right text-white font-semibold">{{ formatPaise(entry.total_earnings) }}</td>
                <td class="px-4 py-2.5 font-mono text-right text-neutral-300">{{ entry.campaigns_participated }}</td>
              </tr>
            </NuxtLink>
          </tbody>
        </table>
      </div>
    </template>

    <SharedEmptyState
      v-else
      :icon="Trophy"
      title="No leaderboard data yet"
      description="Leaderboard rankings appear once clippers have verified submissions and earnings."
    />
  </div>
</template>
