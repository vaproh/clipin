<script setup lang="ts">
import AppSidebar from '~/components/app/AppSidebar.vue'
import AppHeader from '~/components/app/AppHeader.vue'

useSeoMeta({
  robots: 'noindex, nofollow',
})

const mobileNavOpen = ref(false)
</script>

<template>
  <div class="min-h-screen bg-[#010102] text-[#f7f8f8] flex font-sans selection:bg-[#5e6ad2]/30">
    <!-- Desktop sidebar (always visible on md+) -->
    <AppSidebar class="hidden md:flex" />

    <!-- Mobile drawer overlay -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="mobileNavOpen"
          class="fixed inset-0 z-40 bg-black/60 md:hidden"
          @click="mobileNavOpen = false"
        />
      </Transition>
      <Transition name="slide">
        <div v-if="mobileNavOpen" class="fixed inset-y-0 left-0 z-50 md:hidden" tabindex="-1" @keydown.escape="mobileNavOpen = false">
          <AppSidebar class="flex" @navigate="mobileNavOpen = false" />
        </div>
      </Transition>
    </Teleport>

    <div class="flex-1 flex flex-col min-w-0">
      <AppHeader :mobile-nav-open="mobileNavOpen" @toggle-mobile-nav="mobileNavOpen = !mobileNavOpen" />
      <main class="flex-1 p-4 md:p-6 max-w-6xl w-full mx-auto">
        <slot />
      </main>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
.slide-enter-active, .slide-leave-active { transition: transform 0.2s ease; }
.slide-enter-from, .slide-leave-to { transform: translateX(-100%); }
</style>
