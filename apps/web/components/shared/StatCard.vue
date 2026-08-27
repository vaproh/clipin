<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  label: string
  value: string | number
  icon?: Component
  hint?: string
}>()
</script>

<template>
  <div class="rounded bg-neutral-950 border border-neutral-800 p-4 space-y-3">
    <div class="flex items-center justify-between gap-2">
      <span class="text-[11px] font-mono uppercase tracking-wider text-neutral-500 truncate">{{ label }}</span>
      <component :is="icon" v-if="icon" class="w-4 h-4 text-neutral-500 shrink-0" />
    </div>
    <div class="text-xl font-semibold tracking-tight text-white truncate font-mono">{{ value }}</div>
    <p v-if="hint" class="text-[11px] font-mono text-neutral-500">{{ hint }}</p>
  </div>
</template>