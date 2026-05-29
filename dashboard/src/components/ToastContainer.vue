<script setup lang="ts">
import { useToast } from '@/composables/useToast';
const { toasts } = useToast();
</script>

<template>
  <div class="fixed bottom-4 right-4 z-[100] flex flex-col gap-2">
    <div
      v-for="t in toasts"
      :key="t.id"
      class="rounded-md border px-4 py-3 text-sm shadow-lg backdrop-blur-sm"
      :class="{
        'border-success/30 bg-success/10 text-foreground': t.severity === 'success',
        'border-destructive/30 bg-destructive/10 text-foreground': t.severity === 'error',
        'border-primary/30 bg-primary/10 text-foreground': t.severity === 'info',
        'border-warning/30 bg-warning/10 text-foreground': t.severity === 'warn',
      }"
    >
      {{ t.summary }}
    </div>
  </div>
</template>
