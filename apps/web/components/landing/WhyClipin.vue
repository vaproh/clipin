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
  <section class="py-20 border-t border-[#23252a]/60">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 space-y-12">
      <div class="space-y-3">
        <div class="text-xs font-mono uppercase tracking-wider text-[#5e6ad2]">Value Propositions</div>
        <h2 class="text-2xl sm:text-3xl font-semibold tracking-tight text-[#f7f8f8]">
          Why ClipIN
        </h2>
        <p class="text-sm text-[#8a8f98] max-w-xl">
          Engineered for transparency, performance correctness, and scale across India.
        </p>
      </div>

      <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Motion
          v-for="(feature, idx) in features"
          :key="idx"
          :initial="{ opacity: 0, y: 20 }"
          :animate="{ opacity: 1, y: 0 }"
          :transition="{ duration: 0.4, delay: idx * 0.08 }"
          class="rounded-xl bg-[#0f1011] border border-[#23252a] p-6 space-y-4 hover:border-[#34343a] transition-all"
        >
          <div class="w-10 h-10 rounded-lg bg-[#141516] border border-[#23252a] flex items-center justify-center text-[#5e6ad2]">
            <component :is="feature.icon" class="w-5 h-5 text-[#828fff]" />
          </div>

          <div class="space-y-2">
            <h3 class="font-semibold text-base text-[#f7f8f8]">{{ feature.title }}</h3>
            <p class="text-xs text-[#8a8f98] leading-relaxed font-normal">
              {{ feature.description }}
            </p>
          </div>
        </Motion>
      </div>
    </div>
  </section>
</template>
