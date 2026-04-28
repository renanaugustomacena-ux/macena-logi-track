import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';
import { UNAUTHORIZED_EVENT } from './api/client';
import { setAccessToken } from './lib/tokenStore';
import './styles.css';

// Demo mode bootstrap. Activated only when the build was produced with
// VITE_DEMO_MODE=true (the GitHub Pages workflow does this). In that
// mode every /api/* fetch is intercepted by a static mock and the
// auth guard is satisfied with a fake token. In production builds,
// nothing in this branch executes.
if (import.meta.env.VITE_DEMO_MODE === 'true') {
  // Lazy import keeps the demo data out of the production bundle. The
  // top-level await is fine because main.ts is treated as an ES
  // module entry point by Vite.
  const { installDemoFetchMock } = await import('./demo/fetchInterceptor');
  installDemoFetchMock();
  setAccessToken('demo-token');
}

// Bootstrap the LogiTrack SPA. Order matters: Pinia before Router so
// route-guards can access stores; Router before mount so the initial
// navigation resolves before the first paint.
const app = createApp(App);
app.use(createPinia());
app.use(router);

// Centralised 401 handler: any in-flight API call that returns 401
// dispatches UNAUTHORIZED_EVENT (the api client clears the token
// before dispatching). We catch the event and push the user to
// /login so the redirect is immediate, not deferred to the next
// navigation. The redirect query param preserves the path the user
// was on so a refreshed login lands them back where they were.
window.addEventListener(UNAUTHORIZED_EVENT, () => {
  const current = router.currentRoute.value;
  if (current.name === 'login') return;
  void router.replace({ name: 'login', query: { redirect: current.fullPath } });
});

app.mount('#app');
