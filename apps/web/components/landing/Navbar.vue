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
    class="fixed top-0 left-0 right-0 z-50 transition-all duration-200"
    :class="[
      isScrolled
        ? 'bg-[#010102]/85 backdrop-blur-md border-b border-[#23252a] py-3'
        : 'bg-transparent border-b border-transparent py-5'
    ]"
  >
    <div class="max-w-6xl mx-auto px-4 sm:px-6 flex items-center justify-between">
      <!-- Brand Logo -->
      <NuxtLink to="/" class="flex items-center gap-2.5 group">
        <div class="w-7 h-7 rounded-md bg-[#5e6ad2] text-white font-semibold flex items-center justify-center text-xs tracking-tight shadow-sm shadow-[#5e6ad2]/20 group-hover:scale-105 transition-transform">
          CI
        </div>
        <span class="font-semibold text-base tracking-tight text-[#f7f8f8]">ClipIN</span>
        <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-[#18191a] text-[#8a8f98] border border-[#23252a]">INDIA</span>
      </NuxtLink>

      <!-- Navigation Links -->
      <nav class="hidden md:flex items-center gap-7 text-xs font-medium text-[#8a8f98]">
        <a href="#campaigns" class="hover:text-[#f7f8f8] transition-colors">Campaigns</a>
        <a href="#how-it-works" class="hover:text-[#f7f8f8] transition-colors">How it works</a>
        <a href="#for-clippers" class="hover:text-[#f7f8f8] transition-colors">For Clippers</a>
        <a href="#for-campaign-owners" class="hover:text-[#f7f8f8] transition-colors">For Campaign Owners</a>
      </nav>

      <!-- User Auth Controls -->
      <div class="flex items-center gap-3">
        <SignedOut>
          <NuxtLink
            to="/sign-in"
            class="text-xs font-medium text-[#8a8f98] hover:text-[#f7f8f8] px-3 py-1.5 rounded-md hover:bg-[#141516] transition-all"
          >
            Log in
          </NuxtLink>
          <NuxtLink
            to="/sign-up"
            class="text-xs font-medium bg-[#5e6ad2] hover:bg-[#828fff] text-white px-3.5 py-1.5 rounded-md transition-all shadow-sm shadow-[#5e6ad2]/30 flex items-center gap-1.5"
          >
            Get started
          </NuxtLink>
        </SignedOut>

        <SignedIn>
          <NuxtLink
            to="/app"
            class="text-xs font-medium bg-[#141516] hover:bg-[#18191a] text-[#f7f8f8] border border-[#23252a] hover:border-[#34343a] px-3.5 py-1.5 rounded-md transition-all flex items-center gap-1.5"
          >
            Open app
            <ArrowRight class="w-3.5 h-3.5 text-[#8a8f98]" />
          </NuxtLink>
          <UserButton after-sign-out-url="/" />
        </SignedIn>
      </div>
    </div>
  </header>
</template>
