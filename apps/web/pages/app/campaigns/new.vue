<script setup lang="ts">
import { ArrowLeft, ArrowRight, Check, Rocket } from 'lucide-vue-next'
import type { CreateCampaignBody } from '~/composables/useApi'
import { formatPaise } from '~/lib/utils'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const router = useRouter()

const step = ref(1)
const totalSteps = 3

const form = reactive<CreateCampaignBody>({
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

const errors = reactive<Record<string, string>>({})

const termsAccepted = ref(false)

const platformFee = computed(() => Math.round(form.total_budget * 0.1))
const totalWithFee = computed(() => form.total_budget + platformFee.value)
const netBudget = computed(() => form.total_budget - platformFee.value)

const platformOptions = [
  { value: 'youtube', label: 'YouTube' },
  { value: 'instagram', label: 'Instagram' },
  { value: 'tiktok', label: 'TikTok' },
  { value: 'multi', label: 'Multi-platform' },
] as const

function clearErrors() {
  Object.keys(errors).forEach((k) => delete errors[k])
}

function validateStep1(): boolean {
  clearErrors()
  if (!form.title.trim()) errors.title = 'Title is required'
  if (!form.platform) errors.platform = 'Platform is required'
  return Object.keys(errors).length === 0
}

function validateStep2(): boolean {
  clearErrors()
  if (!form.cpm_rate || form.cpm_rate <= 0) errors.cpm_rate = 'CPM rate must be greater than 0'
  if (!form.total_budget || form.total_budget <= 0) errors.total_budget = 'Total budget must be greater than 0'
  if (form.min_views_per_clip < 0) errors.min_views_per_clip = 'Cannot be negative'
  if (form.max_clips_per_clipper < 1) errors.max_clips_per_clipper = 'Must be at least 1'
  if (form.auto_approve_hours < 1) errors.auto_approve_hours = 'Must be at least 1'
  return Object.keys(errors).length === 0
}

function nextStep() {
  const valid = step.value === 1 ? validateStep1() : validateStep2()
  if (valid && step.value < totalSteps) step.value++
}

function prevStep() {
  clearErrors()
  if (step.value > 1) step.value--
}

const { mutate: createCampaign, isPending } = useCreateCampaign()

function launchCampaign() {
  createCampaign(form, {
    onSuccess: (campaign) => {
      router.push(`/app/campaigns/${campaign.id}`)
    },
  })
}

const stepLabels = ['Basics', 'Budget & Rules', 'Review']
</script>

<template>
  <div class="space-y-6 max-w-2xl">
    <!-- Back link -->
    <NuxtLink
      to="/app/campaigns"
      class="inline-flex items-center gap-1.5 text-xs font-mono text-neutral-500 hover:text-white transition-colors"
    >
      <ArrowLeft class="w-3.5 h-3.5" />
      Campaigns
    </NuxtLink>

    <SharedPageHeader title="Create Campaign" description="Set up a new clipping campaign" />

    <!-- Step indicator -->
    <div class="flex items-center gap-0">
      <template v-for="(label, i) in stepLabels" :key="i">
        <div class="flex items-center gap-2">
          <div
            :class="[
              'w-7 h-7 rounded-full flex items-center justify-center text-xs font-mono shrink-0 transition-colors',
              i + 1 < step && 'bg-white text-black',
              i + 1 === step && 'bg-white text-black',
              i + 1 > step && 'bg-neutral-800 text-neutral-500 border border-neutral-700',
            ]"
          >
            <Check v-if="i + 1 < step" class="w-3.5 h-3.5" />
            <span v-else>{{ i + 1 }}</span>
          </div>
          <span
            :class="[
              'text-xs font-mono hidden sm:inline',
              i + 1 <= step ? 'text-white' : 'text-neutral-500',
            ]"
          >
            {{ label }}
          </span>
        </div>
        <div
          v-if="i < stepLabels.length - 1"
          :class="[
            'flex-1 h-px mx-3',
            i + 1 < step ? 'bg-white' : 'bg-neutral-800',
          ]"
        />
      </template>
    </div>

    <!-- Step 1: Basics -->
    <div v-if="step === 1" class="rounded bg-neutral-950 border border-neutral-800 p-5 space-y-5">
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
          placeholder="Optional description for clipppers"
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
        <p class="text-[11px] font-mono text-neutral-500">Link to a detailed brief or reference document</p>
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
        <p v-if="errors.platform" class="text-xs text-red-400 font-mono">{{ errors.platform }}</p>
      </div>
    </div>

    <!-- Step 2: Budget & Rules -->
    <div v-if="step === 2" class="rounded bg-neutral-950 border border-neutral-800 p-5 space-y-5">
      <div class="space-y-1.5">
        <UiLabel for="cpm_rate" class="text-xs font-mono text-neutral-400">CPM Rate (paise per 1,000 views)</UiLabel>
        <div class="relative">
          <span class="absolute left-2.5 top-1/2 -translate-y-1/2 text-sm text-neutral-500">₹</span>
          <UiInput
            id="cpm_rate"
            v-model="form.cpm_rate"
            type="number"
            :min="1"
            placeholder="0"
            class="h-9 pl-7"
          />
        </div>
        <p class="text-[11px] font-mono text-neutral-500">Amount paid per 1,000 verified views (in paise)</p>
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
            placeholder="0"
            class="h-9 pl-7"
          />
        </div>
        <p v-if="errors.total_budget" class="text-xs text-red-400 font-mono">{{ errors.total_budget }}</p>
      </div>

      <!-- Budget breakdown -->
      <div v-if="form.total_budget > 0" class="rounded-lg bg-neutral-900 border border-neutral-800 p-4 space-y-2">
        <div class="text-[11px] font-mono uppercase tracking-wider text-neutral-500">Budget Breakdown</div>
        <div class="space-y-1 text-xs font-mono">
          <div class="flex justify-between text-neutral-400">
            <span>Deposit</span>
            <span>{{ formatPaise(form.total_budget) }}</span>
          </div>
          <div class="flex justify-between text-neutral-400">
            <span>Platform fee (10%)</span>
            <span>-{{ formatPaise(platformFee) }}</span>
          </div>
          <div class="h-px bg-neutral-800" />
          <div class="flex justify-between text-white font-medium">
            <span>Net budget</span>
            <span>{{ formatPaise(netBudget) }}</span>
          </div>
          <div class="flex justify-between text-neutral-400">
            <span>Total charged</span>
            <span>{{ formatPaise(totalWithFee) }}</span>
          </div>
        </div>
      </div>

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
        <p class="text-[11px] font-mono text-neutral-500">Submissions auto-approved if no action taken</p>
      </div>
    </div>

    <!-- Step 3: Review -->
    <div v-if="step === 3" class="rounded bg-neutral-950 border border-neutral-800 p-5 space-y-5">
      <div class="text-[11px] font-mono uppercase tracking-wider text-neutral-500">Campaign Summary</div>

      <div class="space-y-3 text-sm">
        <div class="flex justify-between">
          <span class="text-neutral-400">Title</span>
          <span class="text-white font-medium">{{ form.title }}</span>
        </div>
        <div v-if="form.description" class="flex justify-between">
          <span class="text-neutral-400">Description</span>
          <span class="text-white text-right max-w-[60%] truncate">{{ form.description }}</span>
        </div>
        <div v-if="form.brief_url" class="flex justify-between">
          <span class="text-neutral-400">Brief</span>
          <a :href="form.brief_url" target="_blank" class="text-white underline decoration-neutral-700 hover:decoration-neutral-400 transition-colors truncate max-w-[60%]">Link</a>
        </div>
        <div class="flex justify-between">
          <span class="text-neutral-400">Platform</span>
          <span class="text-white">{{ platformOptions.find(p => p.value === form.platform)?.label }}</span>
        </div>

        <div class="h-px bg-neutral-800" />

        <div class="flex justify-between">
          <span class="text-neutral-400">CPM Rate</span>
          <span class="text-white">{{ formatPaise(form.cpm_rate) }} / 1k views</span>
        </div>
        <div class="flex justify-between">
          <span class="text-neutral-400">Deposit</span>
          <span class="text-white">{{ formatPaise(form.total_budget) }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-neutral-400">Platform fee (10%)</span>
          <span class="text-white">{{ formatPaise(platformFee) }}</span>
        </div>
        <div class="flex justify-between font-medium">
          <span class="text-neutral-400">Total charged</span>
          <span class="text-white">{{ formatPaise(totalWithFee) }}</span>
        </div>

        <div class="h-px bg-neutral-800" />

        <div class="flex justify-between">
          <span class="text-neutral-400">Min views / clip</span>
          <span class="text-white">{{ form.min_views_per_clip.toLocaleString('en-IN') }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-neutral-400">Max clips / clipper</span>
          <span class="text-white">{{ form.max_clips_per_clipper }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-neutral-400">Auto-approve</span>
          <span class="text-white">{{ form.auto_approve_hours }}h</span>
        </div>
      </div>

      <div class="rounded-lg bg-neutral-900 border border-neutral-800 p-3">
        <label class="flex items-start gap-2.5 cursor-pointer">
          <input
            v-model="termsAccepted"
            type="checkbox"
            class="mt-0.5 h-4 w-4 rounded border-neutral-600 bg-transparent text-white focus:ring-neutral-700"
          >
          <span class="text-xs text-neutral-400 leading-relaxed">
            I understand that funds will be escrowed and a 10% platform fee applies at deposit. Unspent budget will be refunded on campaign end.
          </span>
        </label>
      </div>
    </div>

    <!-- Navigation -->
    <div class="flex items-center justify-between pt-2">
      <UiButton
        v-if="step > 1"
        variant="ghost"
        size="sm"
        class="gap-1.5"
        @click="prevStep"
      >
        <ArrowLeft class="w-3.5 h-3.5" />
        Back
      </UiButton>
      <div v-else />

      <UiButton
        v-if="step < totalSteps"
        size="sm"
        class="gap-1.5"
        @click="nextStep"
      >
        Next
        <ArrowRight class="w-3.5 h-3.5" />
      </UiButton>

      <UiButton
        v-else
        size="sm"
        class="gap-1.5"
        :disabled="!termsAccepted || isPending"
        @click="launchCampaign"
      >
        <Rocket class="w-3.5 h-3.5" />
        {{ isPending ? 'Launching...' : 'Launch Campaign' }}
      </UiButton>
    </div>
  </div>
</template>


