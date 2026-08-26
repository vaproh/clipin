<script setup lang="ts">
import { ArrowLeft, Filter } from 'lucide-vue-next'
import type { Submission, SubmissionStatus } from '~/composables/useApi'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const route = useRoute()
const id = computed(() => route.params.id as string)

const { data: campaign, isLoading: campaignLoading } = useCampaign(id)
const { data: user } = useUserQuery()

const isOwner = computed(() => user.value && campaign.value && user.value.id === campaign.value.owner_id)

// Redirect non-owners
watchEffect(() => {
  if (campaign.value && !isOwner.value) {
    navigateTo(`/app/campaigns/${id.value}`)
  }
})

const { data: submissionsData, isLoading } = useCampaignSubmissions(id)
const submissions = computed<Submission[]>(() => submissionsData.value?.submissions ?? [])

const statusFilter = ref<SubmissionStatus | 'all'>('all')

const filteredSubmissions = computed(() => {
  if (statusFilter.value === 'all') return submissions.value
  return submissions.value.filter((s) => s.status === statusFilter.value)
})

const stats = computed(() => ({
  total: submissions.value.length,
  pending: submissions.value.filter((s) => s.status === 'pending').length,
  approved: submissions.value.filter((s) => s.status === 'approved' || s.status === 'auto_approved').length,
  rejected: submissions.value.filter((s) => s.status === 'rejected').length,
}))

// Reject dialog
const rejectTarget = ref<string | null>(null)
const rejectReason = ref('')
const { mutate: approveSubmission, isPending: approving } = useApproveSubmission(id)
const { mutate: rejectSubmission, isPending: rejecting } = useRejectSubmission(id)

function handleApprove(submissionId: string) {
  approveSubmission(submissionId)
}

function handleRejectRequest(submissionId: string) {
  rejectTarget.value = submissionId
  rejectReason.value = ''
}

function confirmReject() {
  if (!rejectTarget.value) return
  rejectSubmission(
    { submissionId: rejectTarget.value, reason: rejectReason.value || undefined },
    { onSuccess: () => { rejectTarget.value = null; rejectReason.value = '' } }
  )
}

function onDialogChange(open: boolean) {
  if (!open) rejectTarget.value = null
}
</script>

<template>
  <div class="space-y-5">
    <!-- Back link -->
    <NuxtLink
      :to="`/app/campaigns/${id}`"
      class="inline-flex items-center gap-1.5 text-xs font-mono text-neutral-500 hover:text-white transition-colors"
    >
      <ArrowLeft class="w-3.5 h-3.5" />
      Back to campaign
    </NuxtLink>

    <SharedPageHeader
      title="Submissions"
      :description="campaign ? `${campaign.title} - review clip submissions` : 'Review clip submissions'"
    />

    <!-- Loading -->
    <SharedLoadingSpinner v-if="isLoading || campaignLoading" />

    <!-- Not owner -->
    <div v-else-if="!isOwner" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
      <p class="text-sm text-neutral-400">You don't have access to this page.</p>
    </div>

    <template v-else>
      <!-- Stats -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <SharedStatCard label="Total" :value="stats.total" />
        <SharedStatCard label="Pending" :value="stats.pending" />
        <SharedStatCard label="Approved" :value="stats.approved" />
        <SharedStatCard label="Rejected" :value="stats.rejected" />
      </div>

      <!-- Filters -->
      <div class="flex items-center gap-2">
        <Filter class="w-3.5 h-3.5 text-neutral-500" />
        <div class="flex items-center gap-1">
          <button
            v-for="opt in (['all', 'pending', 'approved', 'rejected'] as const)"
            :key="opt"
            class="text-[11px] font-mono px-2.5 py-1 rounded transition-colors"
            :class="statusFilter === opt ? 'bg-white text-black' : 'text-neutral-500 hover:text-white'"
            @click="statusFilter = opt"
          >
            {{ opt === 'all' ? 'All' : opt }}
          </button>
        </div>
      </div>

      <!-- Empty -->
      <div v-if="filteredSubmissions.length === 0" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
        <p class="text-sm text-neutral-400">
          {{ statusFilter === 'all' ? 'No submissions yet.' : `No ${statusFilter} submissions.` }}
        </p>
      </div>

      <!-- Submission list -->
      <div v-else class="space-y-3">
        <SubmissionSubmissionCard
          v-for="sub in filteredSubmissions"
          :key="sub.id"
          :submission="sub"
          show-actions
          @approve="handleApprove"
          @reject="handleRejectRequest"
        />
      </div>

      <!-- Reject dialog -->
      <UiDialog :open="!!rejectTarget" @update:open="onDialogChange">
        <UiDialogContent class="sm:max-w-md">
          <UiDialogHeader>
            <UiDialogTitle>Reject Submission</UiDialogTitle>
            <UiDialogDescription>Optionally provide a reason for the clipper.</UiDialogDescription>
          </UiDialogHeader>
          <textarea
            v-model="rejectReason"
            rows="3"
            placeholder="Reason for rejection (optional)"
            class="w-full rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-sm text-white placeholder:text-neutral-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-700 resize-none"
          />
          <UiDialogFooter>
            <UiButton variant="ghost" size="sm" @click="rejectTarget = null">Cancel</UiButton>
            <UiButton size="sm" class="gap-1.5" :disabled="rejecting" @click="confirmReject">
              Reject
            </UiButton>
          </UiDialogFooter>
        </UiDialogContent>
      </UiDialog>
    </template>
  </div>
</template>
