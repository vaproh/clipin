<script setup lang="ts">
import { ref } from 'vue'
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
  <section class="py-20 border-t border-[#23252a]/60">
    <div class="max-w-4xl mx-auto px-4 sm:px-6 space-y-10">
      <div class="space-y-3 text-center">
        <div class="text-xs font-mono uppercase tracking-wider text-[#5e6ad2]">FAQ</div>
        <h2 class="text-2xl sm:text-3xl font-semibold tracking-tight text-[#f7f8f8]">
          Frequently Asked Questions
        </h2>
      </div>

      <div class="space-y-3">
        <div
          v-for="(faq, idx) in faqs"
          :key="idx"
          class="rounded-xl bg-[#0f1011] border border-[#23252a] overflow-hidden transition-colors"
        >
          <button
            @click="toggle(idx)"
            class="w-full p-5 text-left flex items-center justify-between font-medium text-sm text-[#f7f8f8] hover:text-white transition-colors"
          >
            <span>{{ faq.question }}</span>
            <ChevronDown
              class="w-4 h-4 text-[#8a8f98] transition-transform duration-200"
              :class="{ 'rotate-180 text-[#5e6ad2]': openIdx === idx }"
            />
          </button>

          <div
            v-if="openIdx === idx"
            class="px-5 pb-5 text-xs text-[#8a8f98] leading-relaxed font-normal border-t border-[#23252a]/40 pt-3"
          >
            {{ faq.answer }}
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
