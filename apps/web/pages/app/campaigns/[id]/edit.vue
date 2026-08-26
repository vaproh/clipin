<script setup lang="ts">
import { ArrowLeft, Save } from 'lucide-vue-next'
import type { CreateCampaignBody } from '~/composables/useApi'
import { formatPaise } from '~/lib/utils'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id as string)

const { data: campaign, isLoading, isError } = useCampaign(id)

// Redirect if campaign is not editable
watchEffect(() => {
  if (campaign.value && !['draft', 'paused'].includes(campaign.value.status)) {
    router.replace(`/app/campaigns/${id.value}`)
  }
})

const form = reactive<Partial<CreateCampaignBody>>({
  title: '',
  description: '',
  brief_url: '',
  platform: 'youtube',
  cpm_rate: 0,
  total_budget: 0,
  max_clips_per_clipper: 3,
  min_views_per_clip: 1000,
  auto_approve_hours: 48,
})

// Populate form when campaign loads
watchEffect(() => {
  if (campaign.value) {
    form.title = campaign.value.title
    form.description = campaign.value.description ?? ''
    form.brief_url = campaign.value.brief_url ?? ''
    form.platform = campaign.value.platform
    form.cpm_rate = campaign.value.cpm_rate
    form.total_budget = campaign.value.total_budget
    form.max_clips_per_clipper = campaign.value.max_clips_per_clipper
    form.min_views_per_clip = campaign.value.min_views_per_clip
    form.auto_approve_hours = campaign.value.auto_approve_hours
  }
})

const errors = reactive<Record<string, string>>({})

const platformFee = computed(() => Math.round((form.total_budget ?? 0) * 0.1))
const totalWithFee = computed(() => (form.total_budget ?? 0) + platformFee.value)

const platformOptions = [
  { value: 'youtube', label: 'YouTube' },
  { value: 'instagram', label: 'Instagram' },
  { value: 'tiktok', label: 'TikTok' },
  { value: 'multi', label: 'Multi-platform' },
] as const

function clearErrors() {
  Object.keys(errors).forEach((k) => delete errors[k])
}

function validate(): boolean {
  clearErrors()
  if (!form.title?.trim()) errors.title = 'Title is required'
  if (!form.cpm_rate || form.cpm_rate <= 0) errors.cpm_rate = 'CPM rate must be greater than 0'
  if (!form.total_budget || form.total_budget <= 0) errors.total_budget = 'Total budget must be greater than 0'
  return Object.keys(errors).length === 0
}

const { mutate: updateCampaign, isPending } = useUpdateCampaign(id)

function save() {
  if (!validate()) return
  updateCampaign(form, {
    onSuccess: () => {
      router.push(`/app/campaigns/${id.value}`)
    },
  })
}
</script>

<template>
  <div class="space-y-6 max-w-2xl">
    <!-- Back link -->
    <NuxtLink
      :to="`/app/campaigns/${id}`"
      class="inline-flex items-center gap-1.5 text-xs font-mono text-neutral-500 hover:text-white transition-colors"
    >
      <ArrowLeft class="w-3.5 h-3.5" />
      Back to campaign
    </NuxtLink>

    <SharedPageHeader title="Edit Campaign" description="Update your campaign settings" />

    <!-- Loading -->
    <SharedLoadingSpinner v-if="isLoading" />

    <!-- Error / not found -->
    <div v-else-if="isError || !campaign" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
      <p class="text-sm text-neutral-400">Campaign not found.</p>
    </div>

    <!-- Not editable -->
    <div v-else-if="!['draft', 'paused'].includes(campaign.status)" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
      <p class="text-sm text-neutral-400">This campaign cannot be edited in its current state.</p>
    </div>

    <!-- Edit form -->
    <template v-else>
      <div class="rounded bg-neutral-950 border border-neutral-800 p-5 space-y-5">
        <!-- Basics -->
        <div class="space-y-1.5">
          <UiLabel for="title" class="text-xs font-mono text-neutral-400">Title</UiLabel>
          <UiInput
            id="title"
            v-model="form.title"
            placeholder="e.g. Summer Sale Reels"
            class="h-9"
          />
          <p v-if="errors.title" class="text-xs text-red-400 font-mono">{{ errors.title }}</p>
        </div>

        <div class="space-y-1.5">
          <UiLabel for="description" class="text-xs font-mono text-neutral-400">Description</UiLabel>
          <textarea
            id="description"
            v-model="form.description"
            rows="3"
            placeholder="Optional description for clippers"
            class="w-full rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-sm text-white placeholder:text-neutral-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-700 resize-none"
          />
        </div>

        <div class="space-y-1.5">
          <UiLabel for="brief_url" class="text-xs font-mono text-neutral-400">Brief URL</UiLabel>
          <UiInput
            id="brief_url"
            v-model="form.brief_url"
            placeholder="https://docs.example.com/brief"
            class="h-9"
          />
        </div>

        <div class="space-y-1.5">
          <UiLabel class="text-xs font-mono text-neutral-400">Platform</UiLabel>
          <UiSelect v-model="form.platform">
            <UiSelectTrigger class="w-full h-9">
              <UiSelectValue />
            </UiSelectTrigger>
            <UiSelectContent>
              <UiSelectItem
                v-for="opt in platformOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.label }}
              </UiSelectItem>
            </UiSelectContent>
          </UiSelect>
        </div>

        <!-- Divider -->
        <div class="h-px bg-neutral-800" />

        <!-- Budget -->
        <div class="space-y-1.5">
          <UiLabel for="cpm_rate" class="text-xs font-mono text-neutral-400">CPM Rate (paise per 1,000 views)</UiLabel>
          <div class="relative">
            <span class="absolute left-2.5 top-1/2 -translate-y-1/2 text-sm text-neutral-500">₹</span>
            <UiInput
              id="cpm_rate"
              v-model="form.cpm_rate"
              type="number"
              :min="1"
              class="h-9 pl-7"
            />
          </div>
          <p v-if="errors.cpm_rate" class="text-xs text-red-400 font-mono">{{ errors.cpm_rate }}</p>
        </div>

        <div class="space-y-1.5">
          <UiLabel for="total_budget" class="text-xs font-mono text-neutral-400">Total Budget (paise)</UiLabel>
          <div class="relative">
            <span class="absolute left-2.5 top-1/2 -translate-y-1/2 text-sm text-neutral-500">₹</span>
            <UiInput
              id="total_budget"
              v-model="form.total_budget"
              type="number"
              :min="100"
              class="h-9 pl-7"
            />
          </div>
          <p v-if="errors.total_budget" class="text-xs text-red-400 font-mono">{{ errors.total_budget }}</p>
        </div>

        <!-- Budget breakdown -->
        <div v-if="(form.total_budget ?? 0) > 0" class="rounded-lg bg-neutral-900 border border-neutral-800 p-4 space-y-2">
          <div class="text-[11px] font-mono uppercase tracking-wider text-neutral-500">Budget Breakdown</div>
          <div class="space-y-1 text-xs font-mono">
            <div class="flex justify-between text-neutral-400">
              <span>Deposit</span>
              <span>{{ formatPaise(form.total_budget ?? 0) }}</span>
            </div>
            <div class="flex justify-between text-neutral-400">
              <span>Platform fee (10%)</span>
              <span>-{{ formatPaise(platformFee) }}</span>
            </div>
            <div class="h-px bg-neutral-800" />
            <div class="flex justify-between text-white font-medium">
              <span>Total charged</span>
              <span>{{ formatPaise(totalWithFee) }}</span>
            </div>
          </div>
        </div>

        <!-- Rules -->
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-1.5">
            <UiLabel for="min_views" class="text-xs font-mono text-neutral-400">Min views per clip</UiLabel>
            <UiInput
              id="min_views"
              v-model="form.min_views_per_clip"
              type="number"
              :min="0"
              class="h-9"
            />
          </div>
          <div class="space-y-1.5">
            <UiLabel for="max_clips" class="text-xs font-mono text-neutral-400">Max clips per clipper</UiLabel>
            <UiInput
              id="max_clips"
              v-model="form.max_clips_per_clipper"
              type="number"
              :min="1"
              class="h-9"
            />
          </div>
        </div>

        <div class="space-y-1.5">
          <UiLabel for="auto_approve" class="text-xs font-mono text-neutral-400">Auto-approve after (hours)</UiLabel>
          <UiInput
            id="auto_approve"
            v-model="form.auto_approve_hours"
            type="number"
            :min="1"
            class="h-9"
          />
        </div>
      </div>

      <!-- Save -->
      <div class="flex items-center justify-end gap-2 pt-2">
        <NuxtLink :to="`/app/campaigns/${id}`">
          <UiButton variant="ghost" size="sm">
            Cancel
          </UiButton>
        </NuxtLink>
        <UiButton
          size="sm"
          class="gap-1.5"
          :disabled="isPending"
          @click="save"
        >
          <Save class="w-3.5 h-3.5" />
          {{ isPending ? 'Saving...' : 'Save Changes' }}
        </UiButton>
      </div>
    </template>
  </div>
</template>
