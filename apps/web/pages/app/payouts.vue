<script setup lang="ts">
import { CreditCard } from 'lucide-vue-next'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})
</script>

<template>
  <div class="space-y-5">
    <SharedPageHeader title="Payouts" description="Indian payout infrastructure & transfer history" />

    <SharedEmptyState
      :icon="CreditCard"
      title="No payouts yet"
      description="Once you have verified earnings, add a UPI or bank payout method here. Transfers are processed through Razorpay."
    />
  </div>
</template>