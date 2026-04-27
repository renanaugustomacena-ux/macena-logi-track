<template>
  <div class="min-h-screen flex flex-col bg-slate-50 text-slate-900">
    <header class="bg-logitrack-teal text-white shadow">
      <div class="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
        <router-link to="/" class="flex items-center gap-3">
          <span
            class="inline-block h-9 w-9 rounded bg-logitrack-blue flex items-center justify-center font-bold"
            aria-hidden="true"
          >LT</span>
          <span class="text-xl font-semibold tracking-tight">LogiTrack</span>
          <span class="text-xs uppercase tracking-widest text-slate-200/80 hidden md:inline">Supply-Chain Visibility</span>
        </router-link>
        <nav aria-label="Principale" class="flex gap-4 text-sm items-center">
          <router-link to="/" class="hover:underline">Spedizioni</router-link>
          <button
            v-if="authed"
            type="button"
            class="hover:underline text-slate-200/90"
            @click="onLogout"
          >Esci</button>
          <router-link
            v-else
            to="/login"
            class="hover:underline text-slate-200/90"
          >Accedi</router-link>
        </nav>
      </div>
    </header>

    <main class="flex-1 max-w-7xl w-full mx-auto px-6 py-8">
      <router-view />
    </main>

    <footer class="bg-slate-900 text-slate-300 text-sm">
      <div class="max-w-7xl mx-auto px-6 py-6 flex flex-col md:flex-row justify-between gap-2">
        <p>LogiTrack - Mozzecane (VR) - P.IVA 00000000000</p>
        <p>Made in Verona - Corridoio del Brennero</p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
// App shell: top bar, routed main area, footer. Keeps the markup
// accessible (semantic header/main/footer, named landmarks) and the
// brand palette consistent with the landing page.
//
// The previous nav links to /docs/API.md and /docs/ARCHITECTURE.md
// were removed — nginx serves dist/ only, those paths 404'd in
// production. Replaced with a token-aware Login / Logout entry so
// the user has an obvious way to clear session state.

import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { clearAccessToken, hasAccessToken } from '@/lib/tokenStore';

const route = useRoute();
const router = useRouter();

// route.fullPath is reactive, so this getter re-evaluates on
// navigation — no manual subscription needed.
const authed = computed(() => {
  void route.fullPath;
  return hasAccessToken();
});

function onLogout() {
  clearAccessToken();
  void router.replace({ name: 'login' });
}
</script>
