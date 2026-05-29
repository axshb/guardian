import { createRouter, createWebHistory } from 'vue-router';

const routes = [
  { path: '/', component: () => import('./views/LoginView.vue') },
  { path: '/setup', component: () => import('./views/SetupView.vue') },
  {
    path: '/',
    component: () => import('./components/AdminLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: '/assets' },
      { path: 'assets', component: () => import('./views/AssetsView.vue') },
      { path: 'assets/:album', component: () => import('./views/AlbumView.vue') },
      { path: 'referrers', component: () => import('./views/ReferrersView.vue') },
      { path: 'settings', component: () => import('./views/SettingsView.vue') },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(async (to: any) => {
  // Check setup
  if (to.path !== '/setup') {
    try {
      const res = await fetch('/api/config');
      const cfg = await res.json();
      if (!cfg.setupComplete) {
        return '/setup';
      }
    } catch {
      return '/setup';
    }
  }

  // Check auth for admin routes
  if (to.meta.requiresAuth) {
    try {
      const res = await fetch('/api/auth/me');
      const data = await res.json();
      if (!data.isAdmin) {
        return '/';
      }
    } catch {
      return '/';
    }
  }
});

export default router;
