<script setup lang="ts">
definePageMeta({
  layout: 'app',
  middleware: 'auth',
})
</script>

<template>
  <div class="space-y-6">
    <div class="space-y-1">
      <h2 class="text-xl font-semibold text-[#f7f8f8]">Campaigns</h2>
      <p class="text-xs font-mono text-[#8a8f98]">Browse & launch performance clipping campaigns</p>
    </div>

    <div class="p-12 rounded-xl bg-[#0f1011] border border-[#23252a] text-center space-y-3">
      <div class="text-sm font-semibold text-[#f7f8f8]">No active campaigns yet</div>
      <p class="text-xs text-[#8a8f98] max-w-sm mx-auto font-mono">
        Campaign marketplace and submission workflows are being onboarded.
      </p>
    </div>
  </div>
</template>
