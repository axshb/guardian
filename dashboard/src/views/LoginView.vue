<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import Card from '@/components/ui/Card.vue';
import Button from '@/components/ui/Button.vue';

const router = useRouter();
const password = ref('');
const error = ref('');
const loading = ref(false);
const setupComplete = ref(true);

onMounted(async () => {
  try {
    const res = await fetch('/api/config');
    const cfg = await res.json();
    setupComplete.value = cfg.setupComplete;
    if (!cfg.setupComplete) {
      router.push('/setup');
    }
  } catch {
    setupComplete.value = false;
    router.push('/setup');
  }
});

async function handleSubmit() {
  error.value = '';
  loading.value = true;
  try {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: password.value }),
    });
    if (res.ok) {
      router.push('/admin/assets');
    } else {
      const data = await res.json();
      error.value = data.detail || 'Invalid password.';
    }
  } catch {
    error.value = 'Login failed.';
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div v-if="setupComplete" class="min-h-screen flex items-center justify-center bg-background p-4">
    <Card class="w-full max-w-sm p-6">
      <h1 class="text-lg font-semibold mb-1">guardian</h1>
      <p class="text-sm text-muted-foreground mb-4">Sign in to your instance.</p>
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
        <div>
          <input
            v-model="password"
            type="password"
            placeholder="Password"
            class="flex h-9 w-full rounded-md border border-border bg-background px-3 py-1 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            @keyup.enter="handleSubmit"
          />
        </div>
        <Button type="submit" class="w-full" :loading="loading">Unlock</Button>
      </form>
    </Card>
  </div>
</template>
