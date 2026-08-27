<script setup lang="ts">
import { Youtube, Instagram, Music, Link, Unlink } from 'lucide-vue-next'
import type { SocialAccount } from '~/composables/useApi'

const { data: accounts, isLoading } = useSocialAccounts()
const connectAccount = useConnectSocialAccount()
const disconnectAccount = useDisconnectSocialAccount()

const platform = ref('youtube')
const platformUserId = ref('')
const platformUsername = ref('')

const platformOptions = [
  { value: 'youtube', label: 'YouTube', icon: Youtube },
  { value: 'instagram', label: 'Instagram', icon: Instagram },
  { value: 'tiktok', label: 'TikTok', icon: Music },
]

function handleConnect() {
  if (!platformUserId.value.trim()) return
  connectAccount.mutate(
    {
      platform: platform.value,
      platform_user_id: platformUserId.value.trim(),
      platform_username: platformUsername.value.trim() || platformUserId.value.trim(),
    },
    {
      onSuccess: () => {
        platformUserId.value = ''
        platformUsername.value = ''
      },
    },
  )
}

function handleDisconnect(id: string) {
  disconnectAccount.mutate(id)
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('en-IN', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function platformIcon(platform: string) {
  return platformOptions.find((p) => p.value === platform)?.icon ?? Music
}
</script>

<template>
  <div class="rounded bg-neutral-950 border border-neutral-800 space-y-4 p-5">
    <div class="flex items-center justify-between">
      <span class="text-xs font-mono uppercase tracking-wider text-neutral-400">Social Accounts</span>
    </div>

    <p class="text-[11px] text-neutral-500">
      Connect your social accounts to link submissions automatically. OAuth integration coming soon.
      For now, enter your platform user ID manually.
    </p>

    <!-- Loading -->
    <div v-if="isLoading" class="flex justify-center py-6">
      <SharedLoadingSpinner />
    </div>

    <!-- Connected accounts list -->
    <div v-else-if="accounts && accounts.length > 0" class="space-y-2">
      <div
        v-for="account in accounts"
        :key="account.id"
        class="flex items-center justify-between py-2.5 px-3 rounded border border-neutral-800 bg-neutral-900/50"
      >
        <div class="flex items-center gap-2.5">
          <component :is="platformIcon(account.platform)" class="w-4 h-4 text-neutral-400" />
          <div>
            <span class="text-xs font-medium text-white">
              {{ account.platform_username || account.platform_user_id }}
            </span>
            <span class="text-[10px] font-mono text-neutral-500 ml-2">
              {{ account.platform }}
            </span>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <span class="text-[10px] font-mono text-neutral-600">
            Connected {{ formatDate(account.created_at) }}
          </span>
          <UiButton
            variant="ghost"
            size="xs"
            class="text-neutral-500 hover:text-red-400 gap-1"
            :disabled="disconnectAccount.isPending.value"
            @click="handleDisconnect(account.id)"
          >
            <Unlink class="w-3 h-3" />
            Disconnect
          </UiButton>
        </div>
      </div>
    </div>

    <div v-else class="py-4 text-center">
      <Link class="w-4 h-4 text-neutral-600 mx-auto mb-1.5" />
      <span class="text-xs text-neutral-500">No connected accounts</span>
    </div>

    <!-- Connect form -->
    <div class="pt-3 border-t border-neutral-800 space-y-3">
      <span class="text-[10px] font-mono uppercase tracking-wider text-neutral-500">Add Account</span>

      <div class="space-y-2">
        <UiSelect v-model="platform">
          <UiSelectTrigger class="w-full h-8">
            <UiSelectValue />
          </UiSelectTrigger>
          <UiSelectContent>
            <UiSelectItem
              v-for="opt in platformOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </UiSelectItem>
          </UiSelectContent>
        </UiSelect>

        <UiInput
          v-model="platformUserId"
          placeholder="Platform User ID"
          class="h-8 text-xs"
        />

        <UiInput
          v-model="platformUsername"
          placeholder="Username (optional)"
          class="h-8 text-xs"
        />

        <UiButton
          variant="secondary"
          size="sm"
          class="w-full gap-1.5"
          :disabled="!platformUserId.trim() || connectAccount.isPending.value"
          @click="handleConnect"
        >
          <Link class="w-3 h-3" />
          {{ connectAccount.isPending.value ? 'Connecting…' : 'Connect' }}
        </UiButton>
      </div>
    </div>
  </div>
</template>
