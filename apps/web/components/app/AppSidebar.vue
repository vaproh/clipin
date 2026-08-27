<script setup lang="ts">
import type { Component } from 'vue'
import { useRoute } from '#app'
import { LayoutDashboard, Megaphone, Video, Wallet, Trophy, CreditCard, Settings, Shield } from 'lucide-vue-next'

const emit = defineEmits<{ navigate: [] }>()

const route = useRoute()
const { data: profile } = useUserQuery()

interface NavItem {
  name: string
  path: string
  icon: Component
  badge?: string
}

const navItems: NavItem[] = [
  { name: 'Overview', path: '/app', icon: LayoutDashboard },
  { name: 'Campaigns', path: '/app/campaigns', icon: Megaphone },
  { name: 'Submissions', path: '/app/submissions', icon: Video, badge: 'SOON' },
  { name: 'Earnings', path: '/app/earnings', icon: Wallet, badge: 'SOON' },
  { name: 'Leaderboard', path: '/app/leaderboard', icon: Trophy },
  { name: 'Payouts', path: '/app/payouts', icon: CreditCard, badge: 'SOON' },
  { name: 'Settings', path: '/app/settings', icon: Settings },
]

const adminItems: NavItem[] = [
  { name: 'Admin', path: '/app/admin', icon: Shield },
]

const isAdmin = computed(() => profile.value?.role === 'admin')
</script>

<template>
  <aside class="w-60 bg-black border-r border-neutral-900 flex flex-col justify-between p-3.5 h-screen sticky top-0 font-sans">
    <div class="space-y-5">
      <!-- App Brand Logo -->
      <NuxtLink to="/" class="flex items-center gap-2.5 px-2">
        <div class="w-6 h-6 rounded bg-white text-black font-bold flex items-center justify-center text-[10px] tracking-tight">
          CI
        </div>
        <span class="font-semibold text-sm tracking-tight text-white">ClipIN App</span>
      </NuxtLink>

      <!-- Nav Items -->
      <nav class="space-y-0.5 text-xs font-medium">
        <NuxtLink
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="flex items-center justify-between px-2.5 py-2 rounded transition-colors"
          :class="[
            route.path === item.path
              ? 'bg-neutral-900 text-white border border-neutral-800'
              : 'text-neutral-400 hover:text-white hover:bg-neutral-950'
          ]"
          @click="emit('navigate')"
        >
          <div class="flex items-center gap-2.5">
            <component :is="item.icon" class="w-4 h-4 text-neutral-400" :class="{ 'text-white': route.path === item.path }" />
            <span>{{ item.name }}</span>
          </div>

          <span v-if="item.badge" class="text-[9px] font-mono px-1.5 py-0.5 rounded bg-neutral-950 text-neutral-500 border border-neutral-900">
            {{ item.badge }}
          </span>
        </NuxtLink>
      </nav>

      <!-- Admin Nav -->
      <nav v-if="isAdmin" class="space-y-0.5 text-xs font-medium pt-3 border-t border-neutral-900">
        <NuxtLink
          v-for="item in adminItems"
          :key="item.path"
          :to="item.path"
          class="flex items-center justify-between px-2.5 py-2 rounded transition-colors"
          :class="[
            (route.path === item.path || (item.path !== '/app' && route.path.startsWith(item.path + '/')))
              ? 'bg-neutral-900 text-white border border-neutral-800'
              : 'text-neutral-400 hover:text-white hover:bg-neutral-950'
          ]"
          @click="emit('navigate')"
        >
          <div class="flex items-center gap-2.5">
            <component :is="item.icon" class="w-4 h-4 text-neutral-400" :class="{ 'text-white': route.path === item.path || (item.path !== '/app' && route.path.startsWith(item.path + '/')) }" />
            <span>{{ item.name }}</span>
          </div>
        </NuxtLink>
      </nav>
    </div>

    <!-- Status Footer Card -->
    <div class="p-3 rounded bg-neutral-950 border border-neutral-900 text-[11px] font-mono text-neutral-400 space-y-1">
      <div class="flex items-center gap-1.5 text-white">
        <span class="w-1.5 h-1.5 rounded-full bg-white animate-pulse"></span>
        Platform Early Access
      </div>
      <div class="text-neutral-500 text-[10px]">Marketplace Pre-Launch</div>
    </div>
  </aside>
</template>
