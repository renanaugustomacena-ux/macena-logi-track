import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';
import { UNAUTHORIZED_EVENT } from './api/client';
import './styles.css';

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
