<script setup lang="ts">
interface Props {
  class?: string;
  horizontal?: boolean;
}
const props = defineProps<Props>();
</script>

<template>
  <div :class="['overflow-auto', props.horizontal ? 'overflow-x-auto overflow-y-hidden' : 'overflow-y-auto overflow-x-hidden', props.class]">
    <slot />
  </div>
</template>
