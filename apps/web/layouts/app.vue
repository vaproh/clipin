<script setup lang="ts">
import AppSidebar from '~/components/app/AppSidebar.vue'
import AppHeader from '~/components/app/AppHeader.vue'
</script>

<template>
  <div class="min-h-screen bg-[#010102] text-[#f7f8f8] flex font-sans selection:bg-[#5e6ad2]/30">
    <AppSidebar class="hidden md:flex" />

    <div class="flex-1 flex flex-col min-w-0">
      <AppHeader />
      <main class="flex-1 p-6 max-w-6xl w-full mx-auto">
        <slot />
      </main>
    </div>
  </div>
</template>
