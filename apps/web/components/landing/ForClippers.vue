<script setup lang="ts">
import { Motion } from 'motion-v'
import { CheckCircle2, Video, ArrowRight } from 'lucide-vue-next'

const clipperPoints = [
  'Monetize your editing skills with flexible performance campaigns',
  'Post clips directly to your owned Instagram Reels & YouTube Shorts',
  'Track verified views in real time with transparent INR earnings',
  'Direct bank & UPI payouts with auditable ledger calculations',
]
</script>

<template>
  <section id="for-clippers" class="py-16 border-t border-neutral-900 bg-black text-white">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 grid lg:grid-cols-2 gap-10 items-center">
      <!-- Left Column: Copy -->
      <Motion
        :initial="{ opacity: 0, x: -16 }"
        :animate="{ opacity: 1, x: 0 }"
        :transition="{ duration: 0.4 }"
        class="space-y-5"
      >
        <div class="space-y-2">
          <div class="text-xs font-mono uppercase tracking-wider text-neutral-400">For Clippers</div>
          <h2 class="text-2xl sm:text-3xl font-semibold tracking-tight text-white leading-tight">
            Turn editing into a <br class="hidden sm:inline" />
            predictable side hustle
          </h2>
          <p class="text-xs text-neutral-400 leading-relaxed font-normal">
            Stop begging for low-paying client retainers. Browse active campaigns, pick long-form podcasts or videos, edit high-performing short clips, and earn every time your posts get verified views.
          </p>
        </div>

        <!-- Bullet List -->
        <div class="space-y-2.5 pt-1">
          <div
            v-for="(point, idx) in clipperPoints"
            :key="idx"
            class="flex items-start gap-2.5 text-xs font-normal text-neutral-300"
          >
            <CheckCircle2 class="w-4 h-4 text-white shrink-0 mt-0.5" />
            <span>{{ point }}</span>
          </div>
        </div>

        <div class="pt-2">
          <NuxtLink
            to="/sign-up?role=clipper"
            class="h-10 px-5 inline-flex items-center justify-center rounded bg-white text-black text-xs font-medium hover:bg-neutral-200 transition-colors gap-2"
          >
            Start clipping now
            <ArrowRight class="w-4 h-4" />
          </NuxtLink>
        </div>
      </Motion>

      <!-- Right Column: Clipper Dashboard Preview Window -->
      <Motion
        :initial="{ opacity: 0, x: 16 }"
        :animate="{ opacity: 1, x: 0 }"
        :transition="{ duration: 0.4, delay: 0.08 }"
        class="rounded bg-neutral-950 border border-neutral-800 p-5 space-y-5"
      >
        <div class="flex items-center justify-between border-b border-neutral-900 pb-3">
          <div class="flex items-center gap-2.5">
            <div class="w-7 h-7 rounded bg-neutral-900 border border-neutral-800 flex items-center justify-center text-white text-xs font-bold font-mono">
              CR
            </div>
            <div>
              <div class="text-xs font-medium text-white">Clipper Workspace</div>
              <div class="text-[10px] font-mono text-neutral-500">Clipper Dashboard Mockup</div>
            </div>
          </div>
          <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-neutral-900 text-neutral-300 border border-neutral-800">
            VERIFIED CLIPPER
          </span>
        </div>

        <!-- Metrics Row -->
        <div class="grid grid-cols-3 gap-2.5 font-mono">
          <div class="p-3 rounded bg-neutral-900 border border-neutral-800 space-y-1">
            <div class="text-[10px] text-neutral-500">Active Clips</div>
            <div class="text-xs font-semibold text-white">12</div>
          </div>
          <div class="p-3 rounded bg-neutral-900 border border-neutral-800 space-y-1">
            <div class="text-[10px] text-neutral-500">Verified Views</div>
            <div class="text-xs font-semibold text-white">42.5K</div>
          </div>
          <div class="p-3 rounded bg-neutral-900 border border-neutral-800 space-y-1">
            <div class="text-[10px] text-neutral-500">Est. Earnings</div>
            <div class="text-xs font-semibold text-white">₹6,375</div>
          </div>
        </div>

        <!-- Submission Status Card -->
        <div class="p-3.5 rounded bg-neutral-900 border border-neutral-800 space-y-2.5">
          <div class="flex items-center justify-between text-xs font-mono">
            <span class="text-white font-medium">Clip Submission #1042</span>
            <span class="text-neutral-300 text-[10px] bg-neutral-800 px-2 py-0.5 rounded border border-neutral-700">VERIFIED</span>
          </div>

          <div class="text-xs text-neutral-400 font-mono flex items-center gap-2">
            <Video class="w-3.5 h-3.5 text-neutral-300" />
            <span class="truncate">https://instagram.com/reels/C9x2A1kL...</span>
          </div>

          <div class="flex items-center justify-between text-[11px] font-mono text-neutral-400 pt-2 border-t border-neutral-800">
            <span>CPM Rate: ₹150 / 1K</span>
            <span class="text-white">Views: 18,200 (₹2,730)</span>
          </div>
        </div>
      </Motion>
    </div>
  </section>
</template>
