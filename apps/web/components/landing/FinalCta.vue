<script setup lang="ts">
import { Motion } from 'motion-v'
import { ArrowRight } from 'lucide-vue-next'
</script>

<template>
  <section class="py-16 border-t border-neutral-900 bg-black text-white relative overflow-hidden">
    <Motion
      :initial="{ opacity: 0, scale: 0.98 }"
      :animate="{ opacity: 1, scale: 1 }"
      :transition="{ duration: 0.4 }"
      class="max-w-3xl mx-auto px-4 sm:px-6 text-center space-y-5 relative z-10"
    >
      <h2 class="text-2xl sm:text-4xl font-semibold tracking-tight text-white">
        Ready to clip?
      </h2>

      <p class="text-xs sm:text-sm text-neutral-400 max-w-md mx-auto font-normal">
        Join ClipIN and be there when the first campaigns go live.
      </p>

      <div class="flex flex-wrap items-center justify-center gap-3 pt-1">
        <NuxtLink
          to="/sign-up?role=clipper"
          class="h-10 px-5 inline-flex items-center justify-center rounded bg-white text-black text-xs font-medium hover:bg-neutral-200 transition-colors gap-2"
        >
          Join as a clipper
          <ArrowRight class="w-4 h-4" />
        </NuxtLink>

        <NuxtLink
          to="/sign-up?role=owner"
          class="h-10 px-5 inline-flex items-center justify-center rounded bg-neutral-900 text-neutral-200 border border-neutral-800 text-xs font-medium hover:bg-neutral-800 transition-colors"
        >
          I'm a campaign owner
        </NuxtLink>
      </div>
    </Motion>
  </section>
</template>
