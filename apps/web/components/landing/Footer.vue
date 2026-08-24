<template>
  <footer class="border-t border-neutral-900 bg-black py-10 text-xs font-mono text-neutral-400">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 flex flex-col sm:flex-row items-center justify-between gap-4">
      <div class="flex items-center gap-2.5">
        <div class="w-5 h-5 rounded bg-white text-black font-bold flex items-center justify-center text-[9px]">
          CI
        </div>
        <span class="font-semibold text-xs text-white">ClipIN</span>
        <span class="text-neutral-600">© {{ new Date().getFullYear() }} ClipIN</span>
      </div>

      <div class="flex items-center gap-5 text-neutral-400 text-xs font-mono">
        <a href="#how-it-works" class="hover:text-white transition-colors">How it works</a>
        <a href="#for-clippers" class="hover:text-white transition-colors">For Clippers</a>
        <a href="#for-campaign-owners" class="hover:text-white transition-colors">For Campaign Owners</a>
        <a href="mailto:support@clipin.pro" class="hover:text-white transition-colors">Contact</a>
      </div>
    </div>
  </footer>
</template>
