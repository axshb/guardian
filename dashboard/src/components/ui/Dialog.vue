<script setup lang="ts">
import { cn } from '@/lib/utils';
import {
  DialogClose,
  DialogContent,
  DialogOverlay,
  DialogPortal,
  DialogRoot,
  DialogTitle,
} from 'reka-ui';

interface Props {
  open?: boolean;
  class?: string;
}
const props = defineProps<Props>();
const emit = defineEmits<{ 'update:open': [value: boolean] }>();
</script>

<template>
  <DialogRoot :open="props.open" @update:open="(v) => emit('update:open', v)">
    <DialogPortal>
      <DialogOverlay class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" />
      <DialogContent
        :class="cn(
          'fixed left-1/2 top-1/2 z-50 w-full max-w-lg -translate-x-1/2 -translate-y-1/2 border border-border bg-background p-6 shadow-lg rounded-xl',
          props.class,
        )"
      >
        <slot />
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>
