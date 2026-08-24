<script setup lang="ts">
import { SignIn } from '@clerk/vue'
</script>

<template>
  <div class="min-h-screen bg-black text-white flex flex-col justify-center items-center px-4 relative selection:bg-white selection:text-black">
    <div class="mb-8 flex items-center gap-3">
      <NuxtLink to="/" class="flex items-center gap-2.5 group">
        <div class="w-7 h-7 rounded bg-white text-black font-bold flex items-center justify-center text-xs tracking-tight">
          CI
        </div>
        <span class="font-semibold text-base tracking-tight text-white">ClipIN</span>
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
