<script setup lang="ts">
const modelValue = defineModel<boolean>({ default: false })
const reason = defineModel<string>('reason', { default: '' })
defineEmits<{ confirm: [] }>()
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div v-if="modelValue" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" @click.self="modelValue = false">
        <div class="bg-neutral-950 border border-neutral-800 rounded-lg p-6 w-full max-w-sm space-y-4">
          <h3 class="text-sm font-semibold text-white">Reject Submission</h3>
          <textarea v-model="reason" placeholder="Reason for rejection (optional)" class="w-full h-24 bg-neutral-900 border border-neutral-800 rounded p-3 text-xs text-white placeholder:text-neutral-500 resize-none focus:outline-none focus:border-neutral-600" />
          <div class="flex justify-end gap-2">
            <button class="px-3 py-1.5 text-xs text-neutral-400 hover:text-white transition-colors" @click="modelValue = false">Cancel</button>
            <button class="px-3 py-1.5 text-xs bg-white text-black rounded font-medium hover:bg-neutral-200 transition-colors" @click="$emit('confirm')">Reject</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
