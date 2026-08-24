<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRuntimeConfig } from "#app";
import {
  Server,
  Database,
  Shield,
  CheckCircle2,
  AlertCircle,
  ArrowUpRight,
} from "lucide-vue-next";

const config = useRuntimeConfig();

interface HealthStatus {
  status: string;
  service: string;
  environment: string;
  checks?: Record<string, string>;
}

const apiHealth = ref<HealthStatus | null>(null);
const verifierHealth = ref<HealthStatus | null>(null);
const apiError = ref<string | null>(null);
const verifierError = ref<string | null>(null);
const isChecking = ref(false);

const checkHealthServices = async () => {
  isChecking.value = true;
  apiError.value = null;
  verifierError.value = null;

  try {
    const res = await fetch(`${config.public.apiBase}/health`);
    if (res.ok) {
      apiHealth.value = await res.json();
    } else {
      apiError.value = `HTTP ${res.status}`;
    }
  } catch (err: any) {
    apiError.value = err.message || "Offline or CORS restricted";
  }

  try {
    const res = await fetch(`${config.public.verifierBase}/health`);
    if (res.ok) {
      verifierHealth.value = await res.json();
    } else {
      verifierError.value = `HTTP ${res.status}`;
    }
  } catch (err: any) {
    verifierError.value = err.message || "Offline or CORS restricted";
  }

  isChecking.value = false;
};

onMounted(() => {
  checkHealthServices();
});
</script>

<template>
  <div class="space-y-10">
    <!-- Hero / Status Banner -->
    <div
      class="p-8 rounded-xl bg-neutral-900/60 border border-neutral-800/80 space-y-4"
    >
      <div class="flex items-center gap-2">
        <span
          class="inline-block w-2.5 h-2.5 rounded-full bg-emerald-500"
        ></span>
        <span
          class="text-xs font-mono uppercase tracking-wider text-neutral-400"
          >Repository Scaffolding Booted</span
        >
      </div>
      <h1 class="text-3xl font-bold tracking-tight text-neutral-100">
        ClipIN Infrastructure Engine
      </h1>
      <p class="text-neutral-400 text-sm max-w-2xl leading-relaxed">
        Initial monorepo development scaffolding for ClipIN—India's performance
        clipping marketplace. No product features, fake statistics, or mock
        campaign data are enabled in this phase.
      </p>
    </div>

    <!-- Health Diagnostics -->
    <div class="space-y-4">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold tracking-tight text-neutral-200">
          Backend System Health
        </h2>
        <button
          @click="checkHealthServices"
          :disabled="isChecking"
          class="px-3 py-1.5 rounded-md bg-neutral-900 hover:bg-neutral-800 border border-neutral-800 text-xs font-mono transition-colors disabled:opacity-50"
        >
          {{ isChecking ? "Pinging..." : "Refresh Status" }}
        </button>
      </div>

      <div class="grid sm:grid-cols-2 gap-4">
        <!-- API Status Card -->
        <div
          class="p-5 rounded-lg bg-neutral-900/40 border border-neutral-800/80 space-y-3"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <Server class="w-4 h-4 text-neutral-400" />
              <span class="font-medium text-sm">Go Backend API</span>
            </div>
            <span
              v-if="apiHealth"
              class="px-2 py-0.5 rounded text-[11px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center gap-1"
            >
              <CheckCircle2 class="w-3 h-3" /> ONLINE
            </span>
            <span
              v-else
              class="px-2 py-0.5 rounded text-[11px] font-mono bg-amber-500/10 text-amber-400 border border-amber-500/20 flex items-center gap-1"
            >
              <AlertCircle class="w-3 h-3" /> PENDING / UNREACHABLE
            </span>
          </div>

          <div
            class="text-xs font-mono text-neutral-400 space-y-1 pt-1 border-t border-neutral-800/50"
          >
            <div>
              Endpoint:
              <code class="text-neutral-300"
                >{{ config.public.apiBase }}/health</code
              >
            </div>
            <div v-if="apiHealth?.checks">
              PostgreSQL:
              <span
                :class="
                  apiHealth.checks.postgres === 'up'
                    ? 'text-emerald-400'
                    : 'text-amber-400'
                "
                >{{ apiHealth.checks.postgres }}</span
              >
              | Redis:
              <span
                :class="
                  apiHealth.checks.redis === 'up'
                    ? 'text-emerald-400'
                    : 'text-amber-400'
                "
                >{{ apiHealth.checks.redis }}</span
              >
            </div>
            <div v-if="apiError" class="text-rose-400">{{ apiError }}</div>
          </div>
        </div>

        <!-- Verifier Status Card -->
        <div
          class="p-5 rounded-lg bg-neutral-900/40 border border-neutral-800/80 space-y-3"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <Shield class="w-4 h-4 text-neutral-400" />
              <span class="font-medium text-sm">Verifier Microservice</span>
            </div>
            <span
              v-if="verifierHealth"
              class="px-2 py-0.5 rounded text-[11px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center gap-1"
            >
              <CheckCircle2 class="w-3 h-3" /> ONLINE
            </span>
            <span
              v-else
              class="px-2 py-0.5 rounded text-[11px] font-mono bg-amber-500/10 text-amber-400 border border-amber-500/20 flex items-center gap-1"
            >
              <AlertCircle class="w-3 h-3" /> PENDING / UNREACHABLE
            </span>
          </div>

          <div
            class="text-xs font-mono text-neutral-400 space-y-1 pt-1 border-t border-neutral-800/50"
          >
            <div>
              Endpoint:
              <code class="text-neutral-300"
                >{{ config.public.verifierBase }}/health</code
              >
            </div>
            <div v-if="verifierHealth">
              Service: {{ verifierHealth.service }}
            </div>
            <div v-if="verifierError" class="text-rose-400">
              {{ verifierError }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Monorepo Stack Overview -->
    <div class="space-y-4">
      <h2 class="text-lg font-semibold tracking-tight text-neutral-200">
        Infrastructure Components
      </h2>

      <div class="grid sm:grid-cols-3 gap-4 text-xs font-mono">
        <div
          class="p-4 rounded-lg bg-neutral-900/30 border border-neutral-800/60 space-y-1.5"
        >
          <div class="font-semibold text-neutral-200">Frontend Shell</div>
          <div class="text-neutral-400">Nuxt 3 • Vue 3 • TypeScript</div>
          <div class="text-neutral-500">Tailwind • TanStack Query • Zod</div>
        </div>

        <div
          class="p-4 rounded-lg bg-neutral-900/30 border border-neutral-800/60 space-y-1.5"
        >
          <div class="font-semibold text-neutral-200">Go Backend API</div>
          <div class="text-neutral-400">net/http • Chi • Huma OpenAPI</div>
          <div class="text-neutral-500">pgxpool • go-redis client</div>
        </div>

        <div
          class="p-4 rounded-lg bg-neutral-900/30 border border-neutral-800/60 space-y-1.5"
        >
          <div class="font-semibold text-neutral-200">Verification Engine</div>
          <div class="text-neutral-400">Go Service Scaffold</div>
          <div class="text-neutral-500">Standalone Microservice</div>
        </div>
      </div>
    </div>
  </div>
</template>
