<script setup lang="ts">
import { Sparkles, Clock, ShieldCheck, ArrowUpRight } from 'lucide-vue-next'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const { user } = useUser()
</script>

<template>
  <div class="space-y-8">
    <!-- Welcome Greeting Card -->
    <div class="p-8 rounded-xl bg-[#0f1011] border border-[#23252a] space-y-3 relative overflow-hidden">
      <div class="absolute -right-10 -bottom-10 w-64 h-64 bg-[#5e6ad2]/10 rounded-full blur-3xl pointer-events-none"></div>

      <div class="flex items-center gap-2">
        <span class="w-2 h-2 rounded-full bg-[#5e6ad2]"></span>
        <span class="text-xs font-mono uppercase tracking-wider text-[#8a8f98]">Authenticated Session</span>
      </div>

      <h1 class="text-2xl sm:text-3xl font-semibold tracking-tight text-[#f7f8f8]">
        Welcome to ClipIN<template v-if="user?.firstName">, {{ user.firstName }}</template>.
      </h1>

      <p class="text-xs font-mono text-[#8a8f98]">
        Logged in as <span class="text-[#d0d6e0]">{{ user?.primaryEmailAddress?.emailAddress }}</span>
      </p>
    </div>

    <!-- Marketplace Preparation Placeholder Card -->
    <div class="p-8 rounded-xl bg-[#0f1011] border border-[#23252a] text-center py-16 space-y-4">
      <div class="w-12 h-12 rounded-full bg-[#141516] border border-[#23252a] flex items-center justify-center mx-auto text-[#5e6ad2]">
        <Clock class="w-6 h-6 text-[#828fff]" />
      </div>

      <div class="space-y-2 max-w-md mx-auto">
        <h3 class="text-lg font-semibold text-[#f7f8f8]">The marketplace is being prepared.</h3>
        <p class="text-xs text-[#8a8f98] leading-relaxed">
          Campaign creation, clip submissions, performance metrics verification, and INR payouts will open in upcoming release phases.
        </p>
      </div>

      <div class="pt-4 inline-flex items-center gap-2 px-3 py-1.5 rounded-md bg-[#141516] border border-[#23252a] text-xs font-mono text-[#8a8f98]">
        <ShieldCheck class="w-3.5 h-3.5 text-emerald-400" />
        Account verified via Clerk
      </div>
    </div>
  </div>
</template>
