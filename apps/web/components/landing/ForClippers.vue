<script setup lang="ts">
import { ArrowRight, CheckCircle2, TrendingUp, Scissors, Eye, DollarSign } from 'lucide-vue-next'
</script>

<template>
  <section id="for-clippers" class="py-20 border-t border-[#23252a]/60">
    <div class="max-w-6xl mx-auto px-4 sm:px-6">
      <div class="grid lg:grid-cols-2 gap-12 items-center">
        <!-- Left: Copy & CTA -->
        <div class="space-y-6">
          <div class="text-xs font-mono uppercase tracking-wider text-[#5e6ad2]">For Clippers</div>

          <h2 class="text-3xl sm:text-4xl font-semibold tracking-tight text-[#f7f8f8] leading-tight">
            Turn editing into a side hustle.
          </h2>

          <p class="text-sm sm:text-base text-[#8a8f98] leading-relaxed">
            Find campaigns, make clips, publish them, and earn based on verified performance. ClipIN gives editors and side-hustlers direct access to performance campaigns.
          </p>

          <ul class="space-y-3 pt-2 text-xs font-mono text-[#d0d6e0]">
            <li class="flex items-center gap-2.5">
              <CheckCircle2 class="w-4 h-4 text-[#5e6ad2]" />
              Direct access to long-form creator content
            </li>
            <li class="flex items-center gap-2.5">
              <CheckCircle2 class="w-4 h-4 text-[#5e6ad2]" />
              Transparent CPM rates & view verification
            </li>
            <li class="flex items-center gap-2.5">
              <CheckCircle2 class="w-4 h-4 text-[#5e6ad2]" />
              Fast payout processing directly in INR
            </li>
          </ul>

          <div class="pt-4">
            <NuxtLink
              to="/sign-up"
              class="inline-flex items-center gap-2 px-5 py-2.5 rounded-md bg-[#5e6ad2] hover:bg-[#828fff] text-white font-medium text-sm transition-all shadow-md shadow-[#5e6ad2]/20"
            >
              Join as a clipper
              <ArrowRight class="w-4 h-4" />
            </NuxtLink>
          </div>
        </div>

        <!-- Right: Clipper Dashboard UI Preview Card -->
        <div class="space-y-2">
          <div class="flex items-center justify-between text-[11px] font-mono text-[#62666d] px-1">
            <span>CLIPPER DASHBOARD PREVIEW</span>
            <span>Mockup Data</span>
          </div>

          <div class="rounded-xl bg-[#0f1011] border border-[#23252a] p-6 space-y-6 shadow-xl">
            <!-- Top Stat Metrics -->
            <div class="grid grid-cols-3 gap-3 pb-4 border-b border-[#23252a]">
              <div class="space-y-1">
                <div class="text-[10px] font-mono text-[#8a8f98]">Campaign CPM</div>
                <div class="text-sm font-semibold text-[#f7f8f8]">₹150 <span class="text-[10px] text-[#62666d]">/ 1K</span></div>
              </div>
              <div class="space-y-1">
                <div class="text-[10px] font-mono text-[#8a8f98]">Verified Views</div>
                <div class="text-sm font-semibold text-[#f7f8f8]">42,500</div>
              </div>
              <div class="space-y-1">
                <div class="text-[10px] font-mono text-[#8a8f98]">Est. Earnings</div>
                <div class="text-sm font-semibold text-[#828fff]">₹6,375</div>
              </div>
            </div>

            <!-- Submission Row Illustration -->
            <div class="space-y-3">
              <div class="text-xs font-mono text-[#8a8f98]">Recent Submission</div>

              <div class="rounded-lg bg-[#141516] border border-[#23252a] p-3.5 flex items-center justify-between text-xs font-mono">
                <div class="space-y-1">
                  <div class="font-medium text-[#f7f8f8] truncate max-w-[200px]">clip_podcast_highlight_04.mp4</div>
                  <div class="text-[10px] text-[#62666d]">Submitted 2h ago • YouTube Shorts</div>
                </div>

                <div class="text-right space-y-1">
                  <span class="inline-block px-2 py-0.5 rounded text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    VERIFIED
                  </span>
                  <div class="text-[10px] text-[#8a8f98]">18.4K Views</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
