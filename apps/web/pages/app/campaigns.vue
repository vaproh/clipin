<script setup lang="ts">
import { Megaphone } from 'lucide-vue-next'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})
</script>

<template>
  <div class="space-y-5">
    <SharedPageHeader title="Campaigns" description="Browse & launch performance clipping campaigns" />

    <SharedEmptyState
      :icon="Megaphone"
      title="No campaigns yet"
      description="Escrow-funded campaigns appear here once the marketplace opens. Owners fund a budget, clippers earn per 1,000 verified views."
    />
  </div>
</template>