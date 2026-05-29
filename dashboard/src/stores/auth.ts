import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useAuthStore = defineStore('auth', () => {
  const isAdmin = ref(false);

  async function checkAuth() {
    try {
      const res = await fetch('/api/auth/me');
      const data = await res.json();
      isAdmin.value = data.isAdmin;
    } catch {
      isAdmin.value = false;
    }
  }

  async function logout() {
    await fetch('/api/auth/logout', { method: 'POST' });
    isAdmin.value = false;
  }

  return { isAdmin, checkAuth, logout };
});
