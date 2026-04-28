import { createRouter, createWebHistory, type RouteLocationNormalized, type RouteRecordRaw } from 'vue-router';
import { hasAccessToken } from '@/lib/tokenStore';

// Lazy-loaded route components to keep the initial bundle small.
const HomeView = () => import('@/views/HomeView.vue');
const ShipmentView = () => import('@/views/ShipmentView.vue');
const LoginView = () => import('@/views/LoginView.vue');
const RifiutiView = () => import('@/views/RifiutiView.vue');

declare module 'vue-router' {
  // Re-export so `meta.requiresAuth` is type-checked.
  interface RouteMeta {
    title?: string;
    requiresAuth?: boolean;
    publicOnly?: boolean;
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    meta: { title: 'Accedi', publicOnly: true },
  },
  {
    path: '/',
    name: 'home',
    component: HomeView,
    meta: { title: 'Spedizioni', requiresAuth: true },
  },
  {
    path: '/shipments/:id',
    name: 'shipment-detail',
    component: ShipmentView,
    props: true,
    meta: { title: 'Dettaglio spedizione', requiresAuth: true },
  },
  {
    path: '/rifiuti',
    name: 'rifiuti',
    component: RifiutiView,
    meta: { title: 'Rifiuti speciali', requiresAuth: true },
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
});

// Auth guard. Routes flagged requiresAuth bounce to /login when no
// token is present; /login itself bounces to / when the user is
// already authenticated (publicOnly). The redirect query parameter
// preserves the original target so a successful login lands the
// user where they were headed.
router.beforeEach((to: RouteLocationNormalized) => {
  const base = 'LogiTrack';
  document.title = to.meta.title ? `${to.meta.title} - ${base}` : base;

  const authed = hasAccessToken();
  if (to.meta.requiresAuth && !authed) {
    return { name: 'login', query: { redirect: to.fullPath } };
  }
  if (to.meta.publicOnly && authed) {
    return { name: 'home' };
  }
  return true;
});

export default router;
