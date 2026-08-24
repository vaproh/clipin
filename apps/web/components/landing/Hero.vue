<script setup lang="ts">
import { Motion } from 'motion-v'
import { ArrowRight, Play, TrendingUp, Sparkles, Video, Instagram, Youtube } from 'lucide-vue-next'

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
  <section class="pt-32 pb-20 relative overflow-hidden">
    <!-- Subtle Background Radial Gradient -->
    <div class="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[400px] bg-[radial-gradient(ellipse_at_center,rgba(94,106,210,0.07)_0%,transparent_70%)] pointer-events-none"></div>

    <div class="max-w-6xl mx-auto px-4 sm:px-6 relative z-10 space-y-12">
      <!-- Eyebrow & Hero Copy with Motion -->
      <Motion
        :initial="{ opacity: 0, y: 20 }"
        :animate="{ opacity: 1, y: 0 }"
        :transition="{ duration: 0.5, ease: 'easeOut' }"
        class="max-w-3xl space-y-6"
      >
        <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-[#0f1011] border border-[#23252a] text-xs font-mono text-[#8a8f98]">
          <span class="w-1.5 h-1.5 rounded-full bg-[#5e6ad2] animate-pulse"></span>
          Performance Clipping Marketplace for India
        </div>

        <h1 class="text-4xl sm:text-6xl font-semibold tracking-tight text-[#f7f8f8] leading-[1.1]">
          Clip. Post. <br class="hidden sm:inline" />
          <span class="text-[#828fff]">Get Paid.</span>
        </h1>

        <p class="text-base sm:text-lg text-[#8a8f98] leading-relaxed max-w-2xl font-normal">
          ClipIN is India's performance clipping marketplace. Find campaigns, publish short-form clips, and earn from verified views.
        </p>

        <div class="flex flex-wrap items-center gap-3 pt-2">
          <NuxtLink
            to="/sign-up"
            class="px-5 py-2.5 rounded-md bg-[#5e6ad2] hover:bg-[#828fff] text-white font-medium text-sm transition-all shadow-md shadow-[#5e6ad2]/20 flex items-center gap-2"
          >
            Get started
            <ArrowRight class="w-4 h-4" />
          </NuxtLink>

          <a
            href="#for-campaign-owners"
            class="px-5 py-2.5 rounded-md bg-[#0f1011] hover:bg-[#141516] text-[#d0d6e0] border border-[#23252a] hover:border-[#34343a] font-medium text-sm transition-all"
          >
            I'm a campaign owner
          </a>
        </div>
      </Motion>

      <!-- Marketplace UI Preview Window with Motion -->
      <Motion
        :initial="{ opacity: 0, y: 30 }"
        :animate="{ opacity: 1, y: 0 }"
        :transition="{ duration: 0.6, delay: 0.1, ease: 'easeOut' }"
        id="campaigns"
        class="space-y-3 pt-6"
      >
        <div class="flex items-center justify-between px-1">
          <div class="flex items-center gap-2">
            <span class="text-xs font-mono uppercase tracking-wider text-[#8a8f98]">Marketplace preview</span>
            <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-[#18191a] text-[#62666d] border border-[#23252a]">Mockup Data</span>
          </div>
          <span class="text-xs font-mono text-[#62666d]">LIVE CAMPAIGNS</span>
        </div>

        <!-- Linear Style Panel Window -->
        <div class="rounded-xl bg-[#0f1011] border border-[#23252a] p-4 sm:p-6 shadow-2xl shadow-black/80 space-y-4">
          <div class="flex items-center justify-between pb-4 border-b border-[#23252a]/70">
            <div class="flex items-center gap-2">
              <div class="w-3 h-3 rounded-full bg-[#23252a]"></div>
              <div class="w-3 h-3 rounded-full bg-[#23252a]"></div>
              <div class="w-3 h-3 rounded-full bg-[#23252a]"></div>
              <span class="text-xs font-mono text-[#62666d] ml-2">clipin.pro/marketplace</span>
            </div>
            <div class="text-xs font-mono text-[#8a8f98] hidden sm:block">Filtered: India • Shorts & Reels</div>
          </div>

          <!-- Cards Grid -->
          <div class="grid sm:grid-cols-3 gap-4">
            <div
              v-for="(campaign, idx) in mockCampaigns"
              :key="idx"
              class="rounded-lg bg-[#141516] border border-[#23252a] p-4 space-y-3 hover:border-[#34343a] transition-all"
            >
              <div class="flex items-center justify-between">
                <span class="text-[10px] font-mono text-[#5e6ad2] bg-[#5e6ad2]/10 border border-[#5e6ad2]/20 px-2 py-0.5 rounded">
                  {{ campaign.category }}
                </span>
                <span class="text-[10px] font-mono text-[#8a8f98]">{{ campaign.status }}</span>
              </div>

              <div>
                <h3 class="font-semibold text-sm text-[#f7f8f8]">{{ campaign.title }}</h3>
                <div class="text-xs font-mono text-[#d0d6e0] font-medium pt-1">{{ campaign.cpm }}</div>
              </div>

              <div class="flex items-center gap-1.5 pt-2 border-t border-[#23252a]/60 text-[11px] font-mono text-[#8a8f98]">
                <span>Platforms:</span>
                <span class="text-[#d0d6e0]">{{ campaign.platforms.join(' · ') }}</span>
              </div>
            </div>
          </div>
        </div>
      </Motion>
    </div>
  </section>
</template>
