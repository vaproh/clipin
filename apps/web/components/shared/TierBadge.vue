<script setup lang="ts">
const props = defineProps<{
  tier: string
}>()

// Monochrome tier tones: hierarchy via border/text brightness, never hue.
const tones: Record<string, { badge: string; dot: string }> = {
  bronze: { badge: 'bg-neutral-800 text-neutral-400 border-neutral-700', dot: 'bg-neutral-500' },
  silver: { badge: 'bg-neutral-700 text-neutral-200 border-neutral-600', dot: 'bg-neutral-300' },
  gold: { badge: 'bg-neutral-600 text-white border-neutral-500', dot: 'bg-white' },
  platinum: { badge: 'bg-white text-black border-white', dot: 'bg-black' },
  default: { badge: 'bg-neutral-800 text-neutral-400 border-neutral-700', dot: 'bg-neutral-500' },
}

const tone = computed(() => tones[props.tier] ?? tones.default)
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded border px-2 py-0.5 text-[10px] font-mono uppercase tracking-wider"
    :class="tone.badge"
  >
    <span class="w-1 h-1 rounded-full" :class="tone.dot"></span>
    {{ tier }}
  </span>
</template>
