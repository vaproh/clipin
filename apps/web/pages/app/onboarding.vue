<script setup lang="ts">
import { Motion } from 'motion-v'
import { Scissors, Megaphone, ArrowRight, AlertTriangle } from 'lucide-vue-next'
import type { UserRole } from '~/composables/useApi'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const { user } = useUser()
const { mutate: setRole, isPending, isError, error, reset } = useSetRole()

const roles: { value: UserRole; title: string; description: string; icon: typeof Scissors }[] = [
  {
    value: 'clipper',
    title: 'I want to clip',
    description: 'Browse campaigns, publish short-form clips, and earn from verified views.',
    icon: Scissors,
  },
  {
    value: 'owner',
    title: 'I want to run campaigns',
    description: 'Launch performance clipping campaigns and pay clippers for verified reach.',
    icon: Megaphone,
  },
]

const selected = ref<UserRole | null>(null)

// Auto-select role from landing page intent
onMounted(() => {
  const intent = localStorage.getItem('clipin_role_intent') as UserRole | null
  localStorage.removeItem('clipin_role_intent')
  if (intent && ['clipper', 'owner'].includes(intent) && !isPending.value) {
    choose(intent)
  }
})

function choose(role: UserRole) {
  selected.value = role
  setRole(role, {
    onSuccess: () => {
      navigateTo(role === 'owner' ? '/app/campaigns' : '/app')
    },
    onError: () => {
      selected.value = null
    },
  })
}
</script>

<template>
  <div class="max-w-xl mx-auto pt-10 space-y-8">
    <div class="space-y-2 text-center">
      <div class="text-[11px] font-mono uppercase tracking-wider text-neutral-500">Getting started</div>
      <h1 class="text-2xl font-semibold tracking-tight text-white">
        How do you want to use ClipIN?
      </h1>
      <p class="text-xs text-neutral-400 font-mono">
        Signed in as <span class="text-white">{{ user?.primaryEmailAddress?.emailAddress }}</span>. You can change this later in settings.
      </p>
    </div>

    <div v-if="isError" class="rounded border border-neutral-800 bg-neutral-950 p-4 flex items-start gap-3">
      <AlertTriangle class="w-4 h-4 text-neutral-400 mt-0.5 shrink-0" />
      <div class="flex-1 space-y-1">
        <div class="text-xs font-medium text-white">Couldn't save your role</div>
        <div class="text-[11px] font-mono text-neutral-400">{{ error?.message || 'The role endpoint is unreachable. Try again.' }}</div>
      </div>
      <UiButton variant="outline" size="xs" @click="reset">Retry</UiButton>
    </div>

    <div class="grid gap-3">
      <Motion
        v-for="role in roles"
        :key="role.value"
        :while-hover="{ y: -2 }"
        :while-tap="{ scale: 0.99 }"
        :transition="{ duration: 0.15 }"
      >
        <button
          type="button"
          class="w-full text-left rounded border p-5 transition-colors duration-150 disabled:opacity-60"
          :class="[
            selected === role.value
              ? 'border-white bg-neutral-900'
              : 'border-neutral-800 bg-neutral-950 hover:border-neutral-600 hover:bg-neutral-900',
          ]"
          :disabled="isPending"
          @click="choose(role.value)"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="space-y-2">
              <div class="flex items-center gap-2.5">
                <component :is="role.icon" class="w-4 h-4 text-neutral-400" />
                <span class="text-sm font-semibold text-white">{{ role.title }}</span>
              </div>
              <p class="text-xs text-neutral-400 leading-relaxed max-w-xs">{{ role.description }}</p>
            </div>
            <div
              v-if="isPending && selected === role.value"
              class="w-4 h-4 rounded-full border-2 border-neutral-700 border-t-white animate-spin shrink-0"
            ></div>
            <ArrowRight v-else class="w-4 h-4 text-neutral-500 shrink-0 mt-0.5" />
          </div>
        </button>
      </Motion>
    </div>

    <p class="text-[11px] font-mono text-neutral-600 text-center">
      Clippers earn per 1,000 verified views. Owners fund campaigns with an escrowed budget.
    </p>
  </div>
</template>