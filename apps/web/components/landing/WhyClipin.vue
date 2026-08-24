<script setup lang="ts">
import type { Component } from 'vue'
import { Motion } from 'motion-v'
import { BarChart2, IndianRupee, Eye, Store } from 'lucide-vue-next'

interface Feature {
  icon: Component
  title: string
  description: string
}

const features: Feature[] = [
  {
    icon: BarChart2,
    title: 'Performance based',
    description: 'Pay for verified distribution rather than vague reach promises or upfront fixed fees.',
  },
  {
    icon: IndianRupee,
    title: 'Built for India',
    description: 'INR-native pricing, Indian payout infrastructure, and local creator ecosystem focus.',
  },
  {
    icon: Eye,
    title: 'Transparent',
    description: 'Campaign rules, verified views, earnings, and payout states are auditable and clear.',
  },
  {
    icon: Store,
    title: 'One marketplace',
    description: 'A centralized platform for campaign owners and clippers to connect and collaborate.',
  },
]
</script>

<template>
  <section id="why-clipin" class="py-16 border-t border-neutral-900 bg-black text-white">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 space-y-10">
      <div class="space-y-2">
        <div class="text-xs font-mono uppercase tracking-wider text-neutral-400">Value Propositions</div>
        <h2 class="text-xl sm:text-2xl font-semibold tracking-tight text-white">
          Why ClipIN
        </h2>
        <p class="text-xs text-neutral-400 max-w-xl">
          Engineered for transparency, performance correctness, and scale across India.
        </p>
      </div>

      <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-3">
        <Motion
          v-for="(feature, idx) in features"
          :key="idx"
          :initial="{ opacity: 0, y: 16 }"
          :animate="{ opacity: 1, y: 0 }"
          :transition="{ duration: 0.35, delay: idx * 0.06 }"
          class="rounded bg-neutral-950 border border-neutral-800 p-5 space-y-3.5 hover:border-neutral-700 transition-colors"
        >
          <div class="w-8 h-8 rounded bg-neutral-900 border border-neutral-800 flex items-center justify-center text-white">
            <component :is="feature.icon" class="w-4 h-4 text-neutral-300" />
          </div>

          <div class="space-y-1.5">
            <h3 class="font-medium text-xs text-white">{{ feature.title }}</h3>
            <p class="text-xs text-neutral-400 leading-relaxed font-normal">
              {{ feature.description }}
            </p>
          </div>
        </Motion>
      </div>
    </div>
  </section>
</template>
