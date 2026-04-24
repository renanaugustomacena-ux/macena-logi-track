import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';

// Lazy-loaded route components to keep the initial bundle small.
const HomeView = () => import('@/views/HomeView.vue');
const ShipmentView = () => import('@/views/ShipmentView.vue');

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'home', component: HomeView, meta: { title: 'Spedizioni' } },
  {
    path: '/shipments/:id',
    name: 'shipment-detail',
    component: ShipmentView,
    props: true,
    meta: { title: 'Dettaglio spedizione' },
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
});

router.beforeEach((to, _from, next) => {
  const base = 'LogiTrack';
  document.title = to.meta.title ? `${to.meta.title} - ${base}` : base;
  next();
});

export default router;
