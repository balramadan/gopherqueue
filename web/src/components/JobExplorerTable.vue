<template>
  <div class="job-table-container glass-card">
    <div class="table-header">
      <div class="header-title">
        <h2>Job Explorer</h2>
        <span class="count-pill">{{ filteredJobs.length }} Jobs</span>
      </div>

      <div class="table-actions">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="Search Job ID or Type..."
          class="search-input"
        />

        <select v-model="selectedStatus" class="filter-select">
          <option value="ALL">All Statuses</option>
          <option value="PENDING">Pending</option>
          <option value="PROCESSING">Processing</option>
          <option value="DELAYED">Delayed</option>
          <option value="COMPLETED">Completed</option>
          <option value="DEAD_LETTER">DLQ</option>
        </select>
      </div>
    </div>

    <div class="table-responsive">
      <table class="data-table">
        <thead>
          <tr>
            <th>Job ID</th>
            <th>Queue</th>
            <th>Job Type</th>
            <th>Status</th>
            <th>Retry Count</th>
            <th>Last Updated</th>
            <th class="text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="filteredJobs.length === 0">
            <td colspan="7" class="empty-state">
              No matching jobs found in queue storage.
            </td>
          </tr>
          <tr v-for="job in filteredJobs" :key="job.id" class="table-row">
            <td class="font-mono text-bold">{{ job.id }}</td>
            <td><span class="queue-badge">{{ job.queue || 'default' }}</span></td>
            <td>{{ job.type }}</td>
            <td><StatusBadge :status="job.status" /></td>
            <td>{{ job.current_retry }} / {{ job.max_retries }}</td>
            <td class="text-muted">{{ formatDate(job.updated_at) }}</td>
            <td class="text-right">
              <button class="action-btn view-btn" @click="$emit('select-job', job)">
                Inspect
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import StatusBadge from './StatusBadge.vue'
import type { JobItem } from '../stores/telemetry'

const props = defineProps<{
  jobs: JobItem[]
}>()

defineEmits(['select-job'])

const searchQuery = ref('')
const selectedStatus = ref('ALL')

const filteredJobs = computed(() => {
  return props.jobs.filter((job) => {
    const matchesSearch =
      job.id.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      job.type.toLowerCase().includes(searchQuery.value.toLowerCase())

    const matchesStatus =
      selectedStatus.value === 'ALL' ||
      job.status.toUpperCase() === selectedStatus.value

    return matchesSearch && matchesStatus
  })
})

function formatDate(ts: number) {
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleTimeString()
}
</script>

<style scoped>
.job-table-container {
  padding: 20px;
  margin-top: 24px;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 12px;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-title h2 {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text-primary);
}

.count-pill {
  font-size: 0.75rem;
  background: rgba(99, 102, 241, 0.15);
  color: var(--accent-brand);
  padding: 2px 8px;
  border-radius: 9999px;
  font-weight: 600;
}

.table-actions {
  display: flex;
  gap: 10px;
}

.search-input, .filter-select {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid var(--border-glass);
  color: #fff;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 0.85rem;
  outline: none;
}

.search-input:focus, .filter-select:focus {
  border-color: var(--accent-brand);
}

.table-responsive {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}

.data-table th {
  text-align: left;
  padding: 12px 14px;
  color: var(--text-secondary);
  font-weight: 600;
  border-bottom: 1px solid var(--border-glass);
}

.data-table td {
  padding: 12px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}

.table-row:hover {
  background: rgba(255, 255, 255, 0.02);
}

.font-mono {
  font-family: 'JetBrains Mono', monospace;
}

.text-bold {
  font-weight: 600;
  color: var(--text-primary);
}

.text-muted {
  color: var(--text-muted);
}

.text-right {
  text-align: right;
}

.queue-badge {
  background: rgba(255, 255, 255, 0.05);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.empty-state {
  text-align: center;
  padding: 32px !important;
  color: var(--text-muted);
}

.action-btn {
  background: rgba(99, 102, 241, 0.1);
  color: var(--accent-brand);
  border: 1px solid rgba(99, 102, 241, 0.2);
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: var(--accent-brand);
  color: #fff;
}
</style>
