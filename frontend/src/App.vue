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
        <nav aria-label="Principale" class="flex gap-4 text-sm">
          <router-link to="/" class="hover:underline">Spedizioni</router-link>
          <a href="/docs/API.md" class="hover:underline hidden md:inline">API</a>
          <a href="/docs/ARCHITECTURE.md" class="hover:underline hidden md:inline">Architettura</a>
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
</script>
