<script setup lang="ts">
import { Motion } from 'motion-v'
import { CheckCircle2, TrendingUp, IndianRupee, Video, Eye, ArrowRight } from 'lucide-vue-next'

const clipperPoints = [
  'Monetize your editing skills with flexible performance campaigns',
  'Post clips directly to your owned Instagram Reels & YouTube Shorts',
  'Track verified views in real time with transparent INR earnings',
  'Direct bank & UPI payouts with auditable ledger calculations',
]
</script>

<template>
  <section id="for-clippers" class="py-20 border-t border-[#23252a]/60">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 grid lg:grid-cols-2 gap-12 items-center">
      <!-- Left Column: Copy -->
      <Motion
        :initial="{ opacity: 0, x: -20 }"
        :animate="{ opacity: 1, x: 0 }"
        :transition="{ duration: 0.5 }"
        class="space-y-6"
      >
        <div class="space-y-3">
          <div class="text-xs font-mono uppercase tracking-wider text-[#5e6ad2]">For Clippers</div>
          <h2 class="text-3xl sm:text-4xl font-semibold tracking-tight text-[#f7f8f8]">
            Turn editing into a <br class="hidden sm:inline" />
            <span class="text-[#828fff]">predictable side hustle</span>
          </h2>
          <p class="text-sm text-[#8a8f98] leading-relaxed">
            Stop begging for low-paying client retainers. Browse active campaigns, pick long-form podcasts or videos, edit high-performing short clips, and earn every time your posts get verified views.
          </p>
        </div>

        <!-- Bullet List -->
        <div class="space-y-3 pt-2">
          <div
            v-for="(point, idx) in clipperPoints"
            :key="idx"
            class="flex items-start gap-3 text-xs font-normal text-[#d0d6e0]"
          >
            <CheckCircle2 class="w-4 h-4 text-[#5e6ad2] shrink-0 mt-0.5" />
            <span>{{ point }}</span>
          </div>
        </div>

        <div class="pt-4">
          <NuxtLink
            to="/sign-up"
            class="inline-flex items-center gap-2 px-5 py-2.5 rounded-md bg-[#5e6ad2] hover:bg-[#828fff] text-white font-medium text-sm transition-all shadow-md shadow-[#5e6ad2]/20"
          >
            Start clipping now
            <ArrowRight class="w-4 h-4" />
          </NuxtLink>
        </div>
      </Motion>

      <!-- Right Column: Clipper Dashboard Preview Window -->
      <Motion
        :initial="{ opacity: 0, x: 20 }"
        :animate="{ opacity: 1, x: 0 }"
        :transition="{ duration: 0.5, delay: 0.1 }"
        class="rounded-xl bg-[#0f1011] border border-[#23252a] p-6 space-y-6 shadow-xl"
      >
        <div class="flex items-center justify-between border-b border-[#23252a] pb-4">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-full bg-[#5e6ad2]/20 border border-[#5e6ad2]/40 flex items-center justify-center text-[#828fff] text-xs font-bold">
              CR
            </div>
            <div>
              <div class="text-xs font-semibold text-[#f7f8f8]">Clipper Workspace</div>
              <div class="text-[10px] font-mono text-[#8a8f98]">Clipper Dashboard Mockup</div>
            </div>
          </div>
          <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            VERIFIED CLIPPER
          </span>
        </div>

        <!-- Metrics Row -->
        <div class="grid grid-cols-3 gap-3 font-mono">
          <div class="p-3 rounded-lg bg-[#141516] border border-[#23252a] space-y-1">
            <div class="text-[10px] text-[#8a8f98]">Active Clips</div>
            <div class="text-sm font-semibold text-[#f7f8f8]">12</div>
          </div>
          <div class="p-3 rounded-lg bg-[#141516] border border-[#23252a] space-y-1">
            <div class="text-[10px] text-[#8a8f98]">Verified Views</div>
            <div class="text-sm font-semibold text-[#f7f8f8]">42.5K</div>
          </div>
          <div class="p-3 rounded-lg bg-[#141516] border border-[#23252a] space-y-1">
            <div class="text-[10px] text-[#8a8f98]">Est. Earnings</div>
            <div class="text-sm font-semibold text-emerald-400">₹6,375</div>
          </div>
        </div>

        <!-- Submission Status Card -->
        <div class="p-4 rounded-lg bg-[#141516] border border-[#23252a] space-y-3">
          <div class="flex items-center justify-between text-xs font-mono">
            <span class="text-[#f7f8f8] font-semibold">Clip Submission #1042</span>
            <span class="text-emerald-400 text-[10px] bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">VERIFIED</span>
          </div>

          <div class="text-xs text-[#8a8f98] font-mono flex items-center gap-2">
            <Video class="w-3.5 h-3.5 text-[#5e6ad2]" />
            <span class="truncate">https://instagram.com/reels/C9x2A1kL...</span>
          </div>

          <div class="flex items-center justify-between text-[11px] font-mono text-[#8a8f98] pt-2 border-t border-[#23252a]">
            <span>CPM Rate: ₹150 / 1K</span>
            <span class="text-[#d0d6e0]">Views: 18,200 (₹2,730)</span>
          </div>
        </div>
      </Motion>
    </div>
  </section>
</template>
