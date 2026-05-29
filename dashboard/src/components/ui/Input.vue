<script setup lang="ts">
const model = defineModel<string>();
interface Props {
  placeholder?: string;
  disabled?: boolean;
  class?: string;
}
const props = defineProps<Props>();
</script>

<template>
  <input
    v-model="model"
    :placeholder="props.placeholder"
    :disabled="props.disabled"
    class="flex h-9 w-full rounded-md border border-border bg-background px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
    :class="props.class"
  />
</template>
