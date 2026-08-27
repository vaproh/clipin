<script setup lang="ts">
import { Bell } from 'lucide-vue-next'
import { onClickOutside } from '@vueuse/core'

const panelRef = ref<HTMLElement | null>(null)
const isOpen = ref(false)

const { data: notifications, isLoading } = useNotifications()
const { data: unreadData } = useUnreadNotificationCount()
const markRead = useMarkNotificationRead()
const markAllRead = useMarkAllNotificationsRead()

const unreadCount = computed(() => unreadData.value?.count ?? 0)

onClickOutside(panelRef, () => {
  isOpen.value = false
})

function handleMarkRead(id: string) {
  markRead.mutate(id)
}

function handleMarkAllRead() {
  markAllRead.mutate()
}

function handleClickNotification(n: { id: string; link?: string | null }) {
  markRead.mutate(n.id)
  if (n.link) {
    navigateTo(n.link)
  }
  isOpen.value = false
}
</script>

<template>
  <div class="relative">
    <!-- Bell trigger -->
    <button
      class="relative p-1.5 rounded hover:bg-neutral-900 transition-colors"
      @click="isOpen = !isOpen"
    >
      <Bell class="w-4 h-4 text-neutral-400" />
      <span
        v-if="unreadCount > 0"
        class="absolute -top-0.5 -right-0.5 min-w-[14px] h-[14px] rounded-full bg-white text-black text-[9px] font-bold flex items-center justify-center px-0.5"
      >
        {{ unreadCount > 99 ? '99+' : unreadCount }}
      </span>
    </button>

    <!-- Panel -->
    <Teleport to="body">
      <div
        v-if="isOpen"
        class="fixed z-50"
        :style="{ top: '52px', right: '16px' }"
      >
        <div ref="panelRef">
          <AppNotificationPanel
            :notifications="notifications ?? []"
            :is-loading="isLoading"
            @mark-read="handleMarkRead"
            @mark-all-read="handleMarkAllRead"
            @click-notification="handleClickNotification"
          />
        </div>
      </div>
    </Teleport>
  </div>
</template>
