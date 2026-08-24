<template>
  <footer class="border-t border-[#23252a] bg-[#010102] py-12 text-xs font-mono text-[#8a8f98]">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 flex flex-col md:flex-row items-center justify-between gap-6">
      <div class="flex items-center gap-3">
        <div class="w-6 h-6 rounded-md bg-[#5e6ad2] text-white font-semibold flex items-center justify-center text-[10px] tracking-tight">
          CI
        </div>
        <span class="font-semibold text-sm tracking-tight text-[#f7f8f8]">ClipIN</span>
        <span class="text-[#62666d]">© {{ new Date().getFullYear() }} ClipIN</span>
      </div>

      <div class="flex flex-wrap items-center justify-center gap-6 text-[#8a8f98]">
        <a href="#how-it-works" class="hover:text-[#f7f8f8] transition-colors">How it works</a>
        <a href="#for-clippers" class="hover:text-[#f7f8f8] transition-colors">For Clippers</a>
        <a href="#for-campaign-owners" class="hover:text-[#f7f8f8] transition-colors">For Campaign Owners</a>
        <a href="#" class="hover:text-[#f7f8f8] transition-colors">Privacy</a>
        <a href="#" class="hover:text-[#f7f8f8] transition-colors">Terms</a>
        <a href="mailto:support@clipin.pro" class="hover:text-[#f7f8f8] transition-colors">Contact</a>
      </div>
    </div>
  </footer>
</template>
