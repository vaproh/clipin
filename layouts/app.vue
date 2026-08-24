<script setup lang="ts">
import AppSidebar from '~/components/app/AppSidebar.vue'
import AppHeader from '~/components/app/AppHeader.vue'
</script>

<template>
  <div class="min-h-screen bg-black text-white flex font-sans selection:bg-white selection:text-black">
    <AppSidebar class="hidden md:flex" />

    <div class="flex-1 flex flex-col min-w-0">
      <AppHeader />
      <main class="flex-1 p-6 max-w-5xl w-full mx-auto">
        <slot />
      </main>
    </div>
  </div>
</template>
