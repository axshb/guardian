<script setup lang="ts">
const model = defineModel<string>();
interface Props {
  placeholder?: string;
  disabled?: boolean;
  class?: string;
  rows?: number;
}
const props = withDefaults(defineProps<Props>(), { rows: 3 });
</script>

<template>
  <textarea
    v-model="model"
    :placeholder="props.placeholder"
    :disabled="props.disabled"
    :rows="props.rows"
    class="flex min-h-[60px] w-full rounded-md border border-border bg-background px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 resize-none"
    :class="props.class"
  />
</template>
