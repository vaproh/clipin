<script setup lang="ts">
definePageMeta({
  layout: 'app',
  middleware: 'auth',
})
</script>

<template>
  <div class="space-y-6">
    <div class="space-y-1">
      <h2 class="text-xl font-semibold text-[#f7f8f8]">Earnings</h2>
      <p class="text-xs font-mono text-[#8a8f98]">Verified view calculations & earned income</p>
    </div>

    <div class="p-12 rounded-xl bg-[#0f1011] border border-[#23252a] text-center space-y-3">
      <div class="text-sm font-semibold text-[#f7f8f8]">Earnings ledger</div>
      <p class="text-xs text-[#8a8f98] max-w-sm mx-auto font-mono">
        Auditable financial records will appear here as views are verified.
      </p>
    </div>
  </div>
</template>
