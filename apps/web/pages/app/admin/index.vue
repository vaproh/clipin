<script setup lang="ts">
import { Shield, Users, Flag, ScrollText, AlertTriangle, CheckCircle, XCircle, ChevronLeft, ChevronRight } from 'lucide-vue-next'

definePageMeta({
  layout: 'app',
  middleware: 'auth',
})

const { data: profile } = useUserQuery()
const isAdmin = computed(() => profile.value?.role === 'admin')

// Stats
const { data: stats, isLoading: statsLoading } = useAdminStats()

// Fraud flags
const { data: flagsData, isLoading: flagsLoading } = useAdminFraudFlags()
const resolveFlag = useResolveFraudFlag()
const dismissFlag = useDismissFraudFlag()
const resolvingId = ref<string | null>(null)
const resolutionText = ref('')

// Users
const usersPage = ref(1)
const { data: usersData, isLoading: usersLoading } = useAdminUsers(usersPage)
const totalUserPages = computed(() => {
  if (!usersData.value) return 0
  return Math.ceil(usersData.value.total / 20)
})

// Audit logs
const logsPage = ref(1)
const { data: logsData, isLoading: logsLoading } = useAdminAuditLogs(logsPage)
const totalLogPages = computed(() => {
  if (!logsData.value) return 0
  return Math.ceil(logsData.value.total / 20)
})

// Active section
type Section = 'flags' | 'users' | 'logs'
const activeSection = ref<Section>('flags')

// Sorted flags: open first, then by severity
const severityOrder: Record<string, number> = { critical: 0, high: 1, medium: 2, low: 3 }
const sortedFlags = computed(() => {
  if (!flagsData.value) return []
  return [...flagsData.value.flags].sort((a, b) => {
    const aOpen = a.status === 'open' || a.status === 'investigating' ? 0 : 1
    const bOpen = b.status === 'open' || b.status === 'investigating' ? 0 : 1
    if (aOpen !== bOpen) return aOpen - bOpen
    return (severityOrder[a.severity] ?? 4) - (severityOrder[b.severity] ?? 4)
  })
})

const severityBadge: Record<string, string> = {
  critical: 'bg-white text-black border-white',
  high: 'bg-neutral-200 text-black border-neutral-300',
  medium: 'bg-neutral-500 text-white border-neutral-500',
  low: 'bg-neutral-800 text-neutral-300 border-neutral-700',
}

function handleResolve(id: string) {
  if (!resolutionText.value.trim()) return
  resolveFlag.mutate({ id, resolution: resolutionText.value.trim() }, {
    onSuccess: () => {
      resolvingId.value = null
      resolutionText.value = ''
    },
  })
}

function handleDismiss(id: string) {
  dismissFlag.mutate(id)
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })
}

function formatDateTime(iso: string) {
  return new Date(iso).toLocaleString('en-IN', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div v-if="!isAdmin" class="space-y-5">
    <SharedPageHeader title="Access Denied" description="You do not have admin privileges." />
    <div class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
      <Shield class="w-8 h-8 text-neutral-600 mx-auto mb-3" />
      <p class="text-sm text-neutral-400">This page is restricted to admin users.</p>
    </div>
  </div>

  <div v-else class="space-y-5">
    <SharedPageHeader title="Admin Dashboard" description="Platform oversight and management" />

    <!-- Stats -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
      <SharedStatCard
        label="Total users"
        :value="statsLoading ? '...' : (stats?.total_users ?? 0)"
        :icon="Users"
      />
      <SharedStatCard
        label="Open fraud flags"
        :value="statsLoading ? '...' : (stats?.open_fraud_flags ?? 0)"
        :icon="Flag"
      />
      <SharedStatCard
        label="Total campaigns"
        :value="statsLoading ? '...' : (stats?.total_campaigns ?? 0)"
        :icon="ScrollText"
      />
    </div>

    <!-- Section tabs -->
    <div class="flex items-center gap-1 border-b border-neutral-900">
      <button
        v-for="tab in [{ key: 'flags' as Section, label: 'Fraud Flags', icon: Flag }, { key: 'users' as Section, label: 'Users', icon: Users }, { key: 'logs' as Section, label: 'Audit Logs', icon: ScrollText }]"
        :key="tab.key"
        class="flex items-center gap-1.5 px-3 py-2 text-xs font-mono transition-colors border-b-2 -mb-px"
        :class="activeSection === tab.key ? 'border-white text-white' : 'border-transparent text-neutral-500 hover:text-neutral-300'"
        @click="activeSection = tab.key"
      >
        <component :is="tab.icon" class="w-3.5 h-3.5" />
        {{ tab.label }}
      </button>
    </div>

    <!-- Fraud Flags -->
    <template v-if="activeSection === 'flags'">
      <SharedLoadingSpinner v-if="flagsLoading" />

      <div v-else-if="!sortedFlags.length" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
        <Flag class="w-8 h-8 text-neutral-600 mx-auto mb-3" />
        <p class="text-sm text-neutral-400">No fraud flags.</p>
      </div>

      <div v-else class="space-y-2">
        <div
          v-for="flag in sortedFlags"
          :key="flag.id"
          class="rounded bg-neutral-950 border border-neutral-800 p-4 space-y-3"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="space-y-1 min-w-0">
              <div class="flex items-center gap-2 flex-wrap">
                <span
                  class="inline-flex items-center rounded border px-2 py-0.5 text-[10px] font-mono uppercase tracking-wider"
                  :class="severityBadge[flag.severity]"
                >
                  {{ flag.severity }}
                </span>
                <span class="text-[10px] font-mono text-neutral-500">{{ flag.flag_type }}</span>
                <SharedStatusBadge :status="flag.status" />
              </div>
              <p v-if="flag.description" class="text-xs text-neutral-400">{{ flag.description }}</p>
              <div class="flex items-center gap-3 text-[11px] font-mono text-neutral-500">
                <span v-if="flag.user_id">User: {{ flag.user_id.slice(0, 8) }}...</span>
                <span v-if="flag.submission_id">Submission: {{ flag.submission_id.slice(0, 8) }}...</span>
                <span>{{ formatDate(flag.created_at) }}</span>
              </div>
            </div>

            <div v-if="flag.status === 'open' || flag.status === 'investigating'" class="flex items-center gap-1.5 shrink-0">
              <template v-if="resolvingId === flag.id">
                <UiInput
                  v-model="resolutionText"
                  placeholder="Resolution..."
                  class="h-9 w-full sm:w-40 text-[11px]"
                  @keyup.enter="handleResolve(flag.id)"
                />
                <UiButton size="xs" class="h-9" @click="handleResolve(flag.id)">Save</UiButton>
                <UiButton size="xs" variant="ghost" class="h-9" @click="resolvingId = null; resolutionText = ''">Cancel</UiButton>
              </template>
              <template v-else>
                <UiButton size="xs" variant="secondary" class="h-9 gap-1" @click="resolvingId = flag.id; resolutionText = ''">
                  <CheckCircle class="w-3 h-3" />
                  Resolve
                </UiButton>
                <UiButton size="xs" variant="ghost" class="h-9 gap-1" @click="handleDismiss(flag.id)">
                  <XCircle class="w-3 h-3" />
                  Dismiss
                </UiButton>
              </template>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Users -->
    <template v-if="activeSection === 'users'">
      <SharedLoadingSpinner v-if="usersLoading" />

      <div v-else-if="!usersData?.users.length" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
        <Users class="w-8 h-8 text-neutral-600 mx-auto mb-3" />
        <p class="text-sm text-neutral-400">No users found.</p>
      </div>

      <template v-else>
        <div class="rounded bg-neutral-950 border border-neutral-800 overflow-hidden">
          <div class="overflow-x-auto">
          <table class="w-full text-xs font-mono min-w-[480px]">
            <thead>
              <tr class="border-b border-neutral-800">
                <th class="text-left px-4 py-2.5 text-neutral-500 font-medium">User</th>
                <th class="text-left px-4 py-2.5 text-neutral-500 font-medium">Role</th>
                <th class="text-left px-4 py-2.5 text-neutral-500 font-medium hidden sm:table-cell">Joined</th>
                <th class="text-right px-4 py-2.5 text-neutral-500 font-medium">Flags</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="user in usersData.users"
                :key="user.id"
                class="border-b border-neutral-900 last:border-0 hover:bg-neutral-900/50 transition-colors"
              >
                <td class="px-4 py-2.5">
                  <div class="text-white truncate max-w-[200px]">{{ user.display_name || user.email }}</div>
                  <div class="text-neutral-500 truncate max-w-[200px]">{{ user.email }}</div>
                </td>
                <td class="px-4 py-2.5">
                  <SharedStatusBadge :status="user.role" />
                </td>
                <td class="px-4 py-2.5 text-neutral-400 hidden sm:table-cell">{{ formatDate(user.created_at) }}</td>
                <td class="px-4 py-2.5 text-right">
                  <span
                    v-if="user.flag_count > 0"
                    class="inline-flex items-center gap-1 text-neutral-300"
                  >
                    <AlertTriangle class="w-3 h-3" />
                    {{ user.flag_count }}
                  </span>
                  <span v-else class="text-neutral-600">0</span>
                </td>
              </tr>
            </tbody>
          </table>
          </div>
        </div>

        <!-- Pagination -->
        <div v-if="totalUserPages > 1" class="flex items-center justify-center gap-1 pt-2">
          <UiButton
            variant="ghost"
            size="sm"
            class="w-10 h-10 p-0"
            :disabled="usersPage <= 1"
            @click="usersPage--"
          >
            <ChevronLeft class="w-4 h-4" />
          </UiButton>
          <UiButton
            v-for="p in totalUserPages"
            :key="p"
            :variant="p === usersPage ? 'default' : 'ghost'"
            size="sm"
            class="w-10 h-10 p-0 font-mono text-xs"
            @click="usersPage = p"
          >
            {{ p }}
          </UiButton>
          <UiButton
            variant="ghost"
            size="sm"
            class="w-10 h-10 p-0"
            :disabled="usersPage >= totalUserPages"
            @click="usersPage++"
          >
            <ChevronRight class="w-4 h-4" />
          </UiButton>
        </div>
      </template>
    </template>

    <!-- Audit Logs -->
    <template v-if="activeSection === 'logs'">
      <SharedLoadingSpinner v-if="logsLoading" />

      <div v-else-if="!logsData?.logs.length" class="rounded bg-neutral-950 border border-neutral-800 p-8 text-center">
        <ScrollText class="w-8 h-8 text-neutral-600 mx-auto mb-3" />
        <p class="text-sm text-neutral-400">No audit logs.</p>
      </div>

      <template v-else>
        <div class="rounded bg-neutral-950 border border-neutral-800 overflow-hidden">
          <div class="overflow-x-auto">
          <table class="w-full text-xs font-mono min-w-[480px]">
            <thead>
              <tr class="border-b border-neutral-800">
                <th class="text-left px-4 py-2.5 text-neutral-500 font-medium">Action</th>
                <th class="text-left px-4 py-2.5 text-neutral-500 font-medium">Resource</th>
                <th class="text-left px-4 py-2.5 text-neutral-500 font-medium hidden sm:table-cell">Actor</th>
                <th class="text-left px-4 py-2.5 text-neutral-500 font-medium hidden md:table-cell">IP</th>
                <th class="text-right px-4 py-2.5 text-neutral-500 font-medium">Time</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="log in logsData.logs"
                :key="log.id"
                class="border-b border-neutral-900 last:border-0 hover:bg-neutral-900/50 transition-colors"
              >
                <td class="px-4 py-2.5">
                  <span class="text-white">{{ log.action }}</span>
                </td>
                <td class="px-4 py-2.5 text-neutral-400">
                  {{ log.resource_type }}:{{ log.resource_id.slice(0, 8) }}...
                </td>
                <td class="px-4 py-2.5 text-neutral-400 hidden sm:table-cell">
                  {{ log.actor_id.slice(0, 8) }}...
                </td>
                <td class="px-4 py-2.5 text-neutral-500 hidden md:table-cell">
                  {{ log.ip_address ?? '—' }}
                </td>
                <td class="px-4 py-2.5 text-neutral-400 text-right whitespace-nowrap">
                  {{ formatDateTime(log.created_at) }}
                </td>
              </tr>
            </tbody>
          </table>
          </div>
        </div>

        <!-- Pagination -->
        <div v-if="totalLogPages > 1" class="flex items-center justify-center gap-1 pt-2">
          <UiButton
            variant="ghost"
            size="sm"
            class="w-10 h-10 p-0"
            :disabled="logsPage <= 1"
            @click="logsPage--"
          >
            <ChevronLeft class="w-4 h-4" />
          </UiButton>
          <UiButton
            v-for="p in totalLogPages"
            :key="p"
            :variant="p === logsPage ? 'default' : 'ghost'"
            size="sm"
            class="w-10 h-10 p-0 font-mono text-xs"
            @click="logsPage = p"
          >
            {{ p }}
          </UiButton>
          <UiButton
            variant="ghost"
            size="sm"
            class="w-10 h-10 p-0"
            :disabled="logsPage >= totalLogPages"
            @click="logsPage++"
          >
            <ChevronRight class="w-4 h-4" />
          </UiButton>
        </div>
      </template>
    </template>
  </div>
</template>
