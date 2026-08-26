<script setup lang="ts">
import { ArrowLeft, ExternalLink, Clock, Eye, Scissors, Timer, Send, XCircle, BarChart3 } from 'lucide-vue-next'
import { formatPaise } from '~/lib/utils'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const route = useRoute()
const id = computed(() => route.params.id as string)

const { data: campaign, isLoading, isError } = useCampaign(id)

const { data: user } = useUserQuery()

const isOwner = computed(() => user.value && campaign.value && user.value.id === campaign.value.owner_id)

const platformLabel: Record<string, string> = {
  youtube: 'YouTube',
  instagram: 'Instagram',
  tiktok: 'TikTok',
  multi: 'Multi',
}

const progress = computed(() => {
  if (!campaign.value) return 0
  const { total_budget, remaining_budget } = campaign.value
  if (total_budget === 0) return 0
  return ((total_budget - remaining_budget) / total_budget) * 100
})

// Owner mutations
const { mutate: pauseCampaign, isPending: pausing } = usePauseCampaign(id)
const { mutate: resumeCampaign, isPending: resuming } = useResumeCampaign(id)
const { mutate: cancelCampaign, isPending: cancelling } = useCancelCampaign(id)

const actionLoading = computed(() => pausing.value || resuming.value || cancelling.value)

// Submissions
const isClipper = computed(() => user.value?.role === 'clipper')
const isActive = computed(() => campaign.value?.status === 'active')
const canSubmit = computed(() => isClipper.value && isActive.value && !isOwner.value)

const { data: mySubmissionsData } = useMySubmissions()
const mySubmissionCount = computed(() => {
  if (!mySubmissionsData.value) return 0
  return mySubmissionsData.value.submissions.filter((s) => s.campaign_id === id.value).length
})

const { data: campaignSubmissionsData } = useCampaignSubmissions(id)
const campaignSubmissions = computed(() => campaignSubmissionsData.value?.submissions ?? [])
const pendingCount = computed(() => campaignSubmissions.value.filter((s) => s.status === 'pending').length)

// Verification aggregates for displayed submissions
const displayedSubmissionIds = computed(() => campaignSubmissions.value.slice(0, 5).map((s) => s.id))
const { data: verificationAggregate } = useAggregateVerification(displayedSubmissionIds)

// Reject dialog state
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

function formatDate(dateStr: string | null): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('en-IN', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function relativeDate(dateStr: string | null): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = date.getTime() - now.getTime()
  const absDays = Math.ceil(Math.abs(diffMs) / (1000 * 60 * 60 * 24))
  if (diffMs > 0) {
    if (absDays === 1) return 'tomorrow'
    return `in ${absDays} days`
  }
  if (absDays === 1) return 'yesterday'
  return `${absDays} days ago`
}
</script>

<template>
  <div class="space-y-5">
    <!-- Back link -->
    <NuxtLink to="/app/campaigns" class="inline-flex items-center gap-1.5 text-xs font-mono text-neutral-500 hover:text-white transition-colors">
      <ArrowLeft class="w-3.5 h-3.5" />
      Campaigns
    </NuxtLink>

    <!-- Loading -->
    <SharedLoadingSpinner v-if="isLoading" />

    <!-- Error -->
    <div v-else-if="isError || !campaign" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
      <p class="text-sm text-neutral-400">Campaign not found.</p>
    </div>

    <!-- Campaign detail -->
    <template v-else>
      <!-- Header -->
      <div class="space-y-2">
        <div class="flex items-start justify-between gap-3">
          <h2 class="text-lg font-semibold tracking-tight text-white">{{ campaign.title }}</h2>
          <div class="flex items-center gap-1.5 shrink-0">
            <span class="text-[9px] font-mono px-1.5 py-0.5 rounded bg-neutral-900 text-neutral-400 border border-neutral-800">
              {{ platformLabel[campaign.platform] ?? campaign.platform }}
            </span>
            <SharedStatusBadge :status="campaign.status" />
          </div>
        </div>

        <p v-if="campaign.description" class="text-sm text-neutral-400 leading-relaxed max-w-2xl">
          {{ campaign.description }}
        </p>

        <a
          v-if="campaign.brief_url"
          :href="campaign.brief_url"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center gap-1.5 text-xs font-mono text-neutral-400 hover:text-white transition-colors"
        >
          <ExternalLink class="w-3.5 h-3.5" />
          View brief
        </a>
      </div>

      <!-- Budget progress -->
      <div class="rounded bg-neutral-950 border border-neutral-800 p-4 space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-[11px] font-mono uppercase tracking-wider text-neutral-500">Budget</span>
          <span class="text-[11px] font-mono text-neutral-400">{{ Math.round(progress) }}% consumed</span>
        </div>
        <div class="h-2 w-full rounded-full bg-neutral-800 overflow-hidden">
          <div class="h-full rounded-full bg-white" :style="{ width: `${progress}%` }" />
        </div>
        <div class="flex items-center justify-between text-xs font-mono text-neutral-400">
          <span>{{ formatPaise(campaign.remaining_budget) }} remaining</span>
          <span>{{ formatPaise(campaign.total_budget) }} total</span>
        </div>
      </div>

      <!-- Stats grid -->
      <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
        <SharedStatCard label="CPM Rate" :value="formatPaise(campaign.cpm_rate)" :icon="Eye" hint="per 1,000 views" />
        <SharedStatCard label="Min Views" :value="campaign.min_views_per_clip.toLocaleString('en-IN')" :icon="Eye" hint="per clip" />
        <SharedStatCard label="Clips per Clipper" :value="campaign.max_clips_per_clipper" :icon="Scissors" />
        <SharedStatCard label="Auto-Approve" :value="`${campaign.auto_approve_hours}h`" :icon="Timer" hint="hours after submission" />
        <SharedStatCard
          v-if="campaign.max_clips_per_campaign"
          label="Max Clips"
          :value="campaign.max_clips_per_campaign"
          :icon="Scissors"
          hint="per campaign"
        />
        <SharedStatCard
          label="Platform Fee"
          :value="formatPaise(campaign.platform_fee)"
          hint="charged at deposit"
        />
      </div>

      <!-- Time info -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div class="rounded bg-neutral-950 border border-neutral-800 p-4 space-y-2">
          <div class="flex items-center gap-1.5 text-[11px] font-mono uppercase tracking-wider text-neutral-500">
            <Clock class="w-3.5 h-3.5" />
            Timeline
          </div>
          <div class="space-y-1.5 text-xs font-mono text-neutral-400">
            <div class="flex items-center justify-between">
              <span>Starts</span>
              <span>{{ formatDate(campaign.starts_at) }} <span v-if="campaign.starts_at" class="text-neutral-600">({{ relativeDate(campaign.starts_at) }})</span></span>
            </div>
            <div class="flex items-center justify-between">
              <span>Ends</span>
              <span>{{ formatDate(campaign.ends_at) }} <span v-if="campaign.ends_at" class="text-neutral-600">({{ relativeDate(campaign.ends_at) }})</span></span>
            </div>
            <div class="flex items-center justify-between">
              <span>Created</span>
              <span>{{ formatDate(campaign.created_at) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="flex items-center gap-2 pt-2">
        <UiButton v-if="canSubmit" size="sm" class="gap-1.5">
          <Send class="w-3.5 h-3.5" />
          Submit a clip
        </UiButton>
        <template v-if="isOwner">
          <NuxtLink v-if="campaign.status === 'draft' || campaign.status === 'paused'" :to="`/app/campaigns/${campaign.id}/edit`">
            <UiButton variant="outline" size="sm">
              Edit
            </UiButton>
          </NuxtLink>
          <UiButton
            v-if="campaign.status === 'active'"
            variant="ghost"
            size="sm"
            class="text-neutral-400"
            :disabled="actionLoading"
            @click="pauseCampaign()"
          >
            Pause
          </UiButton>
          <UiButton
            v-if="campaign.status === 'paused'"
            variant="ghost"
            size="sm"
            class="text-neutral-400"
            :disabled="actionLoading"
            @click="resumeCampaign()"
          >
            Resume
          </UiButton>
          <UiButton
            v-if="campaign.status === 'active' || campaign.status === 'paused' || campaign.status === 'funded'"
            variant="ghost"
            size="sm"
            class="text-red-400 hover:text-red-300"
            :disabled="actionLoading"
            @click="cancelCampaign()"
          >
            Cancel
          </UiButton>
        </template>
      </div>

      <!-- Clipper: submit form + existing count -->
      <template v-if="canSubmit">
        <SubmissionSubmitClipForm :campaign-id="id" :campaign="campaign" />
        <div v-if="mySubmissionCount > 0" class="text-xs font-mono text-neutral-500">
          You've already submitted {{ mySubmissionCount }} clip{{ mySubmissionCount === 1 ? '' : 's' }} to this campaign.
        </div>
      </template>

      <!-- Owner: submissions -->
      <template v-if="isOwner && campaignSubmissions.length > 0">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-semibold text-white">
            Submissions
            <span class="text-neutral-500 font-normal">({{ campaignSubmissions.length }}<span v-if="pendingCount > 0">, {{ pendingCount }} pending</span>)</span>
          </h3>
          <NuxtLink :to="`/app/campaigns/${id}/submissions`" class="text-xs font-mono text-neutral-500 hover:text-white transition-colors">
            View all
          </NuxtLink>
        </div>

        <!-- Aggregate verification stats -->
        <div
          v-if="verificationAggregate && verificationAggregate.tracked_count > 0"
          class="rounded bg-neutral-950 border border-neutral-800 p-4 flex items-center gap-4"
        >
          <div class="flex items-center gap-1.5 text-xs font-mono text-neutral-400">
            <BarChart3 class="w-3.5 h-3.5" />
            <span>{{ verificationAggregate.total_current_views.toLocaleString('en-IN') }} total views tracked</span>
          </div>
          <div v-if="verificationAggregate.verified_count > 0" class="flex items-center gap-1.5 text-xs font-mono text-emerald-400">
            <Eye class="w-3.5 h-3.5" />
            <span>{{ verificationAggregate.total_eligible_views.toLocaleString('en-IN') }} eligible views ({{ verificationAggregate.verified_count }} verified)</span>
          </div>
        </div>

        <div class="space-y-3">
          <SubmissionSubmissionCard
            v-for="sub in campaignSubmissions.slice(0, 5)"
            :key="sub.id"
            :submission="sub"
            show-actions
            show-verification
            @approve="handleApprove"
            @reject="handleRejectRequest"
          />
        </div>
      </template>

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
            <UiButton variant="ghost" size="sm" @click="rejectTarget.value = null">Cancel</UiButton>
            <UiButton size="sm" class="gap-1.5" :disabled="rejecting" @click="confirmReject">
              <XCircle class="w-3.5 h-3.5" />
              {{ rejecting ? 'Rejecting...' : 'Reject' }}
            </UiButton>
          </UiDialogFooter>
        </UiDialogContent>
      </UiDialog>
    </template>
  </div>
</template>
