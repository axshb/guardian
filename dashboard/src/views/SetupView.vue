<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import Card from '@/components/ui/Card.vue';
import Button from '@/components/ui/Button.vue';
import Input from '@/components/ui/Input.vue';
import { THEME_IDS, themeLabel, themePreview } from '@/themes/themes';

const router = useRouter();

const password = ref('');
const confirm = ref('');
const selectedTheme = ref('gruvbox');
const error = ref('');
const loading = ref(false);

onMounted(() => {
  document.documentElement.dataset.theme = selectedTheme.value;
});

function previewTheme(theme: string) {
  selectedTheme.value = theme;
  document.documentElement.dataset.theme = theme;
}

async function handleSubmit() {
  error.value = '';
  if (password.value !== confirm.value) {
    error.value = 'Passwords do not match.';
    return;
  }
  if (password.value.length < 8) {
    error.value = 'Password must be at least 8 characters.';
    return;
  }

  loading.value = true;
  try {
    const res = await fetch('/api/setup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        password: password.value,
        theme: selectedTheme.value,
      }),
    });
    if (res.ok) {
      router.push('/assets');
    } else {
      const data = await res.json();
      error.value = data.detail || 'Setup failed.';
    }
  } catch {
    error.value = 'Setup failed.';
  } finally {
    loading.value = false;
  }
}

</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-background p-4">
    <Card class="w-full max-w-md p-6">
      <h1 class="text-lg font-semibold mb-1">Welcome to guardian</h1>
      <p class="text-sm text-muted-foreground mb-5">Complete setup to get started.</p>
      <form @submit.prevent="handleSubmit" class="space-y-5">
        <div class="space-y-2">
          <label class="text-sm">Admin Password</label>
          <input v-model="password" type="password" placeholder="Choose a strong password" class="flex h-9 w-full rounded-md border border-border bg-background px-3 py-1 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" />
        </div>
        <div class="space-y-2">
          <label class="text-sm">Confirm Password</label>
          <input v-model="confirm" type="password" placeholder="Repeat password" class="flex h-9 w-full rounded-md border border-border bg-background px-3 py-1 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" />
        </div>
          <div class="space-y-2">
            <label class="text-sm">Theme</label>
            <div class="grid grid-cols-2 gap-2">
              <label
                v-for="id in THEME_IDS"
                :key="id"
                class="flex items-center gap-2 rounded-md border px-3 py-2 cursor-pointer hover:bg-muted transition-colors"
                :class="selectedTheme === id ? 'border-primary bg-muted' : 'border-border'"
              >
                <input
                  type="radio"
                  name="theme"
                  :value="id"
                  :checked="selectedTheme === id"
                  @change="previewTheme(id)"
                  class="accent-primary"
                />
                <span class="text-sm">{{ themeLabel(id) }}</span>
                <div class="ml-auto flex gap-1">
                  <span class="inline-block h-3 w-3 rounded-full" :style="{ background: themePreview(id).primary }" />
                  <span class="inline-block h-3 w-3 rounded-full" :style="{ background: themePreview(id).bg }" />
                  <span class="inline-block h-3 w-3 rounded-full" :style="{ background: themePreview(id).accent }" />
                </div>
              </label>
            </div>
          </div>
          <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
          <Button type="submit" class="w-full" :loading="loading">{{ loading ? 'Setting up...' : 'Complete Setup' }}</Button>
        </form>
    </Card>
  </div>
</template>
