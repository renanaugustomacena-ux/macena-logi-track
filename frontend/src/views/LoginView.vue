<template>
  <section class="max-w-md mx-auto bg-white rounded-lg shadow-sm border border-slate-200 p-6 mt-12">
    <h1 class="text-xl font-semibold text-slate-800 mb-1">Accedi a LogiTrack</h1>
    <p class="text-sm text-slate-500 mb-6">
      Inserisci le credenziali del tuo tenant. Le credenziali demo sono
      configurate nelle variabili d'ambiente <code>LOGITRACK_IDENTITY_DEMO_*</code>
      del backend.
    </p>

    <form class="space-y-4" @submit.prevent="onSubmit">
      <div>
        <label for="login-username" class="block text-xs uppercase tracking-wide text-slate-500 mb-1">Email</label>
        <input
          id="login-username"
          v-model="username"
          type="email"
          autocomplete="username"
          required
          :disabled="loading"
          class="w-full border border-slate-200 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-logitrack-blue"
        />
      </div>
      <div>
        <label for="login-password" class="block text-xs uppercase tracking-wide text-slate-500 mb-1">Password</label>
        <input
          id="login-password"
          v-model="password"
          type="password"
          autocomplete="current-password"
          required
          :disabled="loading"
          class="w-full border border-slate-200 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-logitrack-blue"
        />
      </div>

      <p v-if="errorMessage" class="text-sm text-red-700 bg-red-50 border border-red-200 rounded px-3 py-2" role="alert">
        {{ errorMessage }}
      </p>

      <button
        type="submit"
        :disabled="loading || !username || !password"
        class="w-full bg-logitrack-blue text-white text-sm font-medium rounded px-4 py-2 hover:bg-blue-700 disabled:opacity-60 disabled:cursor-not-allowed"
      >
        {{ loading ? 'Accesso in corso...' : 'Accedi' }}
      </button>
    </form>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ApiClient } from '@/api/client';
import { setAccessToken } from '@/lib/tokenStore';

interface LoginResponse {
  tokenType: string;
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  issuedAt: number;
}

const username = ref('');
const password = ref('');
const loading = ref(false);
const errorMessage = ref<string | null>(null);

const router = useRouter();
const route = useRoute();

// Use a dedicated client without the global onUnauthorized hook so a
// failed login on /login does not redirect back to /login in a loop.
const loginClient = new ApiClient({ onUnauthorized: () => {} });

async function onSubmit() {
  errorMessage.value = null;
  loading.value = true;
  try {
    const res = await loginClient.post<LoginResponse, { username: string; password: string }>(
      '/auth/login',
      { username: username.value, password: password.value },
    );
    setAccessToken(res.accessToken);
    const redirect = (route.query.redirect as string | undefined) ?? '/';
    await router.replace(redirect);
  } catch (err) {
    const e = err as { status?: number; detail?: string };
    if (e.status === 401) {
      errorMessage.value = 'Credenziali non valide.';
    } else if (e.status === 503) {
      errorMessage.value = 'Backend di identità non configurato. Contatta l\'amministratore.';
    } else {
      errorMessage.value = e.detail ?? 'Errore di rete. Riprova.';
    }
  } finally {
    loading.value = false;
  }
}
</script>
