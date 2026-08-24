<script setup lang="ts">
definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const { user } = useUser()
</script>

<template>
  <div class="space-y-6">
    <div class="space-y-1">
      <h2 class="text-xl font-semibold text-[#f7f8f8]">Settings</h2>
      <p class="text-xs font-mono text-[#8a8f98]">Profile & account preferences</p>
    </div>

    <div class="p-6 rounded-xl bg-[#0f1011] border border-[#23252a] space-y-4">
      <div class="text-xs font-mono text-[#8a8f98] uppercase tracking-wider">Account Profile</div>

      <div class="space-y-3 text-xs font-mono">
        <div class="flex items-center justify-between py-2 border-b border-[#23252a]">
          <span class="text-[#8a8f98]">Name</span>
          <span class="text-[#f7f8f8]">{{ user?.fullName || user?.firstName || 'Not provided' }}</span>
        </div>
        <div class="flex items-center justify-between py-2 border-b border-[#23252a]">
          <span class="text-[#8a8f98]">Email</span>
          <span class="text-[#f7f8f8]">{{ user?.primaryEmailAddress?.emailAddress }}</span>
        </div>
        <div class="flex items-center justify-between py-2">
          <span class="text-[#8a8f98]">Auth Identity</span>
          <span class="text-[#5e6ad2]">Managed via Clerk</span>
        </div>
      </div>
    </div>
  </div>
</template>
