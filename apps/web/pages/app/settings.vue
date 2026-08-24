<script setup lang="ts">
definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const { user } = useUser()
</script>

<template>
  <div class="space-y-5">
    <div class="space-y-1">
      <h2 class="text-lg font-semibold text-white">Settings</h2>
      <p class="text-xs font-mono text-neutral-400">Profile & account preferences</p>
    </div>

    <div class="p-5 rounded bg-neutral-950 border border-neutral-800 space-y-4">
      <div class="text-xs font-mono text-neutral-400 uppercase tracking-wider">Account Profile</div>

      <div class="space-y-2.5 text-xs font-mono">
        <div class="flex items-center justify-between py-2 border-b border-neutral-900">
          <span class="text-neutral-400">Name</span>
          <span class="text-white">{{ user?.fullName || user?.firstName || 'Not provided' }}</span>
        </div>
        <div class="flex items-center justify-between py-2 border-b border-neutral-900">
          <span class="text-neutral-400">Email</span>
          <span class="text-white">{{ user?.primaryEmailAddress?.emailAddress }}</span>
        </div>
        <div class="flex items-center justify-between py-2">
          <span class="text-neutral-400">Auth Identity</span>
          <span class="text-white">Managed via Clerk</span>
        </div>
      </div>
    </div>
  </div>
</template>
