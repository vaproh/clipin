<script setup lang="ts">
import { CheckCircle, XCircle, X } from 'lucide-vue-next'

const props = defineProps<{
  count: number
  loading?: boolean
}>()

const emit = defineEmits<{
  approve: []
  reject: [reason: string]
  clear: []
}>()

const showRejectInput = ref(false)
const rejectReason = ref('')

function handleReject() {
  if (!showRejectInput.value) {
    showRejectInput.value = true
    return
  }
  emit('reject', rejectReason.value)
  rejectReason.value = ''
  showRejectInput.value = false
}

function cancelReject() {
  showRejectInput.value = false
  rejectReason.value = ''
}
</script>

<template>
  <Transition name="slide-up">
    <div
      v-if="count > 0"
      class="fixed bottom-0 left-0 right-0 z-50 bg-black border-t border-neutral-800"
    >
      <div class="max-w-4xl mx-auto px-4 py-3 flex items-center gap-4">
        <span class="text-sm font-mono text-neutral-300 shrink-0">
          {{ count }} selected
        </span>

        <div class="flex items-center gap-2 ml-auto">
          <button
            type="button"
            class="text-xs font-mono text-neutral-500 hover:text-white transition-colors"
            @click="emit('clear')"
          >
            Clear
          </button>

          <UiButton
            size="sm"
            class="gap-1.5"
            :disabled="loading"
            @click="emit('approve')"
          >
            <CheckCircle class="w-3.5 h-3.5" />
            Approve all
          </UiButton>

          <UiButton
            size="sm"
            variant="outline"
            class="gap-1.5 text-neutral-400"
            :disabled="loading"
            @click="handleReject"
          >
            <XCircle class="w-3.5 h-3.5" />
            Reject all
          </UiButton>
        </div>
      </div>

      <!-- Reject reason input -->
      <Transition name="slide-up">
        <div v-if="showRejectInput" class="border-t border-neutral-800">
          <div class="max-w-4xl mx-auto px-4 py-3 flex items-center gap-3">
            <input
              v-model="rejectReason"
              type="text"
              placeholder="Reason for rejection (optional)"
              class="flex-1 rounded border border-neutral-700 bg-neutral-900 px-3 py-1.5 text-xs font-mono text-white placeholder:text-neutral-500 focus:outline-none focus:border-neutral-500"
            />
            <UiButton size="sm" variant="outline" :disabled="loading" @click="cancelReject">
              Cancel
            </UiButton>
            <UiButton size="sm" class="gap-1.5" :disabled="loading" @click="handleReject">
              <XCircle class="w-3.5 h-3.5" />
              Reject
            </UiButton>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

<style scoped>
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.2s ease, opacity 0.2s ease;
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}
</style>
