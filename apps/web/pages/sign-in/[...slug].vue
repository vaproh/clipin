<template>
  <div class="min-h-screen bg-[#010102] text-[#f7f8f8] flex flex-col justify-center items-center px-4 relative selection:bg-[#5e6ad2]/30">
    <!-- Subtle Background Glow -->
    <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(94,106,210,0.06)_0%,transparent_70%)] pointer-events-none"></div>

    <div class="mb-8 flex items-center gap-3">
      <NuxtLink to="/" class="flex items-center gap-2.5 group">
        <div class="w-8 h-8 rounded-md bg-[#5e6ad2] text-white font-semibold flex items-center justify-center text-xs tracking-tight shadow-sm shadow-[#5e6ad2]/20">
          CI
        </div>
        <span class="font-semibold text-lg tracking-tight text-[#f7f8f8]">ClipIN</span>
      </NuxtLink>
    </div>

    <div class="w-full max-w-md flex justify-center">
      <SignIn
        path="/sign-in"
        routing="path"
        sign-up-url="/sign-up"
        force-redirect-url="/app"
      />
    </div>
  </div>
</template>
