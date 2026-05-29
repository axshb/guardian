<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue';
import Sortable from 'sortablejs';

interface Props {
  modelValue: any[];
  itemKey: string;
  tag?: string;
  class?: string;
}
const props = withDefaults(defineProps<Props>(), { tag: 'div' });
const emit = defineEmits<{ 'update:modelValue': [value: any[]]; end: [] }>();

const containerRef = ref<HTMLElement | null>(null);
let sortableInstance: Sortable | null = null;

onMounted(() => {
  if (!containerRef.value) return;
  sortableInstance = new Sortable(containerRef.value, {
    animation: 150,
    onEnd(evt) {
      const newList = [...props.modelValue];
      const [moved] = newList.splice(evt.oldIndex!, 1);
      newList.splice(evt.newIndex!, 0, moved);
      emit('update:modelValue', newList);
      emit('end');
    },
  });
});

onUnmounted(() => {
  sortableInstance?.destroy();
});
</script>

<template>
  <component :is="props.tag" ref="containerRef" :class="props.class">
    <slot />
  </component>
</template>
