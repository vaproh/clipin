<script setup lang="ts">
import { Motion } from 'motion-v'
import { ArrowRight } from 'lucide-vue-next'
</script>

<template>
  <section class="py-20 border-t border-[#23252a]/60 relative overflow-hidden">
    <div class="absolute inset-0 bg-[radial-gradient(ellipse_at_center,rgba(94,106,210,0.08)_0%,transparent_70%)] pointer-events-none"></div>

    <Motion
      :initial="{ opacity: 0, scale: 0.96 }"
      :animate="{ opacity: 1, scale: 1 }"
      :transition="{ duration: 0.5 }"
      class="max-w-4xl mx-auto px-4 sm:px-6 text-center space-y-6 relative z-10"
    >
      <h2 class="text-3xl sm:text-5xl font-semibold tracking-tight text-[#f7f8f8]">
        Ready to clip?
      </h2>

      <p class="text-sm sm:text-base text-[#8a8f98] max-w-xl mx-auto font-normal">
        Join ClipIN and be there when the first campaigns go live.
      </p>

      <div class="flex flex-wrap items-center justify-center gap-3 pt-2">
        <NuxtLink
          to="/sign-up"
          class="px-5 py-2.5 rounded-md bg-[#5e6ad2] hover:bg-[#828fff] text-white font-medium text-sm transition-all shadow-md shadow-[#5e6ad2]/20 flex items-center gap-2"
        >
          Join as a clipper
          <ArrowRight class="w-4 h-4" />
        </NuxtLink>

        <NuxtLink
          to="/sign-up"
          class="px-5 py-2.5 rounded-md bg-[#0f1011] hover:bg-[#141516] text-[#d0d6e0] border border-[#23252a] hover:border-[#34343a] font-medium text-sm transition-all"
        >
          I'm a campaign owner
        </NuxtLink>
      </div>
    </Motion>
  </section>
</template>
