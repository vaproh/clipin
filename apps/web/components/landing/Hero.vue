<script setup lang="ts">
import { Motion } from 'motion-v'
import { ArrowRight } from 'lucide-vue-next'

const mockCampaigns = [
  {
    title: 'Podcast XYZ',
    cpm: '₹150 / 1K views',
    platforms: ['Instagram', 'YouTube'],
    category: 'Talk Show',
    status: 'ACTIVE PREVIEW',
  },
  {
    title: 'Business Creator',
    cpm: '₹120 / 1K views',
    platforms: ['Instagram'],
    category: 'Finance & Tech',
    status: 'ACTIVE PREVIEW',
  },
  {
    title: 'Gaming Podcast',
    cpm: '₹200 / 1K views',
    platforms: ['YouTube Shorts'],
    category: 'Gaming',
    status: 'ACTIVE PREVIEW',
  },
]
</script>

<template>
  <section class="pt-28 pb-16 relative overflow-hidden bg-black text-white">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 space-y-10">
      <!-- Eyebrow & Hero Copy with Motion -->
      <Motion
        :initial="{ opacity: 0, y: 16 }"
        :animate="{ opacity: 1, y: 0 }"
        :transition="{ duration: 0.4 }"
        class="max-w-2xl space-y-5"
      >
        <div class="inline-flex items-center gap-2 px-2.5 py-1 rounded bg-neutral-900 border border-neutral-800 text-xs font-mono text-neutral-300">
          <span class="w-1.5 h-1.5 rounded-full bg-white"></span>
          Performance Clipping Marketplace for India
        </div>

        <h1 class="text-3xl sm:text-5xl font-semibold tracking-tight text-white leading-tight">
          Clip. Post. Get Paid.
        </h1>

        <p class="text-xs sm:text-sm text-neutral-400 leading-relaxed font-normal">
          ClipIN is India's performance clipping marketplace. Find campaigns, publish short-form clips, and earn from verified views.
        </p>

        <!-- Aligned Action Buttons -->
        <div class="flex flex-wrap items-center gap-3 pt-1">
          <NuxtLink
            to="/sign-up?role=clipper"
            class="h-10 px-5 inline-flex items-center justify-center rounded bg-white text-black text-xs font-medium hover:bg-neutral-200 transition-colors gap-2"
          >
            Get started
            <ArrowRight class="w-4 h-4" />
          </NuxtLink>

          <a
            href="#for-campaign-owners"
            class="h-10 px-5 inline-flex items-center justify-center rounded bg-neutral-900 text-neutral-200 border border-neutral-800 text-xs font-medium hover:bg-neutral-800 transition-colors"
          >
            I'm a campaign owner
          </a>
        </div>
      </Motion>

      <!-- Marketplace UI Preview Window with Motion -->
      <Motion
        :initial="{ opacity: 0, y: 24 }"
        :animate="{ opacity: 1, y: 0 }"
        :transition="{ duration: 0.5, delay: 0.1 }"
        id="campaigns"
        class="space-y-3 pt-4"
      >
        <div class="flex items-center justify-between px-0.5">
          <div class="flex items-center gap-2">
            <span class="text-xs font-mono text-neutral-400 uppercase tracking-wider">Marketplace preview</span>
            <span class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-neutral-900 text-neutral-400 border border-neutral-800">Mockup</span>
          </div>
          <span class="text-xs font-mono text-neutral-500">LIVE CAMPAIGNS</span>
        </div>

        <!-- Monochrome Panel Window -->
        <div class="rounded-lg bg-neutral-950 border border-neutral-800 p-5 space-y-4">
          <div class="flex items-center justify-between pb-3 border-b border-neutral-900">
            <div class="flex items-center gap-2">
              <div class="w-2.5 h-2.5 rounded-full bg-neutral-800"></div>
              <div class="w-2.5 h-2.5 rounded-full bg-neutral-800"></div>
              <div class="w-2.5 h-2.5 rounded-full bg-neutral-800"></div>
              <span class="text-xs font-mono text-neutral-500 ml-2">clipin.pro/marketplace</span>
            </div>
            <div class="text-xs font-mono text-neutral-400 hidden sm:block">Filtered: India • Shorts & Reels</div>
          </div>

          <!-- Cards Grid -->
          <div class="grid sm:grid-cols-3 gap-3">
            <div
              v-for="(campaign, idx) in mockCampaigns"
              :key="idx"
              class="rounded bg-neutral-900 border border-neutral-800 p-4 space-y-3 hover:border-neutral-700 transition-colors"
            >
              <div class="flex items-center justify-between">
                <span class="text-[10px] font-mono text-neutral-300 bg-neutral-800 border border-neutral-700 px-2 py-0.5 rounded">
                  {{ campaign.category }}
                </span>
                <span class="text-[10px] font-mono text-neutral-500">{{ campaign.status }}</span>
              </div>

              <div class="space-y-1">
                <h3 class="font-medium text-xs text-white">{{ campaign.title }}</h3>
                <div class="text-xs font-mono text-neutral-300 font-medium">{{ campaign.cpm }}</div>
              </div>

              <div class="flex items-center justify-between pt-2 border-t border-neutral-800 text-[11px] font-mono text-neutral-400">
                <span>Platforms</span>
                <span class="text-white">{{ campaign.platforms.join(' · ') }}</span>
              </div>
            </div>
          </div>
        </div>
      </Motion>
    </div>
  </section>
</template>
