<script setup lang="ts">
import { Wallet } from 'lucide-vue-next'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})
</script>

<template>
  <div class="space-y-5">
    <SharedPageHeader title="Earnings" description="Verified view calculations & earned income" />

    <SharedEmptyState
      :icon="Wallet"
      title="No earnings yet"
      description="Your earnings ledger is empty. Verified views from published clips are converted to earnings here, recorded in an auditable ledger."
    />
  </div>
</template>