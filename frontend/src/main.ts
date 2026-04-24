import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';
import './styles.css';

// Bootstrap the LogiTrack SPA. Order matters: Pinia before Router so
// route-guards can access stores; Router before mount so the initial
// navigation resolves before the first paint.
const app = createApp(App);
app.use(createPinia());
app.use(router);
app.mount('#app');
