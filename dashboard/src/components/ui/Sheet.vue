<script setup lang="ts">
import { cn } from '@/lib/utils';
import {
  DialogContent,
  DialogOverlay,
  DialogPortal,
  DialogRoot,
} from 'reka-ui';

interface Props {
  open?: boolean;
  side?: 'left' | 'right';
  class?: string;
}
const props = withDefaults(defineProps<Props>(), { side: 'right' });
const emit = defineEmits<{ 'update:open': [value: boolean] }>();
</script>

<template>
  <DialogRoot :open="props.open" @update:open="(v) => emit('update:open', v)">
    <DialogPortal>
      <DialogOverlay class="fixed inset-0 z-50 bg-black/30 backdrop-blur-sm" />
      <DialogContent
        :class="cn(
          'sheet-panel fixed z-50 h-full border border-border bg-background p-6 shadow-lg',
          props.side === 'left' ? 'left-0 top-0 w-80 left-side' : 'right-0 top-0 w-80',
          props.class,
        )"
      >
        <slot />
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>

<style>
.sheet-panel[data-state='open'] {
  animation: slide-in-right 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.sheet-panel[data-state='closed'] {
  animation: slide-out-right 0.2s ease-in;
}
.sheet-panel.left-side[data-state='open'] {
  animation: slide-in-left 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.sheet-panel.left-side[data-state='closed'] {
  animation: slide-out-left 0.2s ease-in;
}
@keyframes slide-in-right {
  from { transform: translateX(100%); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}
@keyframes slide-out-right {
  from { transform: translateX(0); opacity: 1; }
  to { transform: translateX(100%); opacity: 0; }
}
@keyframes slide-in-left {
  from { transform: translateX(-100%); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}
@keyframes slide-out-left {
  from { transform: translateX(0); opacity: 1; }
  to { transform: translateX(-100%); opacity: 0; }
}
</style>
