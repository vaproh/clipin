<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  icon: Component
  title: string
  description: string
  actionLabel?: string
  actionTo?: string
}>()
</script>

<template>
  <div class="rounded bg-neutral-950 border border-neutral-800 py-12 px-6 text-center space-y-4">
    <div class="w-10 h-10 mx-auto rounded bg-neutral-900 border border-neutral-800 flex items-center justify-center">
      <component :is="icon" class="w-5 h-5 text-neutral-400" />
    </div>
    <div class="space-y-1.5">
      <h3 class="text-sm font-semibold text-white">{{ title }}</h3>
      <p class="text-xs text-neutral-400 max-w-sm mx-auto leading-relaxed">{{ description }}</p>
    </div>
    <NuxtLink v-if="actionLabel && actionTo" :to="actionTo" class="inline-flex pt-1">
      <UiButton variant="secondary" size="sm">{{ actionLabel }}</UiButton>
    </NuxtLink>
  </div>
</template>