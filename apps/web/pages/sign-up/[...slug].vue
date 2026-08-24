<template>
  <div class="min-h-screen bg-black text-white flex flex-col justify-center items-center px-4 relative selection:bg-white selection:text-black py-12">
    <div class="mb-8 flex items-center gap-3">
      <NuxtLink to="/" class="flex items-center gap-2.5 group">
        <div class="w-7 h-7 rounded bg-white text-black font-bold flex items-center justify-center text-xs tracking-tight">
          CI
        </div>
        <span class="font-semibold text-base tracking-tight text-white">ClipIN</span>
      </NuxtLink>
    </div>

    <div class="w-full max-w-md flex justify-center min-h-[480px]">
      <ClientOnly>
        <SignUp
          path="/sign-up"
          routing="path"
          sign-in-url="/sign-in"
          force-redirect-url="/app"
        />
        <template #fallback>
          <div class="w-full max-w-sm p-8 rounded bg-neutral-950 border border-neutral-800 text-center text-xs font-mono text-neutral-400 space-y-3">
            <div class="w-5 h-5 border-2 border-neutral-700 border-t-white rounded-full animate-spin mx-auto"></div>
            <div>Loading sign up...</div>
          </div>
        </template>
      </ClientOnly>
    </div>
  </div>
</template>
