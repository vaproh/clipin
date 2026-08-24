<script setup lang="ts">
const { user } = useUser()
</script>

<template>
  <header class="h-16 border-b border-[#23252a] bg-[#010102]/90 backdrop-blur px-6 flex items-center justify-between sticky top-0 z-40">
    <div class="flex items-center gap-2">
      <h1 class="text-sm font-semibold text-[#f7f8f8]">Dashboard</h1>
      <span class="text-xs text-[#62666d]">•</span>
      <span class="text-xs font-mono text-[#8a8f98]">Workspace</span>
    </div>

    <div class="flex items-center gap-3">
      <div v-if="user" class="text-right hidden sm:block">
        <div class="text-xs font-medium text-[#f7f8f8]">{{ user.fullName || user.firstName || 'Clipper' }}</div>
        <div class="text-[10px] font-mono text-[#8a8f98]">{{ user.primaryEmailAddress?.emailAddress }}</div>
      </div>
      <UserButton after-sign-out-url="/" />
    </div>
  </header>
</template>
