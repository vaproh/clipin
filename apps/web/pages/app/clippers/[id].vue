<script setup lang="ts">
import { ArrowLeft, Eye, Scissors, CheckCircle, Clock, Youtube, Instagram, Users, IndianRupee } from 'lucide-vue-next'
import { formatPaise, platformLabel } from '~/lib/utils'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const route = useRoute()
const id = computed(() => route.params.id as string)

const { data: profile, isLoading, isError } = useClipperProfile(id)

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-IN', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function initials(name: string | null): string {
  if (!name) return '?'
  return name.split(/\s+/).map(w => w[0]).join('').slice(0, 2).toUpperCase()
}

const approvalRate = computed(() => {
  if (!profile.value) return null
  const { total_submissions, approved_submissions } = profile.value.stats
  if (total_submissions === 0) return null
  return Math.round((approved_submissions / total_submissions) * 100)
})

const platformIcons: Record<string, typeof Youtube> = {
  youtube: Youtube,
  instagram: Instagram,
}
</script>

<template>
  <div class="space-y-5">
    <!-- Back link -->
    <NuxtLink to="/app/campaigns" class="inline-flex items-center gap-1.5 text-xs font-mono text-neutral-500 hover:text-white transition-colors">
      <ArrowLeft class="w-3.5 h-3.5" />
      Back
    </NuxtLink>

    <!-- Loading -->
    <SharedLoadingSpinner v-if="isLoading" />

    <!-- Error -->
    <div v-else-if="isError || !profile" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
      <p class="text-sm text-neutral-400">Clipper profile not found.</p>
    </div>

    <!-- Profile -->
    <template v-else>
      <!-- Header with avatar -->
      <div class="flex items-start gap-4">
        <div v-if="profile.avatar_url" class="w-12 h-12 rounded-full overflow-hidden border border-neutral-800 shrink-0">
          <img :src="profile.avatar_url" :alt="profile.display_name || 'Clipper'" class="w-full h-full object-cover" />
        </div>
        <div v-else class="w-12 h-12 rounded-full bg-neutral-900 border border-neutral-800 flex items-center justify-center shrink-0">
          <span class="text-sm font-mono font-medium text-neutral-400">{{ initials(profile.display_name) }}</span>
        </div>
        <div class="space-y-1 min-w-0">
          <SharedPageHeader
            :title="profile.display_name || 'Clipper'"
            :description="`Joined ${formatDate(profile.created_at)}`"
          />
          <p v-if="profile.bio" class="text-sm text-neutral-400 leading-relaxed max-w-xl">{{ profile.bio }}</p>
        </div>
      </div>

      <!-- Stats grid -->
      <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
        <SharedStatCard
          label="Total Submissions"
          :value="profile.stats.total_submissions"
          :icon="Scissors"
        />
        <SharedStatCard
          label="Approved"
          :value="profile.stats.approved_submissions"
          :icon="CheckCircle"
          :hint="approvalRate !== null ? `${approvalRate}% approval rate` : undefined"
        />
        <SharedStatCard
          label="Pending"
          :value="profile.stats.pending_submissions"
          :icon="Clock"
        />
        <SharedStatCard
          label="Total Views"
          :value="profile.stats.total_views.toLocaleString('en-IN')"
          :icon="Eye"
        />
        <SharedStatCard
          label="Total Earnings"
          :value="formatPaise(profile.stats.total_earnings)"
          :icon="IndianRupee"
        />
        <SharedStatCard
          label="Campaigns"
          :value="profile.stats.campaigns_participated"
          :icon="Users"
        />
      </div>

      <!-- Connected platforms -->
      <div class="space-y-3">
        <h3 class="text-sm font-semibold text-white">Connected Platforms</h3>

        <div v-if="profile.social_accounts.length > 0" class="flex flex-wrap gap-2">
          <div
            v-for="account in profile.social_accounts"
            :key="account.platform"
            class="flex items-center gap-2 rounded bg-neutral-950 border border-neutral-800 px-3 py-2"
          >
            <component :is="platformIcons[account.platform] ?? Users" class="w-4 h-4 text-neutral-400" />
            <span class="text-xs font-mono text-neutral-300">{{ platformLabel(account.platform) }}</span>
            <span v-if="account.platform_username" class="text-xs font-mono text-neutral-500">@{{ account.platform_username }}</span>
          </div>
        </div>

        <div v-else class="rounded bg-neutral-950 border border-neutral-800 p-6 text-center">
          <p class="text-xs text-neutral-500">No connected platforms.</p>
        </div>
      </div>
    </template>
  </div>
</template>
