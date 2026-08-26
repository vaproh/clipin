<script setup lang="ts">
import { Menu, X } from 'lucide-vue-next'
import { UserButton } from '@clerk/vue'

defineProps<{ mobileNavOpen: boolean }>()
defineEmits<{ toggleMobileNav: [] }>()

const { user } = useUser()
const { data: profile } = useUserQuery()
</script>

<template>
  <header class="h-14 border-b border-neutral-900 bg-black/90 backdrop-blur px-4 md:px-6 flex items-center justify-between sticky top-0 z-30">
    <div class="flex items-center gap-2">
      <!-- Mobile hamburger -->
      <button
        class="md:hidden p-1.5 -ml-1.5 rounded hover:bg-neutral-900 transition-colors"
        @click="$emit('toggleMobileNav')"
      >
        <X v-if="mobileNavOpen" class="w-5 h-5 text-neutral-300" />
        <Menu v-else class="w-5 h-5 text-neutral-300" />
      </button>
      <h1 class="text-xs font-semibold text-white">Dashboard</h1>
      <span class="text-xs text-neutral-600">•</span>
      <span class="text-xs font-mono text-neutral-400">Workspace</span>
    </div>

    <div class="flex items-center gap-3">
      <SharedStatusBadge v-if="profile?.role" :status="profile.role" />
      <div v-if="user" class="text-right hidden sm:block">
        <div class="text-xs font-medium text-white">{{ user.fullName || user.firstName || 'Clipper' }}</div>
        <div class="text-[10px] font-mono text-neutral-400">{{ user.primaryEmailAddress?.emailAddress }}</div>
      </div>
      <UserButton after-sign-out-url="/" />
    </div>
  </header>
</template>
