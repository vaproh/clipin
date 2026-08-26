<script setup lang="ts">
import { Video, Clock, CheckCircle, XCircle } from 'lucide-vue-next'
import type { Submission } from '~/composables/useApi'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const { data, isLoading } = useMySubmissions()

const submissions = computed<Submission[]>(() => data.value?.submissions ?? [])

const sortedSubmissions = computed(() => {
  const order: Record<string, number> = { pending: 0, auto_approved: 1, approved: 2, rejected: 3, disputed: 4 }
  return [...submissions.value].sort((a, b) => (order[a.status] ?? 5) - (order[b.status] ?? 5))
})

const stats = computed(() => ({
  total: submissions.value.length,
  pending: submissions.value.filter((s) => s.status === 'pending').length,
  approved: submissions.value.filter((s) => s.status === 'approved' || s.status === 'auto_approved').length,
  rejected: submissions.value.filter((s) => s.status === 'rejected').length,
}))
</script>

<template>
  <div class="space-y-5">
    <SharedPageHeader title="My Submissions" description="Track your clip submissions and verification status" />

    <!-- Loading -->
    <SharedLoadingSpinner v-if="isLoading" />

    <!-- Empty state -->
    <SharedEmptyState
      v-else-if="submissions.length === 0"
      :icon="Video"
      title="No submissions yet"
      description="Submit a clip link from a campaign to start earning from verified views."
      action-label="Browse campaigns"
      action-to="/app/campaigns"
    />

    <template v-else>
      <!-- Stats -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <SharedStatCard label="Total" :value="stats.total" :icon="Video" />
        <SharedStatCard label="Pending" :value="stats.pending" :icon="Clock" />
        <SharedStatCard label="Approved" :value="stats.approved" :icon="CheckCircle" />
        <SharedStatCard label="Rejected" :value="stats.rejected" :icon="XCircle" />
      </div>

      <!-- Submission list -->
      <div class="space-y-3">
        <SubmissionSubmissionCard
          v-for="sub in sortedSubmissions"
          :key="sub.id"
          :submission="sub"
        />
      </div>
    </template>
  </div>
</template>
