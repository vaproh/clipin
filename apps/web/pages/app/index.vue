<script setup lang="ts">
import { Megaphone, Scissors, Wallet, ArrowRight, Plus, BarChart3, CircleDollarSign } from 'lucide-vue-next'
import { formatPaise } from '~/lib/utils'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const { user } = useUser()
const { data: profile, isLoading } = useUserQuery()

const roleLabel = computed(() => {
  if (profile.value) return profile.value.role
  return null
})

const isOwner = computed(() => profile.value?.role === 'owner')
const isClipper = computed(() => profile.value?.role === 'clipper')

// Owner data
const { data: ownerStats, isLoading: statsLoading } = useOwnerStats()
const { data: myCampaigns, isLoading: campaignsLoading } = useMyCampaigns()

const activeCampaigns = computed(() =>
  myCampaigns.value?.campaigns.filter((c) => c.status === 'active' || c.status === 'funded') ?? []
)

const totalSpent = computed(() => {
  if (!ownerStats.value) return 0
  return ownerStats.value.total_budget - ownerStats.value.total_remaining
})

// Clipper data
const { data: earnings } = useMyEarnings()
</script>

<template>
  <div class="space-y-6">
    <!-- Welcome card -->
    <div class="rounded bg-neutral-950 border border-neutral-800 p-6 space-y-2 relative overflow-hidden">
      <div class="flex items-center gap-2">
        <span class="w-1.5 h-1.5 rounded-full bg-white"></span>
        <span class="text-xs font-mono uppercase tracking-wider text-neutral-400">Workspace</span>
      </div>

      <h1 class="text-xl sm:text-2xl font-semibold tracking-tight text-white">
        Welcome to ClipIN<template v-if="user?.firstName">, {{ user.firstName }}</template>.
      </h1>

      <p class="text-xs font-mono text-neutral-400">
        Logged in as <span class="text-white">{{ user?.primaryEmailAddress?.emailAddress }}</span>
      </p>
    </div>

    <!-- Role selection prompt -->
    <div v-if="!roleLabel && !isLoading" class="rounded bg-neutral-950 border border-neutral-800 p-5 flex items-center justify-between gap-4">
      <div class="space-y-0.5">
        <div class="text-sm font-semibold text-white">Choose your role</div>
        <p class="text-xs text-neutral-400">Pick how you want to use ClipIN: clip for campaigns or run your own.</p>
      </div>
      <NuxtLink to="/app/onboarding" class="shrink-0">
        <UiButton size="sm" class="gap-1.5">
          Get started
          <ArrowRight class="w-3.5 h-3.5" />
        </UiButton>
      </NuxtLink>
    </div>

    <!-- Owner Dashboard -->
    <template v-else-if="isOwner">
      <!-- Stats -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <SharedStatCard
          label="Total campaigns"
          :value="statsLoading ? '...' : (ownerStats?.total_campaigns ?? 0)"
          :icon="Megaphone"
        />
        <SharedStatCard
          label="Active campaigns"
          :value="statsLoading ? '...' : (ownerStats?.active_campaigns ?? 0)"
          :icon="BarChart3"
        />
        <SharedStatCard
          label="Total spent"
          :value="statsLoading ? '...' : formatPaise(totalSpent ?? 0)"
          :icon="CircleDollarSign"
        />
        <SharedStatCard
          label="Remaining"
          :value="statsLoading ? '...' : formatPaise(ownerStats?.total_remaining ?? 0)"
          :icon="Wallet"
          hint="across all campaigns"
        />
      </div>

      <!-- Create Campaign CTA -->
      <NuxtLink to="/app/campaigns/new" class="block">
        <div class="rounded border border-dashed border-neutral-700 p-4 flex items-center justify-center gap-2 text-sm text-neutral-400 hover:border-neutral-500 hover:text-white transition-colors cursor-pointer">
          <Plus class="w-4 h-4" />
          Create Campaign
        </div>
      </NuxtLink>

      <!-- My Campaigns -->
      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-semibold text-white">My Campaigns</h3>
          <NuxtLink to="/app/campaigns" class="text-xs font-mono text-neutral-500 hover:text-white transition-colors">
            Browse all
          </NuxtLink>
        </div>

        <SharedLoadingSpinner v-if="campaignsLoading" />

        <SharedEmptyState
          v-else-if="!myCampaigns?.campaigns.length"
          :icon="Megaphone"
          title="No campaigns yet"
          description="Create your first campaign to start getting clips."
        />

        <div v-else class="space-y-2">
          <NuxtLink
            v-for="campaign in activeCampaigns"
            :key="campaign.id"
            :to="`/app/campaigns/${campaign.id}`"
            class="block"
          >
            <div class="rounded bg-neutral-950 border border-neutral-800 p-4 flex items-center justify-between gap-3 hover:border-neutral-700 transition-colors">
              <div class="space-y-1 min-w-0">
                <div class="text-sm font-medium text-white truncate">{{ campaign.title }}</div>
                <div class="flex items-center gap-2 text-[11px] font-mono text-neutral-500">
                  <SharedStatusBadge :status="campaign.status" />
                    <span>{{ formatPaise(campaign.remaining_budget ?? 0) }} remaining</span>
                </div>
              </div>
              <div class="flex items-center gap-2 shrink-0">
                <NuxtLink
                  :to="`/app/campaigns/${campaign.id}#analytics`"
                  class="text-[11px] font-mono text-neutral-500 hover:text-white transition-colors"
                  @click.stop
                >
                  Analytics
                </NuxtLink>
                <ArrowRight class="w-3.5 h-3.5 text-neutral-500" />
              </div>
            </div>
          </NuxtLink>

          <!-- Show non-active campaigns if any -->
          <template v-for="campaign in myCampaigns?.campaigns.filter((c) => c.status !== 'active' && c.status !== 'funded')" :key="campaign.id">
            <NuxtLink :to="`/app/campaigns/${campaign.id}`" class="block">
              <div class="rounded bg-neutral-950 border border-neutral-800 p-4 flex items-center justify-between gap-3 hover:border-neutral-700 transition-colors opacity-60">
                <div class="space-y-1 min-w-0">
                  <div class="text-sm font-medium text-white truncate">{{ campaign.title }}</div>
                  <div class="flex items-center gap-2 text-[11px] font-mono text-neutral-500">
                    <SharedStatusBadge :status="campaign.status" />
                    <span>{{ formatPaise(campaign.total_budget ?? 0) }} total</span>
                  </div>
                </div>
                <ArrowRight class="w-3.5 h-3.5 text-neutral-500 shrink-0" />
              </div>
            </NuxtLink>
          </template>
        </div>
      </div>
    </template>

    <!-- Clipper Dashboard -->
    <template v-else-if="isClipper">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <SharedStatCard label="Role" value="Clipper" :icon="Scissors" />
        <SharedStatCard
          label="Earnings"
          :value="earnings ? formatPaise(earnings.total_earnings ?? 0) : '—'"
          :icon="Wallet"
          hint="verified view earnings"
        />
      </div>

      <NuxtLink to="/app/campaigns" class="block">
        <div class="rounded border border-dashed border-neutral-700 p-4 flex items-center justify-center gap-2 text-sm text-neutral-400 hover:border-neutral-500 hover:text-white transition-colors cursor-pointer">
          <Megaphone class="w-4 h-4" />
          Browse Campaigns
        </div>
      </NuxtLink>
    </template>

    <!-- Fallback (role set but not loaded yet or unknown) -->
    <div v-else-if="!isLoading" class="grid grid-cols-1 sm:grid-cols-3 gap-3">
      <SharedStatCard label="Role" :value="roleLabel ?? 'not set'" :icon="Scissors" />
      <SharedStatCard label="Active campaigns" value="0" :icon="Megaphone" />
      <SharedStatCard
        label="Earnings"
        :value="earnings ? formatPaise(earnings.total_earnings ?? 0) : '—'"
        :icon="Wallet"
        hint="verified view earnings"
      />
    </div>
  </div>
</template>
