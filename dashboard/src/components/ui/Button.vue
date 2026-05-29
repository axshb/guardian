<script setup lang="ts">
import { cn } from '@/lib/utils';

interface Props {
  variant?: 'default' | 'ghost' | 'danger' | 'secondary';
  size?: 'default' | 'sm';
  disabled?: boolean;
  loading?: boolean;
  type?: 'button' | 'submit';
  class?: string;
}
const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
  size: 'default',
  type: 'button',
});
</script>

<template>
  <button
    :type="props.type"
    :disabled="props.disabled || props.loading"
    :class="cn(
      'inline-flex items-center justify-center rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50',
      props.size === 'sm' ? 'h-8 px-3 text-xs' : 'h-9 px-4 py-2 text-sm',
      props.variant === 'default' && 'bg-primary text-primary-foreground hover:bg-primary/90',
      props.variant === 'ghost' && 'bg-muted/30 hover:bg-muted hover:text-foreground',
      props.variant === 'danger' && 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
      props.variant === 'secondary' && 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
      props.class,
    )"
  >
    <span v-if="props.loading" class="mr-2 h-3.5 w-3.5 animate-spin rounded-full border-2 border-current border-t-transparent" />
    <slot />
  </button>
</template>
