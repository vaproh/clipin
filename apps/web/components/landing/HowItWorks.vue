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
  <section id="how-it-works" class="py-16 border-t border-neutral-900 bg-black text-white">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 space-y-10">
      <!-- Section Header -->
      <div class="space-y-2">
        <div class="text-xs font-mono uppercase tracking-wider text-neutral-400">Workflow</div>
        <h2 class="text-xl sm:text-2xl font-semibold tracking-tight text-white">
          How ClipIN Works
        </h2>
        <p class="text-xs text-neutral-400 max-w-xl">
          A streamlined 4-step workflow designed for clippers and content creators across India.
        </p>
      </div>

      <!-- 4-Step Cards Grid -->
      <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-3">
        <Motion
          v-for="(step, idx) in steps"
          :key="step.number"
          :initial="{ opacity: 0, y: 16 }"
          :animate="{ opacity: 1, y: 0 }"
          :transition="{ duration: 0.35, delay: idx * 0.06 }"
          class="rounded bg-neutral-950 border border-neutral-800 p-5 space-y-4 hover:border-neutral-700 transition-colors"
        >
          <div class="flex items-center justify-between">
            <div class="w-8 h-8 rounded bg-neutral-900 border border-neutral-800 flex items-center justify-center text-white">
              <component :is="step.icon" class="w-4 h-4 text-neutral-300" />
            </div>
            <span class="font-mono text-xs font-semibold text-neutral-500">{{ step.number }}</span>
          </div>

          <div class="space-y-1.5">
            <h3 class="font-medium text-xs text-white">{{ step.title }}</h3>
            <p class="text-xs text-neutral-400 leading-relaxed font-normal">
              {{ step.description }}
            </p>
          </div>
        </Motion>
      </div>
    </div>
  </section>
</template>
