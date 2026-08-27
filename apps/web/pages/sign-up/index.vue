<script setup lang="ts">
import { Check } from 'lucide-vue-next'

definePageMeta({
  layout: false,
})

useSeoMeta({
  title: 'Sign Up - ClipIN',
  description: 'Create your ClipIN account. Join India\'s performance clipping marketplace and start earning from verified views.',
  robots: 'noindex, nofollow',
})
</script>

<template>
  <div class="min-h-screen bg-[#0a0a0a] text-white flex selection:bg-white selection:text-black">
    <!-- Left Panel - Brand -->
    <div class="hidden md:flex md:w-1/2 bg-neutral-950 border-r border-neutral-900 flex-col justify-between p-10 lg:p-14 relative overflow-hidden">
      <!-- Subtle gradient orb -->
      <div class="absolute top-0 left-0 w-96 h-96 bg-neutral-800/20 rounded-full blur-3xl -translate-x-1/2 -translate-y-1/2 pointer-events-none" />

      <!-- Logo -->
      <NuxtLink to="/" class="inline-flex items-center gap-2.5 relative z-10">
        <div class="w-8 h-8 rounded bg-white text-black font-bold flex items-center justify-center text-sm tracking-tight">CI</div>
        <span class="font-semibold text-lg tracking-tight">ClipIN</span>
      </NuxtLink>

      <!-- Content -->
      <div class="relative z-10 space-y-8 max-w-md">
        <div class="space-y-3">
          <h1 class="text-2xl lg:text-3xl font-bold tracking-tight leading-tight">Start earning<br />from clips</h1>
          <p class="text-sm text-neutral-400 leading-relaxed">Create your account and join India's performance clipping marketplace.</p>
        </div>

        <ul class="space-y-3">
          <li v-for="item in ['Join 1000+ clippers earning monthly', 'Open marketplace - no approvals needed', 'Transparent CPM rates, instant tracking']" :key="item" class="flex items-start gap-2.5">
            <Check class="w-4 h-4 text-neutral-400 mt-0.5 shrink-0" />
            <span class="text-sm text-neutral-300">{{ item }}</span>
          </li>
        </ul>

        <div class="flex items-center gap-3 text-[10px] font-mono text-neutral-500 pt-2">
          <span>₹10,00,000+ paid to clippers</span>
          <span class="text-neutral-700">|</span>
          <span>500+ active campaigns</span>
        </div>
      </div>

      <!-- Footer -->
      <div class="text-[10px] text-neutral-600 relative z-10">
        © 2026 ClipIN. All rights reserved.
      </div>
    </div>

    <!-- Right Panel - Auth -->
    <div class="flex-1 flex flex-col items-center justify-center p-6 md:p-10">
      <!-- Mobile Logo -->
      <div class="mb-8 md:hidden flex items-center gap-2.5">
        <NuxtLink to="/" class="inline-flex items-center gap-2.5">
          <div class="w-7 h-7 rounded bg-white text-black font-bold flex items-center justify-center text-xs tracking-tight">CI</div>
          <span class="font-semibold text-base tracking-tight">ClipIN</span>
        </NuxtLink>
      </div>

      <!-- Auth Form -->
      <div class="w-full max-w-sm">
        <ClientOnly>
          <SignUp
            path="/sign-up"
            routing="path"
            sign-in-url="/sign-in"
            force-redirect-url="/app"
          />
          <template #fallback>
            <div class="w-full p-8 rounded-lg border border-neutral-800 text-center text-xs font-mono text-neutral-500 space-y-3">
              <div class="w-5 h-5 border-2 border-neutral-700 border-t-white rounded-full animate-spin mx-auto" />
              <div>Loading...</div>
            </div>
          </template>
        </ClientOnly>

        <p class="mt-6 text-center text-[10px] text-neutral-600">
          By continuing, you agree to our <a href="#" class="text-neutral-400 hover:text-white transition-colors">Terms of Service</a> and <a href="#" class="text-neutral-400 hover:text-white transition-colors">Privacy Policy</a>.
        </p>
      </div>

      <!-- Mobile features -->
      <div class="mt-10 md:hidden space-y-2 max-w-sm">
        <div v-for="item in ['Open marketplace, start earning', 'No approvals needed', 'Transparent CPM rates']" :key="item" class="flex items-center gap-2">
          <Check class="w-3.5 h-3.5 text-neutral-500 shrink-0" />
          <span class="text-xs text-neutral-500">{{ item }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
