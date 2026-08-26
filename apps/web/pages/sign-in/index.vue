<script setup lang="ts">
definePageMeta({
  layout: false,
})

useSeoMeta({
  title: 'Sign In - ClipIN',
  robots: 'noindex, nofollow',
})
</script>

<template>
  <div class="min-h-screen bg-black text-white flex flex-col justify-center items-center px-4 relative selection:bg-white selection:text-black py-12">
    <!-- Subtle Radial Glow Background -->
    <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(255,255,255,0.03)_0%,transparent_65%)] pointer-events-none"></div>

    <!-- Header Brand Logo -->
    <div class="mb-6 flex items-center gap-3 relative z-10">
      <NuxtLink to="/" class="inline-flex items-center gap-2.5 group">
        <div class="w-7 h-7 rounded bg-white text-black font-bold flex items-center justify-center text-xs tracking-tight shadow-sm">
          CI
        </div>
        <span class="font-semibold text-base tracking-tight text-white">ClipIN</span>
      </NuxtLink>
    </div>

    <!-- SignIn Component Container -->
    <div class="w-full max-w-md flex justify-center min-h-[480px] relative z-10">
      <ClientOnly>
        <SignIn
          path="/sign-in"
          routing="path"
          sign-up-url="/sign-up"
          force-redirect-url="/app"
        />
        <template #fallback>
          <div class="w-full max-w-sm p-8 rounded bg-neutral-950 border border-neutral-800 text-center text-xs font-mono text-neutral-400 space-y-3">
            <div class="w-5 h-5 border-2 border-neutral-700 border-t-white rounded-full animate-spin mx-auto"></div>
            <div>Loading sign in...</div>
          </div>
        </template>
      </ClientOnly>
    </div>
  </div>
</template>
