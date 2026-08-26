<script setup lang="ts">
import { Video } from 'lucide-vue-next'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})
</script>

<template>
  <div class="space-y-5">
    <SharedPageHeader title="Submissions" description="Track published clip post links & view verification" />

    <SharedEmptyState
      :icon="Video"
      title="No clip submissions"
      description="Publish a short-form clip for a campaign and submit its post link here. Verified views determine your earnings."
    />
  </div>
</template>