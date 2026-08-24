<script setup lang="ts">
import { UserButton } from '@clerk/vue'

const { user } = useUser()
</script>

<template>
  <header class="h-14 border-b border-neutral-900 bg-black/90 backdrop-blur px-6 flex items-center justify-between sticky top-0 z-40">
    <div class="flex items-center gap-2">
      <h1 class="text-xs font-semibold text-white">Dashboard</h1>
      <span class="text-xs text-neutral-600">•</span>
      <span class="text-xs font-mono text-neutral-400">Workspace</span>
    </div>

    <div class="flex items-center gap-3">
      <div v-if="user" class="text-right hidden sm:block">
        <div class="text-xs font-medium text-white">{{ user.fullName || user.firstName || 'Clipper' }}</div>
        <div class="text-[10px] font-mono text-neutral-400">{{ user.primaryEmailAddress?.emailAddress }}</div>
      </div>
      <UserButton after-sign-out-url="/" />
    </div>
  </header>
</template>
