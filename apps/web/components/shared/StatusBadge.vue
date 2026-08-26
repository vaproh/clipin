<script setup lang="ts">
const props = defineProps<{
  status: string
}>()

// Monochrome status tones: hierarchy via border/text brightness, never hue.
const tones: Record<string, { badge: string; dot: string }> = {
  active: { badge: 'bg-white text-black border-white', dot: 'bg-black' },
  owner: { badge: 'bg-white text-black border-white', dot: 'bg-black' },
  admin: { badge: 'bg-white text-black border-white', dot: 'bg-black' },
  approved: { badge: 'bg-neutral-800 text-white border-neutral-600', dot: 'bg-white' },
  clipper: { badge: 'bg-neutral-800 text-white border-neutral-600', dot: 'bg-white' },
  pending: { badge: 'bg-neutral-900 text-neutral-200 border-neutral-600', dot: 'bg-neutral-400' },
  paused: { badge: 'bg-neutral-900 text-neutral-300 border-neutral-700', dot: 'bg-neutral-500' },
  draft: { badge: 'bg-neutral-950 text-neutral-400 border-neutral-800', dot: 'bg-neutral-600' },
  rejected: { badge: 'bg-neutral-950 text-neutral-500 border-neutral-800', dot: 'bg-neutral-700' },
  auto_approved: { badge: 'bg-neutral-800 text-neutral-300 border-neutral-600', dot: 'bg-neutral-400' },
  disputed: { badge: 'bg-neutral-950 text-neutral-400 border-neutral-700', dot: 'bg-neutral-600' },
  ended: { badge: 'bg-neutral-950 text-neutral-500 border-neutral-800', dot: 'bg-neutral-700' },
  investigating: { badge: 'bg-neutral-900 text-neutral-300 border-neutral-700', dot: 'bg-neutral-500' },
  resolved: { badge: 'bg-white text-black border-white', dot: 'bg-black' },
  dismissed: { badge: 'bg-neutral-950 text-neutral-500 border-neutral-800', dot: 'bg-neutral-700' },
  processing: { badge: 'bg-neutral-900 text-neutral-300 border-neutral-700', dot: 'bg-neutral-500' },
  completed: { badge: 'bg-white text-black border-white', dot: 'bg-black' },
  default: { badge: 'bg-neutral-900 text-neutral-300 border-neutral-800', dot: 'bg-neutral-500' },
}

const tone = computed(() => tones[props.status] ?? tones.default)
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded border px-2 py-0.5 text-[10px] font-mono uppercase tracking-wider"
    :class="tone.badge"
  >
    <span class="w-1 h-1 rounded-full" :class="tone.dot"></span>
    {{ status }}
  </span>
</template>