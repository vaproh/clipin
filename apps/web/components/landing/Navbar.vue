<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ArrowRight } from 'lucide-vue-next'

const isScrolled = ref(false)

const handleScroll = () => {
  isScrolled.value = window.scrollY > 20
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<template>
  <header
    class="fixed top-0 left-0 right-0 z-50 transition-colors duration-200"
    :class="[
      isScrolled
        ? 'bg-black/90 backdrop-blur-md border-b border-neutral-800'
        : 'bg-black/50 backdrop-blur-sm border-b border-neutral-900'
    ]"
  >
    <div class="max-w-6xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between">
      <!-- Brand Logo -->
      <NuxtLink to="/" class="inline-flex items-center gap-2.5 group">
        <div class="w-6 h-6 rounded bg-white text-black font-bold flex items-center justify-center text-[10px] tracking-tight">
          CI
        </div>
        <span class="font-semibold text-sm tracking-tight text-white">ClipIN</span>
        <span class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-neutral-900 text-neutral-400 border border-neutral-800">
          INDIA
        </span>
      </NuxtLink>

      <!-- Navigation Links -->
      <nav class="hidden md:flex items-center gap-6 text-xs font-medium text-neutral-400">
        <a href="#how-it-works" class="hover:text-white transition-colors">How it works</a>
        <a href="#for-clippers" class="hover:text-white transition-colors">For Clippers</a>
        <a href="#for-campaign-owners" class="hover:text-white transition-colors">For Campaign Owners</a>
        <a href="#why-clipin" class="hover:text-white transition-colors">Why ClipIN</a>
      </nav>

      <!-- User Auth Controls -->
      <div class="flex items-center gap-2.5">
        <SignedOut>
          <NuxtLink
            to="/sign-in"
            class="h-8 px-3 inline-flex items-center justify-center rounded text-xs font-medium text-neutral-300 hover:text-white hover:bg-neutral-900 transition-colors"
          >
            Log in
          </NuxtLink>
          <NuxtLink
            to="/sign-up"
            class="h-8 px-3.5 inline-flex items-center justify-center rounded bg-white text-black text-xs font-medium hover:bg-neutral-200 transition-colors"
          >
            Get started
          </NuxtLink>
        </SignedOut>

        <SignedIn>
          <NuxtLink
            to="/app"
            class="h-8 px-3.5 inline-flex items-center justify-center rounded bg-neutral-900 text-white border border-neutral-800 text-xs font-medium hover:bg-neutral-800 transition-colors gap-1.5"
          >
            Open app
            <ArrowRight class="w-3.5 h-3.5 text-neutral-400" />
          </NuxtLink>
          <UserButton after-sign-out-url="/" />
        </SignedIn>
      </div>
    </div>
  </header>
</template>
