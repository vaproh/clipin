<script setup lang="ts">
import { ArrowRight, RefreshCw } from 'lucide-vue-next'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const { user } = useUser()
const { data: profile, isLoading, refetch } = useUserQuery()

const displayName = computed(() => profile.value?.display_name || user.value?.fullName || user.value?.firstName || 'Not provided')
const email = computed(() => profile.value?.email || user.value?.primaryEmailAddress?.emailAddress || '—')
const role = computed(() => profile.value?.role)
</script>

<template>
  <div class="space-y-5">
    <SharedPageHeader title="Settings" description="Profile & account preferences" />

    <div class="rounded bg-neutral-950 border border-neutral-800 space-y-4 p-5">
      <div class="flex items-center justify-between">
        <span class="text-xs font-mono uppercase tracking-wider text-neutral-400">Account Profile</span>
        <UiButton v-if="!isLoading" variant="ghost" size="xs" class="gap-1" @click="refetch()">
          <RefreshCw class="w-3 h-3" />
          Refresh
        </UiButton>
      </div>

      <div class="space-y-2.5 text-xs font-mono">
        <div class="flex items-center justify-between min-w-0 py-2 border-b border-neutral-900">
          <span class="text-neutral-400">Name</span>
          <span class="text-white text-right min-w-0 truncate ml-4">{{ displayName }}</span>
        </div>
        <div class="flex items-center justify-between min-w-0 py-2 border-b border-neutral-900">
          <span class="text-neutral-400">Email</span>
          <span class="text-white text-right min-w-0 truncate ml-4">{{ email }}</span>
        </div>
        <div class="flex items-center justify-between py-2 border-b border-neutral-900">
          <span class="text-neutral-400">Role</span>
          <SharedStatusBadge v-if="role" :status="role" />
          <span v-else class="text-neutral-500">{{ isLoading ? 'Loading…' : 'Not set' }}</span>
        </div>
        <div class="flex items-center justify-between min-w-0 py-2">
          <span class="text-neutral-400">Auth Identity</span>
          <span class="text-white text-right min-w-0 truncate ml-4">Managed via Clerk</span>
        </div>
      </div>

      <div class="pt-2 flex items-center justify-between rounded border border-neutral-800 bg-neutral-950 px-4 py-3">
        <div class="space-y-0.5">
          <div class="text-xs font-semibold text-white">Role & onboarding</div>
          <p class="text-[11px] font-mono text-neutral-500">Choose whether you clip for campaigns or run your own.</p>
        </div>
        <NuxtLink to="/app/onboarding">
          <UiButton variant="secondary" size="sm" class="gap-1 shrink-0">
            Change role
            <ArrowRight class="w-3.5 h-3.5" />
          </UiButton>
        </NuxtLink>
      </div>
    </div>
  </div>
</template>