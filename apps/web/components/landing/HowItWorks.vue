<script setup lang="ts">
import type { Component } from 'vue'
import { Motion } from 'motion-v'
import { Search, Scissors, ShieldCheck, Wallet } from 'lucide-vue-next'

interface Step {
  number: string
  icon: Component
  title: string
  description: string
}

const steps: Step[] = [
  {
    number: '01',
    icon: Search,
    title: '1. Find a campaign',
    description: 'Browse active campaigns, review CPM rates, content guidelines, and select long-form content you want to clip.',
  },
  {
    number: '02',
    icon: Scissors,
    title: '2. Create & post',
    description: 'Turn the provided source content into engaging YouTube Shorts or Instagram Reels and publish them on your accounts.',
  },
  {
    number: '03',
    icon: ShieldCheck,
    title: '3. Get verified',
    description: 'Submit your published clip URLs. ClipIN continuously tracks and normalizes eligible view metrics.',
  },
  {
    number: '04',
    icon: Wallet,
    title: '4. Get paid',
    description: 'Verified views are automatically converted into earnings according to campaign CPM rates, paid directly in INR.',
  },
]
</script>

<template>
  <section id="how-it-works" class="py-20 border-t border-[#23252a]/60">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 space-y-12">
      <!-- Section Header -->
      <div class="space-y-3">
        <div class="text-xs font-mono uppercase tracking-wider text-[#5e6ad2]">Workflow</div>
        <h2 class="text-2xl sm:text-3xl font-semibold tracking-tight text-[#f7f8f8]">
          How ClipIN Works
        </h2>
        <p class="text-sm text-[#8a8f98] max-w-xl">
          A streamlined 4-step workflow designed for clippers and content creators across India.
        </p>
      </div>

      <!-- 4-Step Cards Grid with Motion -->
      <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Motion
          v-for="(step, idx) in steps"
          :key="step.number"
          :initial="{ opacity: 0, y: 20 }"
          :animate="{ opacity: 1, y: 0 }"
          :transition="{ duration: 0.4, delay: idx * 0.08 }"
          class="rounded-xl bg-[#0f1011] border border-[#23252a] p-6 space-y-4 hover:border-[#34343a] transition-all relative group"
        >
          <div class="flex items-center justify-between">
            <div class="w-10 h-10 rounded-lg bg-[#141516] border border-[#23252a] flex items-center justify-center text-[#f7f8f8] group-hover:border-[#5e6ad2]/50 transition-colors">
              <component :is="step.icon" class="w-5 h-5 text-[#828fff]" />
            </div>
            <span class="font-mono text-xs font-semibold text-[#62666d]">{{ step.number }}</span>
          </div>

          <div class="space-y-2">
            <h3 class="font-semibold text-base text-[#f7f8f8]">{{ step.title }}</h3>
            <p class="text-xs text-[#8a8f98] leading-relaxed font-normal">
              {{ step.description }}
            </p>
          </div>
        </Motion>
      </div>
    </div>
  </section>
</template>
