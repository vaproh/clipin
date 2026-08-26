<script setup lang="ts">
import { Megaphone, Scissors, Wallet, ArrowRight } from 'lucide-vue-next'

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
</script>

<template>
  <div class="space-y-6">
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

    <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
      <SharedStatCard label="Role" :value="isLoading ? '—' : (roleLabel ?? 'not set')" :icon="Scissors" />
      <SharedStatCard label="Active campaigns" value="0" :icon="Megaphone" />
      <SharedStatCard label="Earnings" value="—" :icon="Wallet" hint="Ledger opens when views are verified" />
    </div>

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
  </div>
</template>