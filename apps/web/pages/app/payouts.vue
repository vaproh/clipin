<script setup lang="ts">
import { CreditCard, Check, Pencil, ArrowRight, Banknote } from 'lucide-vue-next'
import { formatPaise } from '~/lib/utils'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const { data: user } = useUserQuery()
const { data: earnings } = useMyEarnings()
const { data: payoutsRes, isLoading: payoutsLoading } = useMyPayouts()
const updateUPI = useUpdateUPI()
const requestPayout = useRequestPayout()

const upiInput = ref('')
const editingUPI = ref(false)

const currentUPI = computed(() => user.value?.upi_id ?? null)
const hasUPI = computed(() => !!currentUPI.value)

const payouts = computed(() => payoutsRes.value?.payout_requests ?? [])

// Available balance: total earnings minus pending/processing payouts
const pendingPayoutTotal = computed(() =>
  payouts.value
    .filter((p) => p.status === 'pending' || p.status === 'processing')
    .reduce((sum, p) => sum + p.amount, 0),
)

const totalEarnings = computed(() => earnings.value?.total_earnings ?? 0)
const availableBalance = computed(() => totalEarnings.value - pendingPayoutTotal.value)
const canWithdraw = computed(() => availableBalance.value >= 50_000) // min ₹500 in paise

const amountRupees = ref('')
const amountPaise = computed(() => Math.round(parseFloat(amountRupees.value || '0') * 100))
const amountValid = computed(() => amountPaise.value >= 50_000 && amountPaise.value <= availableBalance.value)

const successMsg = ref('')
const errorMsg = ref('')

function startEditUPI() {
  upiInput.value = currentUPI.value ?? ''
  editingUPI.value = true
}

function cancelEditUPI() {
  editingUPI.value = false
  upiInput.value = ''
}

async function saveUPI() {
  if (!upiInput.value.trim()) return
  errorMsg.value = ''
  successMsg.value = ''
  try {
    await updateUPI.mutateAsync(upiInput.value.trim())
    editingUPI.value = false
    successMsg.value = 'UPI ID saved.'
    setTimeout(() => { successMsg.value = '' }, 3000)
  } catch {
    errorMsg.value = 'Failed to save UPI ID.'
  }
}

async function submitPayout() {
  if (!amountValid.value || !hasUPI.value) return
  errorMsg.value = ''
  successMsg.value = ''
  try {
    await requestPayout.mutateAsync(amountPaise.value)
    amountRupees.value = ''
    successMsg.value = 'Payout requested.'
    setTimeout(() => { successMsg.value = '' }, 3000)
  } catch {
    errorMsg.value = 'Payout request failed. Please try again.'
  }
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })
}

function formatTime(iso: string) {
  return new Date(iso).toLocaleTimeString('en-IN', { hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="space-y-5">
    <SharedPageHeader title="Payouts" description="Request payouts and view transfer history" />

    <!-- Feedback -->
    <div v-if="successMsg" class="rounded border border-neutral-700 bg-neutral-900 px-4 py-3 text-sm text-white">
      {{ successMsg }}
    </div>
    <div v-if="errorMsg" class="rounded border border-neutral-700 bg-neutral-950 px-4 py-3 text-sm text-neutral-400">
      {{ errorMsg }}
    </div>

    <!-- UPI section -->
    <div class="rounded bg-neutral-950 border border-neutral-800 p-4 space-y-3">
      <div class="flex items-center justify-between gap-2">
        <span class="text-[11px] font-mono uppercase tracking-wider text-neutral-500">UPI ID</span>
        <button
          v-if="hasUPI && !editingUPI"
          class="text-[11px] font-mono text-neutral-500 hover:text-white transition-colors flex items-center gap-1"
          @click="startEditUPI"
        >
          <Pencil class="w-3 h-3" />
          edit
        </button>
      </div>

      <!-- Display current UPI -->
      <div v-if="hasUPI && !editingUPI" class="text-sm font-mono text-white">
        {{ currentUPI }}
      </div>

      <!-- Edit / set UPI -->
      <template v-else>
        <div class="flex items-center gap-2">
          <UiInput
            v-model="upiInput"
            placeholder="name@upi"
            class="flex-1"
          />
          <UiButton
            size="sm"
            :disabled="!upiInput.trim() || updateUPI.isPending.value"
            @click="saveUPI"
          >
            <Check v-if="!updateUPI.isPending.value" class="w-3.5 h-3.5 mr-1" />
            {{ updateUPI.isPending.value ? 'Saving...' : 'Save' }}
          </UiButton>
          <UiButton
            v-if="hasUPI"
            size="sm"
            variant="ghost"
            @click="cancelEditUPI"
          >
            Cancel
          </UiButton>
        </div>
        <p v-if="!hasUPI" class="text-[11px] font-mono text-neutral-500">
          Set a UPI ID to receive payouts. Must be a valid VPA (e.g. name@upi).
        </p>
      </template>
    </div>

    <!-- Payout request -->
    <div class="rounded bg-neutral-950 border border-neutral-800 p-4 space-y-4">
      <div class="flex items-center justify-between gap-2">
        <span class="text-[11px] font-mono uppercase tracking-wider text-neutral-500">Request payout</span>
        <Banknote class="w-4 h-4 text-neutral-500" />
      </div>

      <div class="flex items-baseline gap-2">
        <span class="text-xs font-mono text-neutral-500">Available</span>
        <span class="text-lg font-semibold font-mono text-white">{{ formatPaise(availableBalance) }}</span>
      </div>

      <div v-if="canWithdraw && hasUPI" class="space-y-3">
        <div class="flex items-center gap-2">
          <span class="text-sm text-neutral-400">₹</span>
          <UiInput
            v-model="amountRupees"
            type="number"
            min="500"
            step="100"
            placeholder="500"
            class="flex-1"
          />
        </div>
        <p class="text-[11px] font-mono text-neutral-500">
          Min ₹500. Max {{ formatPaise(availableBalance) }}.
        </p>
        <UiButton
          :disabled="!amountValid || requestPayout.isPending.value"
          @click="submitPayout"
        >
          {{ requestPayout.isPending.value ? 'Requesting...' : 'Request Payout' }}
        </UiButton>
      </div>

      <div v-else-if="!hasUPI" class="text-xs font-mono text-neutral-500">
        Add a UPI ID above before requesting a payout.
      </div>
      <div v-else class="text-xs font-mono text-neutral-500">
        Available balance must be at least ₹500 to request a payout.
      </div>
    </div>

    <!-- Payout history -->
    <div class="space-y-3">
      <h3 class="text-sm font-semibold text-white">Payout history</h3>

      <SharedLoadingSpinner v-if="payoutsLoading" />

      <template v-else-if="payouts.length > 0">
        <div class="rounded bg-neutral-950 border border-neutral-800 divide-y divide-neutral-800">
          <div
            v-for="payout in payouts"
            :key="payout.id"
            class="px-4 py-3 flex items-center justify-between gap-3"
          >
            <div class="space-y-0.5 min-w-0">
              <div class="flex items-center gap-2">
                <span class="text-sm font-mono font-semibold text-white">{{ formatPaise(payout.amount) }}</span>
                <SharedStatusBadge :status="payout.status" />
              </div>
              <div class="text-[11px] font-mono text-neutral-500">
                {{ payout.upi_id }} &middot; {{ formatDate(payout.created_at) }} {{ formatTime(payout.created_at) }}
              </div>
              <div v-if="payout.failure_reason" class="text-[11px] font-mono text-neutral-600">
                {{ payout.failure_reason }}
              </div>
            </div>
            <ArrowRight class="w-3.5 h-3.5 text-neutral-600 shrink-0" />
          </div>
        </div>
      </template>

      <SharedEmptyState
        v-else
        :icon="CreditCard"
        title="No payouts yet"
        description="Once you have verified earnings, request a payout to your UPI. Transfers are processed through Razorpay."
      />
    </div>
  </div>
</template>
