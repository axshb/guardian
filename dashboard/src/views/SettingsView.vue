<script setup lang="ts">
import { ref, onMounted } from 'vue';
import Card from '@/components/ui/Card.vue';
import { useToast } from '@/composables/useToast';
import { THEME_IDS, themeLabel, themePreview } from '@/themes/themes';

const toast = useToast();

const config = ref<any>(null);
const saving = ref(false);

onMounted(async () => {
  const res = await fetch('/api/config');
  config.value = await res.json();
});

async function selectTheme(theme: string) {
  if (theme === config.value?.theme || saving.value) return;
  saving.value = true;
  document.documentElement.dataset.theme = theme;
  try {
    await fetch('/api/config', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ theme }),
    });
    config.value.theme = theme;
    toast.add({ severity: 'success', summary: 'Theme updated', life: 3000 });
  } catch {
    toast.add({ severity: 'error', summary: 'Update failed', life: 3000 });
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <div class="space-y-6" v-if="config">
    <div>
      <h1 class="text-2xl font-semibold">Settings</h1>
      <p class="text-sm text-muted-foreground mt-1">Configure your guardian instance.</p>
    </div>

    <Card class="p-6">
      <h3 class="text-sm font-medium mb-1">Appearance</h3>
      <p class="text-sm text-muted-foreground mb-4">Choose a color theme for the dashboard.</p>
      <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
        <button
          v-for="id in THEME_IDS"
          :key="id"
          @click="selectTheme(id)"
          :disabled="saving"
          class="relative rounded-lg border p-4 text-left transition-colors"
          :class="config.theme === id ? 'border-primary bg-muted' : 'border-border hover:border-muted-foreground'"
        >
          <div class="text-sm font-medium">{{ themeLabel(id) }}</div>
          <div class="mt-2 flex gap-1">
            <span class="inline-block h-3 w-3 rounded-full" :style="{ background: themePreview(id).primary }" />
            <span class="inline-block h-3 w-3 rounded-full" :style="{ background: themePreview(id).bg }" />
            <span class="inline-block h-3 w-3 rounded-full" :style="{ background: themePreview(id).accent }" />
          </div>
        </button>
      </div>
    </Card>

    <Card class="p-6">
      <h3 class="text-sm font-medium mb-1">Instance</h3>
      <p class="text-sm text-muted-foreground mb-4">Current configuration values.</p>
      <div class="space-y-2 text-sm">
        <div class="flex justify-between">
          <span class="text-muted-foreground">Version</span>
          <span>guardian</span>
        </div>
      </div>
    </Card>
  </div>
</template>
