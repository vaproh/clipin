<script setup lang="ts">
import { CheckCircle, XCircle, Megaphone, Banknote, Info } from 'lucide-vue-next'
import type { Notification } from '~/composables/useApi'

defineProps<{
  notifications: Notification[]
  isLoading: boolean
}>()

const emit = defineEmits<{
  markRead: [id: string]
  markAllRead: []
  clickNotification: [notification: Notification]
}>()

const typeIcons: Record<string, typeof CheckCircle> = {
  submission_approved: CheckCircle,
  submission_rejected: XCircle,
  campaign_update: Megaphone,
  payout_completed: Banknote,
  system: Info,
}

function relativeTime(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1) return 'Just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  const days = Math.floor(hrs / 24)
  if (days < 7) return `${days}d ago`
  return new Date(dateStr).toLocaleDateString('en-IN', { day: 'numeric', month: 'short' })
}
</script>

<template>
  <div class="w-80 max-h-96 flex flex-col rounded-lg border border-neutral-800 bg-neutral-950 shadow-xl">
    <!-- Header -->
    <div class="flex items-center justify-between px-3 py-2.5 border-b border-neutral-800">
      <span class="text-xs font-semibold text-white">Notifications</span>
      <button
        v-if="notifications.length > 0"
        class="text-[10px] font-mono text-neutral-400 hover:text-white transition-colors"
        @click="emit('markAllRead')"
      >
        Mark all read
      </button>
    </div>

    <!-- Loading -->
    <div v-if="isLoading" class="flex-1 flex items-center justify-center py-8">
      <SharedLoadingSpinner />
    </div>

    <!-- Empty -->
    <div v-else-if="notifications.length === 0" class="flex-1 flex flex-col items-center justify-center py-8 px-4">
      <Info class="w-5 h-5 text-neutral-600 mb-2" />
      <span class="text-xs text-neutral-500">No notifications yet</span>
    </div>

    <!-- List -->
    <div v-else class="flex-1 overflow-y-auto">
      <button
        v-for="n in notifications"
        :key="n.id"
        class="w-full text-left px-3 py-2.5 border-b border-neutral-900 hover:bg-neutral-900 transition-colors flex items-start gap-2.5"
        :class="{ 'bg-neutral-900/50': !n.is_read }"
        @click="emit('clickNotification', n)"
      >
        <component
          :is="typeIcons[n.type] || Info"
          class="w-3.5 h-3.5 mt-0.5 shrink-0"
          :class="n.is_read ? 'text-neutral-600' : 'text-neutral-300'"
        />
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <span
              class="text-xs font-medium truncate"
              :class="n.is_read ? 'text-neutral-400' : 'text-white'"
            >
              {{ n.title }}
            </span>
            <span v-if="!n.is_read" class="w-1.5 h-1.5 rounded-full bg-white shrink-0" />
          </div>
          <p v-if="n.body" class="text-[10px] text-neutral-500 mt-0.5 line-clamp-1">{{ n.body }}</p>
          <span class="text-[10px] font-mono text-neutral-600 mt-0.5 block">{{ relativeTime(n.created_at) }}</span>
        </div>
      </button>
    </div>
  </div>
</template>
