<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ArrowRight, Menu, X } from 'lucide-vue-next'
import { UserButton } from '@clerk/vue'

const { isSignedIn } = useAuth()

const isScrolled = ref(false)
const mobileMenuOpen = ref(false)

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
        <template v-if="!isSignedIn">
          <NuxtLink
            to="/sign-in"
            class="h-8 px-3 hidden sm:inline-flex items-center justify-center rounded text-xs font-medium text-neutral-300 hover:text-white hover:bg-neutral-900 transition-colors"
          >
            Log in
          </NuxtLink>
          <NuxtLink
            to="/sign-up"
            class="h-8 px-3.5 hidden sm:inline-flex items-center justify-center rounded bg-white text-black text-xs font-medium hover:bg-neutral-200 transition-colors"
          >
            Get started
          </NuxtLink>
        </template>

        <template v-else>
          <NuxtLink
            to="/app"
            class="h-8 px-3.5 hidden sm:inline-flex items-center justify-center rounded bg-neutral-900 text-white border border-neutral-800 text-xs font-medium hover:bg-neutral-800 transition-colors gap-1.5"
          >
            Open app
            <ArrowRight class="w-3.5 h-3.5 text-neutral-400" />
          </NuxtLink>
          <UserButton after-sign-out-url="/" />
        </template>

        <!-- Mobile hamburger -->
        <button
          class="md:hidden p-1.5 -mr-1.5 rounded hover:bg-neutral-900 transition-colors"
          @click="mobileMenuOpen = !mobileMenuOpen"
        >
          <X v-if="mobileMenuOpen" class="w-5 h-5 text-neutral-300" />
          <Menu v-else class="w-5 h-5 text-neutral-300" />
        </button>
      </div>
    </div>
  </header>

  <!-- Mobile menu -->
  <Transition name="slide-down">
    <div v-if="mobileMenuOpen" class="fixed top-14 left-0 right-0 z-50 md:hidden border-b border-neutral-800 bg-black/95 backdrop-blur-md">
      <nav class="max-w-6xl mx-auto px-4 sm:px-6 py-4 space-y-1">
        <a
          v-for="link in [
            { href: '#how-it-works', label: 'How it works' },
            { href: '#for-clippers', label: 'For Clippers' },
            { href: '#for-campaign-owners', label: 'For Campaign Owners' },
            { href: '#why-clipin', label: 'Why ClipIN' },
          ]"
          :key="link.href"
          :href="link.href"
          class="block px-3 py-2.5 rounded text-sm text-neutral-300 hover:text-white hover:bg-neutral-900 transition-colors"
          @click="mobileMenuOpen = false"
        >
          {{ link.label }}
        </a>
        <div class="pt-2 border-t border-neutral-800 mt-2 space-y-1">
          <template v-if="!isSignedIn">
            <NuxtLink
              to="/sign-in"
              class="block px-3 py-2.5 rounded text-sm text-neutral-300 hover:text-white hover:bg-neutral-900 transition-colors"
              @click="mobileMenuOpen = false"
            >
              Log in
            </NuxtLink>
            <NuxtLink
              to="/sign-up"
              class="block px-3 py-2.5 rounded text-sm font-medium text-white bg-neutral-900 hover:bg-neutral-800 transition-colors text-center"
              @click="mobileMenuOpen = false"
            >
              Get started
            </NuxtLink>
          </template>
          <template v-else>
            <NuxtLink
              to="/app"
              class="block px-3 py-2.5 rounded text-sm font-medium text-white bg-neutral-900 hover:bg-neutral-800 transition-colors text-center"
              @click="mobileMenuOpen = false"
            >
              Open app
            </NuxtLink>
          </template>
        </div>
      </nav>
    </div>
  </Transition>
</template>

<style scoped>
.slide-down-enter-active, .slide-down-leave-active { transition: all 0.2s ease; }
.slide-down-enter-from, .slide-down-leave-to { opacity: 0; transform: translateY(-8px); }
</style>
