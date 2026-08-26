<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <div class="space-y-1">
    <h2 class="text-lg font-semibold tracking-tight text-white">{{ title }}</h2>
    <p v-if="description" class="text-xs font-mono text-neutral-400">{{ description }}</p>
  </div>
</template>