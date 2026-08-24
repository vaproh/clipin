<script setup lang="ts">
import { Motion } from 'motion-v'
import { CheckCircle2, ArrowRight } from 'lucide-vue-next'

const ownerPoints = [
  'Pay strictly for verified view performance instead of fixed upfront retainers',
  'Set target CPM, budget cap, content guidelines, and target platforms',
  'Automated views verification pipeline filters low-quality activity',
  'Audit every view, submission, and payout transaction in a unified ledger',
]
</script>

<template>
  <section id="for-campaign-owners" class="py-16 border-t border-neutral-900 bg-black text-white">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 grid lg:grid-cols-2 gap-10 items-center">
      <!-- Left Column: Setup Preview Window -->
      <Motion
        :initial="{ opacity: 0, x: -16 }"
        :animate="{ opacity: 1, x: 0 }"
        :transition="{ duration: 0.4 }"
        class="rounded bg-neutral-950 border border-neutral-800 p-5 space-y-5 order-2 lg:order-1"
      >
        <div class="flex items-center justify-between border-b border-neutral-900 pb-3">
          <div class="flex items-center gap-2.5">
            <div class="w-7 h-7 rounded bg-neutral-900 border border-neutral-800 flex items-center justify-center text-white text-xs font-bold font-mono">
              CO
            </div>
            <div>
              <div class="text-xs font-medium text-white">Campaign Setup</div>
              <div class="text-[10px] font-mono text-neutral-500">Owner Dashboard Mockup</div>
            </div>
          </div>
          <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-neutral-900 text-neutral-300 border border-neutral-800">
            CAMPAIGN DRAFT
          </span>
        </div>

        <!-- Form Setup Fields Mockup -->
        <div class="space-y-3 font-mono text-xs">
          <div class="space-y-1">
            <label class="text-[10px] text-neutral-400">Campaign Title</label>
            <div class="p-2.5 rounded bg-neutral-900 border border-neutral-800 text-white">
              Tech Podcast Season 3 Clipping Campaign
            </div>
          </div>

          <div class="grid grid-cols-2 gap-2.5">
            <div class="space-y-1">
              <label class="text-[10px] text-neutral-400">CPM Offer (per 1K views)</label>
              <div class="p-2.5 rounded bg-neutral-900 border border-neutral-800 text-white font-semibold">
                ₹150
              </div>
            </div>
            <div class="space-y-1">
              <label class="text-[10px] text-neutral-400">Total Budget Cap</label>
              <div class="p-2.5 rounded bg-neutral-900 border border-neutral-800 text-white font-semibold">
                ₹50,000
              </div>
            </div>
          </div>

          <div class="space-y-1">
            <label class="text-[10px] text-neutral-400">Target Platforms & Region</label>
            <div class="p-2.5 rounded bg-neutral-900 border border-neutral-800 text-neutral-300 flex items-center justify-between text-[11px]">
              <span>Instagram Reels · YouTube Shorts</span>
              <span class="text-white">India</span>
            </div>
          </div>
        </div>
      </Motion>

      <!-- Right Column: Copy -->
      <Motion
        :initial="{ opacity: 0, x: 16 }"
        :animate="{ opacity: 1, x: 0 }"
        :transition="{ duration: 0.4, delay: 0.08 }"
        class="space-y-5 order-1 lg:order-2"
      >
        <div class="space-y-2">
          <div class="text-xs font-mono uppercase tracking-wider text-neutral-400">For Campaign Owners</div>
          <h2 class="text-2xl sm:text-3xl font-semibold tracking-tight text-white leading-tight">
            Turn one piece of content into <br class="hidden sm:inline" />
            distributed reach
          </h2>
          <p class="text-xs text-neutral-400 leading-relaxed font-normal">
            Content creators, podcasters, and brand owners can launch performance clipping campaigns in minutes. Set your budget and CPM rate — clippers handle editing and distribution.
          </p>
        </div>

        <!-- Bullet List -->
        <div class="space-y-2.5 pt-1">
          <div
            v-for="(point, idx) in ownerPoints"
            :key="idx"
            class="flex items-start gap-2.5 text-xs font-normal text-neutral-300"
          >
            <CheckCircle2 class="w-4 h-4 text-white shrink-0 mt-0.5" />
            <span>{{ point }}</span>
          </div>
        </div>

        <div class="pt-2">
          <NuxtLink
            to="/sign-up"
            class="h-10 px-5 inline-flex items-center justify-center rounded bg-white text-black text-xs font-medium hover:bg-neutral-200 transition-colors gap-2"
          >
            Launch a campaign
            <ArrowRight class="w-4 h-4" />
          </NuxtLink>
        </div>
      </Motion>
    </div>
  </section>
</template>
