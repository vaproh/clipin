<script setup lang="ts">
import { Cpu, Terminal, ShieldCheck, Activity } from "lucide-vue-next";
</script>

<template>
  <div
    class="min-h-screen bg-neutral-950 text-neutral-100 flex flex-col font-sans selection:bg-neutral-800 selection:text-neutral-100"
  >
    <!-- Header -->
    <header
      class="border-b border-neutral-800/80 bg-neutral-950/80 backdrop-blur sticky top-0 z-50"
    >
      <div
        class="max-w-6xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between"
      >
        <div class="flex items-center gap-3">
          <div
            class="w-8 h-8 rounded-lg bg-neutral-100 text-neutral-950 font-bold flex items-center justify-center text-sm tracking-tighter"
          >
            CI
          </div>
          <NuxtLink
            to="/"
            class="font-semibold text-lg tracking-tight hover:opacity-80 transition-opacity"
          >
            ClipIN
            <span
              class="text-xs px-2 py-0.5 rounded bg-neutral-800 text-neutral-400 font-mono font-normal"
              >DEV SCAFFOLD</span
            >
          </NuxtLink>
        </div>

        <div class="flex items-center gap-4 text-xs font-mono text-neutral-400">
          <div
            class="flex items-center gap-1.5 px-2.5 py-1 rounded bg-neutral-900 border border-neutral-800"
          >
            <span
              class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"
            ></span>
            System Ready
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 max-w-6xl w-full mx-auto px-4 sm:px-6 py-10">
      <slot />
    </main>

    <!-- Footer -->
    <footer
      class="border-t border-neutral-800/80 bg-neutral-950 py-6 text-xs text-neutral-500 font-mono"
    >
      <div
        class="max-w-6xl mx-auto px-4 sm:px-6 flex flex-col sm:flex-row items-center justify-between gap-4"
      >
        <div>
          ClipIN Marketplace Scaffold • India Performance Clipping Platform
        </div>
        <div class="flex items-center gap-6 text-neutral-400">
          <a
            href="/api/docs"
            target="_blank"
            class="hover:text-neutral-200 transition-colors"
            >API OpenAPI Docs</a
          >
          <span>•</span>
          <span>Nuxt 3 + Vue 3 + Go</span>
        </div>
      </div>
    </footer>
  </div>
</template>
