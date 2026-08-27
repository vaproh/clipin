<script setup lang="ts">
import { ChevronDown } from 'lucide-vue-next'

const faqs = [
  {
    question: 'What is clipping?',
    answer: 'Clipping is turning existing long-form content (podcasts, YouTube videos, streams) into short-form videos (Reels, Shorts) for social distribution. Content owners set CPM campaigns, and clippers earn from verified view performance.',
  },
  {
    question: 'Who can become a clipper?',
    answer: 'Anyone who can edit engaging short-form video clips and post them to supported platforms. Video editors, students, side-hustlers, micro-creators, and page managers are all eligible.',
  },
  {
    question: 'How do clippers get paid?',
    answer: 'Clippers submit their published clip URLs. ClipIN verifies eligible views, calculates earnings based on campaign CPM terms, and processes payouts directly in INR.',
  },
  {
    question: 'What can campaign owners promote?',
    answer: 'Initially, ClipIN focuses on long-form video creators, podcasters, YouTubers, and publishers looking to expand short-form reach. Content and campaign eligibility rules apply.',
  },
  {
    question: 'Is ClipIN live?',
    answer: 'ClipIN is currently in early-stage setup, onboarding clippers and campaign owners ahead of official campaign launches.',
  },
]

const openIdx = ref<number | null>(0)

const toggle = (idx: number) => {
  openIdx.value = openIdx.value === idx ? null : idx
}
</script>

<template>
  <section class="py-16 border-t border-neutral-900 bg-black text-white">
    <div class="max-w-3xl mx-auto px-4 sm:px-6 space-y-8">
      <div class="space-y-2 text-center">
        <div class="text-xs font-mono uppercase tracking-wider text-neutral-400">FAQ</div>
        <h2 class="text-xl sm:text-2xl font-semibold tracking-tight text-white">
          Frequently Asked Questions
        </h2>
      </div>

      <div class="space-y-2.5">
        <div
          v-for="(faq, idx) in faqs"
          :key="idx"
          class="rounded bg-neutral-950 border border-neutral-800 overflow-hidden transition-colors"
        >
          <button
            @click="toggle(idx)"
            class="w-full p-4 text-left flex items-center justify-between font-medium text-xs text-white hover:text-neutral-200 transition-colors"
            :aria-expanded="openIdx === idx"
            :aria-controls="`faq-panel-${idx}`"
          >
            <span>{{ faq.question }}</span>
            <ChevronDown
              class="w-4 h-4 text-neutral-400 transition-transform duration-200"
              :class="{ 'rotate-180 text-white': openIdx === idx }"
            />
          </button>

          <div
            v-if="openIdx === idx"
            :id="`faq-panel-${idx}`"
            role="region"
            class="px-4 pb-4 text-xs text-neutral-400 leading-relaxed font-normal border-t border-neutral-900 pt-3"
          >
            {{ faq.answer }}
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
