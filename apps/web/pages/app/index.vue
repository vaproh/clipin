<script setup lang="ts">
import { Clock, ShieldCheck } from 'lucide-vue-next'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const { user } = useUser()
</script>

<template>
  <div class="space-y-6">
    <!-- Welcome Greeting Card -->
    <div class="p-6 rounded bg-neutral-950 border border-neutral-800 space-y-2 relative overflow-hidden">
      <div class="flex items-center gap-2">
        <span class="w-1.5 h-1.5 rounded-full bg-white"></span>
        <span class="text-xs font-mono uppercase tracking-wider text-neutral-400">Authenticated Session</span>
      </div>

      <h1 class="text-xl sm:text-2xl font-semibold tracking-tight text-white">
        Welcome to ClipIN<template v-if="user?.firstName">, {{ user.firstName }}</template>.
      </h1>

      <p class="text-xs font-mono text-neutral-400">
        Logged in as <span class="text-white">{{ user?.primaryEmailAddress?.emailAddress }}</span>
      </p>
    </div>

    <!-- Marketplace Preparation Placeholder Card -->
    <div class="p-8 rounded bg-neutral-950 border border-neutral-800 text-center py-12 space-y-3">
      <div class="w-10 h-10 rounded bg-neutral-900 border border-neutral-800 flex items-center justify-center mx-auto text-white">
        <Clock class="w-5 h-5 text-neutral-300" />
      </div>

      <div class="space-y-1.5 max-w-md mx-auto">
        <h3 class="text-sm font-semibold text-white">The marketplace is being prepared.</h3>
        <p class="text-xs text-neutral-400 leading-relaxed font-normal">
          Campaign creation, clip submissions, performance metrics verification, and INR payouts will open in upcoming release phases.
        </p>
      </div>

      <div class="pt-2 inline-flex items-center gap-2 px-3 py-1 rounded bg-neutral-900 border border-neutral-800 text-xs font-mono text-neutral-300">
        <ShieldCheck class="w-3.5 h-3.5 text-white" />
        Account verified via Clerk
      </div>
    </div>
  </div>
</template>
