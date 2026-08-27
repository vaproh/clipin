<script setup lang="ts">
import { ExternalLink, Clock, AlertCircle, CheckCircle, XCircle } from 'lucide-vue-next'
import type { Submission } from '~/composables/useApi'
import { platformLabel } from '~/lib/utils'

const props = defineProps<{
  submission: Submission
  showActions?: boolean
  showVerification?: boolean
  showClipper?: boolean
  selectable?: boolean
  selected?: boolean
}>()

const emit = defineEmits<{
  approve: [id: string]
  reject: [id: string]
  'update:selected': [value: boolean]
}>()

const submissionId = toRef(props.submission, 'id')
const isApproved = computed(() =>
  props.submission.status === 'approved' || props.submission.status === 'auto_approved'
)
const { data: verificationStatus } = useVerificationStatus(
  computed(() => isApproved.value && props.showVerification ? props.submission.id : '')
)

function relativeDate(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const absMinutes = Math.floor(Math.abs(diffMs) / (1000 * 60))
  if (absMinutes < 1) return 'just now'
  if (absMinutes < 60) return `${absMinutes}m ago`
  const absHours = Math.floor(absMinutes / 60)
  if (absHours < 24) return `${absHours}h ago`
  const absDays = Math.floor(absHours / 24)
  return `${absDays}d ago`
}

function truncateUrl(url: string, max = 50): string {
  return url.length > max ? url.slice(0, max) + '...' : url
}
</script>

<template>
  <div class="rounded bg-neutral-950 border border-neutral-800 p-4 space-y-3" :class="selected ? 'border-neutral-600' : ''">
    <div class="flex items-start gap-3">
      <!-- Checkbox -->
      <button
        v-if="selectable && submission.status === 'pending'"
        type="button"
        class="mt-0.5 shrink-0 w-4 h-4 rounded border flex items-center justify-center transition-colors"
        :class="selected ? 'bg-white border-white' : 'border-neutral-600 hover:border-neutral-400'"
        @click.stop="emit('update:selected', !selected)"
      >
        <svg v-if="selected" class="w-2.5 h-2.5 text-black" viewBox="0 0 12 12" fill="none">
          <path d="M2 6l3 3 5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>

      <div class="flex items-start justify-between gap-3 flex-1 min-w-0">
        <div class="space-y-1.5 min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <span class="text-[9px] font-mono px-1.5 py-0.5 rounded bg-neutral-900 text-neutral-400 border border-neutral-800 shrink-0">
              {{ platformLabel(submission.platform) }}
            </span>
            <SharedStatusBadge :status="submission.status" />
            <NuxtLink
              v-if="showClipper"
              :to="`/app/clippers/${submission.clipper_id}`"
              class="text-[9px] font-mono px-1.5 py-0.5 rounded bg-neutral-900 text-neutral-500 border border-neutral-800 hover:text-white transition-colors shrink-0"
              @click.stop
            >
              {{ submission.clipper_id.slice(0, 8) }}...
            </NuxtLink>
          </div>
          <SubmissionVerificationBadge
            v-if="showVerification && isApproved && verificationStatus"
            :status="verificationStatus"
          />
          <a
            :href="submission.post_url"
            target="_blank"
            rel="noopener noreferrer"
            class="flex items-center gap-1.5 text-xs font-mono text-neutral-300 hover:text-white transition-colors break-all"
          >
            {{ truncateUrl(submission.post_url) }}
            <ExternalLink class="w-3 h-3 shrink-0" />
          </a>
        </div>

        <span class="text-[10px] font-mono text-neutral-500 shrink-0 flex items-center gap-1">
          <Clock class="w-3 h-3" />
          {{ relativeDate(submission.created_at) }}
        </span>
      </div>
    </div>

    <!-- Rejection reason -->
    <div v-if="submission.status === 'rejected' && submission.rejection_reason" class="flex items-start gap-2 rounded-lg bg-neutral-900 border border-neutral-800 p-3">
      <AlertCircle class="w-3.5 h-3.5 text-neutral-400 mt-0.5 shrink-0" />
      <p class="text-xs text-neutral-400 leading-relaxed">{{ submission.rejection_reason }}</p>
    </div>

    <!-- Owner actions -->
    <div v-if="showActions && submission.status === 'pending'" class="flex items-center gap-2 pt-1">
      <UiButton size="sm" class="gap-1.5" @click="emit('approve', submission.id)">
        <CheckCircle class="w-3.5 h-3.5" />
        Approve
      </UiButton>
      <UiButton size="sm" variant="outline" class="gap-1.5 text-neutral-400" @click="emit('reject', submission.id)">
        <XCircle class="w-3.5 h-3.5" />
        Reject
      </UiButton>
    </div>
  </div>
</template>
