<script setup lang="ts">
import type { Ref } from 'vue'

interface PwaState {
  needRefresh: Ref<boolean>
  showInstallPrompt: Ref<boolean>
  isPWAInstalled: Ref<boolean>
  install: () => Promise<void>
  updateServiceWorker: (reload?: boolean) => Promise<void>
}

const online = useOnline()
const { $pwa } = useNuxtApp() as unknown as { $pwa?: PwaState }
</script>

<template>
  <ClientOnly>
    <div v-if="!online" class="fixed inset-x-0 top-14 z-20 border-b border-amber-500/30 bg-amber-950/90 px-4 py-2 text-center text-[11px] font-mono text-amber-200 backdrop-blur">
      Offline mode. Cached pages remain available; changes will resume when you reconnect.
    </div>

    <div v-if="$pwa?.needRefresh" class="fixed inset-x-3 bottom-20 z-40 mx-auto flex max-w-md items-center justify-between gap-3 rounded border border-neutral-700 bg-neutral-950 p-3 text-xs shadow-2xl md:bottom-4">
      <span class="text-neutral-300">A new ClipIN version is ready.</span>
      <button class="shrink-0 font-mono text-white underline underline-offset-4" @click="$pwa.updateServiceWorker(true)">
        Refresh
      </button>
    </div>

    <div v-if="$pwa?.showInstallPrompt && !$pwa?.isPWAInstalled" class="fixed inset-x-3 bottom-20 z-40 mx-auto flex max-w-md items-center justify-between gap-3 rounded border border-neutral-700 bg-neutral-950 p-3 text-xs shadow-2xl md:bottom-4">
      <div>
        <div class="font-medium text-white">Install ClipIN</div>
        <div class="text-neutral-500">Open campaigns faster from your home screen.</div>
      </div>
      <button class="shrink-0 rounded bg-white px-3 py-1.5 font-medium text-black" @click="$pwa.install()">
        Install
      </button>
    </div>
  </ClientOnly>
</template>
