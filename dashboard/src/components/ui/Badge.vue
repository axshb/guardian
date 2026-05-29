<script setup lang="ts">
import { cn } from '@/lib/utils';

interface Props {
  severity?: 'default' | 'success' | 'danger' | 'warning' | 'info' | 'secondary';
  class?: string;
}
withDefaults(defineProps<Props>(), { severity: 'default' });
</script>

<template>
  <span
    :class="cn(
      'inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium uppercase tracking-wide transition-colors',
      severity === 'default' && 'border-transparent bg-primary text-primary-foreground',
      severity === 'success' && 'border-transparent bg-success text-success-foreground',
      severity === 'danger' && 'border-transparent bg-destructive text-destructive-foreground',
      severity === 'warning' && 'border-transparent bg-warning text-warning-foreground',
      severity === 'info' && 'border-transparent bg-primary text-primary-foreground',
      severity === 'secondary' && 'border-transparent bg-muted text-muted-foreground',
      $attrs.class as string,
    )"
  >
    <slot />
  </span>
</template>
