<script setup lang="ts">
import { LayoutDashboard, Megaphone, Video, Wallet } from 'lucide-vue-next'

const route = useRoute()

const items = [
  { label: 'Overview', path: '/app', icon: LayoutDashboard },
  { label: 'Campaigns', path: '/app/campaigns', icon: Megaphone },
  { label: 'Submissions', path: '/app/submissions', icon: Video },
  { label: 'Earnings', path: '/app/earnings', icon: Wallet },
]

function isActive(path: string) {
  return path === '/app' ? route.path === path : route.path.startsWith(path)
}
</script>

<template>
  <nav class="fixed inset-x-0 bottom-0 z-30 border-t border-neutral-800 bg-black/95 px-2 pb-[env(safe-area-inset-bottom)] backdrop-blur md:hidden" aria-label="Mobile navigation">
    <div class="mx-auto grid max-w-lg grid-cols-4">
      <NuxtLink
        v-for="item in items"
        :key="item.path"
        :to="item.path"
        class="flex min-h-16 flex-col items-center justify-center gap-1 text-[10px] font-medium transition-colors"
        :class="isActive(item.path) ? 'text-white' : 'text-neutral-500 hover:text-neutral-300'"
      >
        <component :is="item.icon" class="h-4 w-4" :class="isActive(item.path) ? 'text-white' : 'text-neutral-500'" />
        <span>{{ item.label }}</span>
      </NuxtLink>
    </div>
  </nav>
</template>
