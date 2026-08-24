<script setup lang="ts">
import type { Component } from 'vue'
import { useRoute } from '#app'
import { LayoutDashboard, Megaphone, Video, Wallet, CreditCard, Settings } from 'lucide-vue-next'

const route = useRoute()

interface NavItem {
  name: string
  path: string
  icon: Component
  badge?: string
}

const navItems: NavItem[] = [
  { name: 'Overview', path: '/app', icon: LayoutDashboard },
  { name: 'Campaigns', path: '/app/campaigns', icon: Megaphone, badge: 'SOON' },
  { name: 'Submissions', path: '/app/submissions', icon: Video, badge: 'SOON' },
  { name: 'Earnings', path: '/app/earnings', icon: Wallet, badge: 'SOON' },
  { name: 'Payouts', path: '/app/payouts', icon: CreditCard, badge: 'SOON' },
  { name: 'Settings', path: '/app/settings', icon: Settings },
]
</script>

<template>
  <aside class="w-64 bg-[#0f1011] border-r border-[#23252a] flex flex-col justify-between p-4 h-screen sticky top-0 font-sans">
    <div class="space-y-6">
      <!-- App Brand Logo -->
      <NuxtLink to="/" class="flex items-center gap-2.5 px-2">
        <div class="w-7 h-7 rounded-md bg-[#5e6ad2] text-white font-semibold flex items-center justify-center text-xs tracking-tight shadow-sm">
          CI
        </div>
        <span class="font-semibold text-base tracking-tight text-[#f7f8f8]">ClipIN App</span>
      </NuxtLink>

      <!-- Nav Items -->
      <nav class="space-y-1 text-xs font-medium">
        <NuxtLink
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="flex items-center justify-between px-3 py-2 rounded-md transition-colors"
          :class="[
            route.path === item.path
              ? 'bg-[#141516] text-[#f7f8f8] border border-[#23252a] font-semibold'
              : 'text-[#8a8f98] hover:text-[#f7f8f8] hover:bg-[#141516]/50'
          ]"
        >
          <div class="flex items-center gap-2.5">
            <component :is="item.icon" class="w-4 h-4 text-[#8a8f98]" :class="{ 'text-[#5e6ad2]': route.path === item.path }" />
            <span>{{ item.name }}</span>
          </div>

          <span v-if="item.badge" class="text-[9px] font-mono px-1.5 py-0.5 rounded bg-[#18191a] text-[#62666d] border border-[#23252a]">
            {{ item.badge }}
          </span>
        </NuxtLink>
      </nav>
    </div>

    <!-- Status Footer Card -->
    <div class="p-3 rounded-lg bg-[#141516] border border-[#23252a] text-[11px] font-mono text-[#8a8f98] space-y-1">
      <div class="flex items-center gap-1.5 text-[#f7f8f8]">
        <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
        Platform Early Access
      </div>
      <div class="text-[#62666d] text-[10px]">Marketplace Pre-Launch</div>
    </div>
  </aside>
</template>
