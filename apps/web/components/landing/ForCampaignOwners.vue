<script setup lang="ts">
import { ArrowRight, CheckCircle2, Sliders, Target, Layers } from 'lucide-vue-next'
</script>

<template>
  <section id="for-campaign-owners" class="py-20 border-t border-[#23252a]/60">
    <div class="max-w-6xl mx-auto px-4 sm:px-6">
      <div class="grid lg:grid-cols-2 gap-12 items-center">
        <!-- Left: Campaign Owner Setup UI Illustration -->
        <div class="space-y-2 order-2 lg:order-1">
          <div class="flex items-center justify-between text-[11px] font-mono text-[#62666d] px-1">
            <span>CAMPAIGN SETUP PREVIEW</span>
            <span>Mockup Data</span>
          </div>

          <div class="rounded-xl bg-[#0f1011] border border-[#23252a] p-6 space-y-5 shadow-xl">
            <div class="flex items-center justify-between pb-3 border-b border-[#23252a]">
              <div>
                <h4 class="font-semibold text-sm text-[#f7f8f8]">Campaign Setup</h4>
                <p class="text-[11px] text-[#8a8f98] font-mono">Performance Campaign Configuration</p>
              </div>
              <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-[#18191a] text-[#5e6ad2] border border-[#23252a]">
                DRAFT
              </span>
            </div>

            <!-- Parameters Grid -->
            <div class="space-y-3 font-mono text-xs">
              <div class="flex items-center justify-between py-1.5 border-b border-[#23252a]/50">
                <span class="text-[#8a8f98]">CPM</span>
                <span class="font-semibold text-[#f7f8f8]">₹150 <span class="text-[10px] text-[#62666d]">/ 1K views</span></span>
              </div>

              <div class="flex items-center justify-between py-1.5 border-b border-[#23252a]/50">
                <span class="text-[#8a8f98]">Budget</span>
                <span class="font-semibold text-[#f7f8f8]">₹50,000</span>
              </div>

              <div class="flex items-center justify-between py-1.5 border-b border-[#23252a]/50">
                <span class="text-[#8a8f98]">Platforms</span>
                <span class="font-semibold text-[#f7f8f8]">Instagram · YouTube</span>
              </div>

              <div class="flex items-center justify-between py-1.5">
                <span class="text-[#8a8f98]">Audience</span>
                <span class="font-semibold text-[#f7f8f8]">India</span>
              </div>
            </div>

            <div class="pt-2">
              <div class="w-full py-2 rounded bg-[#141516] border border-[#23252a] text-center text-xs font-mono text-[#8a8f98]">
                Est. Total Verified Reach: ~3.33M Views
              </div>
            </div>
          </div>
        </div>

        <!-- Right: Copy & CTA -->
        <div class="space-y-6 order-1 lg:order-2">
          <div class="text-xs font-mono uppercase tracking-wider text-[#5e6ad2]">For Campaign Owners</div>

          <h2 class="text-3xl sm:text-4xl font-semibold tracking-tight text-[#f7f8f8] leading-tight">
            Turn one piece of content into distributed reach.
          </h2>

          <p class="text-sm sm:text-base text-[#8a8f98] leading-relaxed">
            Launch performance-based clipping campaigns and let independent clippers distribute your content across short-form platforms.
          </p>

          <ul class="space-y-3 pt-2 text-xs font-mono text-[#d0d6e0]">
            <li class="flex items-center gap-2.5">
              <CheckCircle2 class="w-4 h-4 text-[#5e6ad2]" />
              Set custom CPM limits & total campaign budget
            </li>
            <li class="flex items-center gap-2.5">
              <CheckCircle2 class="w-4 h-4 text-[#5e6ad2]" />
              Pay strictly for verified eligible view metrics
            </li>
            <li class="flex items-center gap-2.5">
              <CheckCircle2 class="w-4 h-4 text-[#5e6ad2]" />
              Distribute across YouTube Shorts and Instagram Reels
            </li>
          </ul>

          <div class="pt-4">
            <NuxtLink
              to="/sign-up"
              class="inline-flex items-center gap-2 px-5 py-2.5 rounded-md bg-[#5e6ad2] hover:bg-[#828fff] text-white font-medium text-sm transition-all shadow-md shadow-[#5e6ad2]/20"
            >
              Launch a campaign
              <ArrowRight class="w-4 h-4" />
            </NuxtLink>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
