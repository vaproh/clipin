<script setup lang="ts">
definePageMeta({
  layout: 'app',
  middleware: 'auth',
})
</script>

<template>
  <div class="space-y-5">
    <div class="space-y-1">
      <h2 class="text-lg font-semibold text-white">Campaigns</h2>
      <p class="text-xs font-mono text-neutral-400">Browse & launch performance clipping campaigns</p>
    </div>

    <div class="p-8 rounded bg-neutral-950 border border-neutral-800 text-center space-y-2">
      <div class="text-xs font-semibold text-white">No active campaigns yet</div>
      <p class="text-xs text-neutral-400 max-w-sm mx-auto font-mono">
        Campaign marketplace and submission workflows are being onboarded.
      </p>
    </div>
  </div>
</template>
