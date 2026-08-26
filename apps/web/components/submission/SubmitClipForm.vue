<script setup lang="ts">
import { Send, CheckCircle, ExternalLink } from 'lucide-vue-next'
import type { Campaign } from '~/composables/useApi'

const props = defineProps<{
  campaignId: string
  campaign: Campaign
}>()

const postUrl = ref('')
const platform = ref(props.campaign.platform === 'multi' ? 'youtube' : props.campaign.platform)

const platformOptions = computed(() => {
  if (props.campaign.platform === 'multi') {
    return [
      { value: 'youtube', label: 'YouTube' },
      { value: 'instagram', label: 'Instagram' },
      { value: 'tiktok', label: 'TikTok' },
    ]
  }
  const labels: Record<string, string> = { youtube: 'YouTube', instagram: 'Instagram', tiktok: 'TikTok' }
  return [{ value: props.campaign.platform, label: labels[props.campaign.platform] ?? props.campaign.platform }]
})

const emit = defineEmits<{ submitted: [] }>()

const { mutate: submitClip, isPending, isSuccess, error, reset } = useSubmitClip(computed(() => props.campaignId))

function handleSubmit() {
  if (!postUrl.value.trim()) return
  submitClip(
    { post_url: postUrl.value.trim(), platform: platform.value },
    {
      onSuccess: () => {
        postUrl.value = ''
        emit('submitted')
      },
    }
  )
}

function dismissSuccess() {
  reset()
}
</script>

<template>
  <div class="rounded bg-neutral-950 border border-neutral-800 p-5 space-y-4">
    <div class="flex items-center gap-2">
      <Send class="w-4 h-4 text-neutral-400" />
      <h3 class="text-sm font-semibold text-white">Submit a Clip</h3>
    </div>

    <!-- Success state -->
    <div v-if="isSuccess" class="flex items-start gap-3 rounded-lg bg-neutral-900 border border-neutral-800 p-4">
      <CheckCircle class="w-4 h-4 text-white mt-0.5 shrink-0" />
      <div class="space-y-1.5 flex-1">
        <p class="text-sm text-white font-medium">Clip submitted</p>
        <p class="text-xs text-neutral-400">Your submission is pending review. The owner will approve it or it will be auto-approved after {{ campaign.auto_approve_hours }}h.</p>
      </div>
      <button class="text-xs font-mono text-neutral-500 hover:text-white transition-colors shrink-0" @click="dismissSuccess">
        Dismiss
      </button>
    </div>

    <!-- Form -->
    <form v-else class="space-y-3" @submit.prevent="handleSubmit">
      <div class="space-y-1.5">
        <UiLabel for="post-url" class="text-xs font-mono text-neutral-400">Post URL</UiLabel>
        <UiInput
          id="post-url"
          v-model="postUrl"
          placeholder="https://www.instagram.com/reel/..."
          class="h-9"
          required
        />
        <p class="text-[11px] font-mono text-neutral-500">Paste the URL of your published clip</p>
      </div>

      <div v-if="platformOptions.length > 1" class="space-y-1.5">
        <UiLabel class="text-xs font-mono text-neutral-400">Platform</UiLabel>
        <UiSelect v-model="platform">
          <UiSelectTrigger class="w-full h-9">
            <UiSelectValue />
          </UiSelectTrigger>
          <UiSelectContent>
            <UiSelectItem v-for="opt in platformOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </UiSelectItem>
          </UiSelectContent>
        </UiSelect>
      </div>

      <!-- Error -->
      <p v-if="error" class="text-xs text-red-400 font-mono">{{ error.message }}</p>

      <div class="flex items-center justify-end pt-1">
        <UiButton type="submit" size="sm" class="gap-1.5" :disabled="isPending || !postUrl.trim()">
          <Send class="w-3.5 h-3.5" />
          {{ isPending ? 'Submitting...' : 'Submit Clip' }}
        </UiButton>
      </div>
    </form>
  </div>
</template>
