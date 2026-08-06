<template>
  <div class="app-container">
    <Navbar />

    <main class="dashboard-content">
      <!-- Metrics Grid -->
      <section class="metrics-grid">
        <MetricCard
          title="Pending Jobs"
          :count="telemetryStore.metrics.pending"
          theme="pending"
          subtext="Ready for execution in FIFO queue"
          :icon="Clock"
        />
        <MetricCard
          title="Processing Jobs"
          :count="telemetryStore.metrics.processing"
          theme="processing"
          subtext="Currently running on worker pool"
          :icon="Activity"
        />
        <MetricCard
          title="Delayed Jobs"
          :count="telemetryStore.metrics.delayed"
          theme="delayed"
          subtext="Scheduled or rate-limit deferred"
          :icon="Calendar"
        />
        <MetricCard
          title="Dead Letter Queue"
          :count="telemetryStore.metrics.dlq"
          theme="dlq"
          subtext="Failed after max retry attempts"
          :icon="AlertTriangle"
        />
      </section>

      <!-- Active Overview Summary Bar -->
      <section class="overview-bar glass-card">
        <div class="summary-item">
          <span class="label">Total Active Workload</span>
          <span class="val">{{ telemetryStore.totalActive }} Jobs</span>
        </div>
        <div class="summary-item">
          <span class="label">Primary Storage Engine</span>
          <span class="val accent">Redis (gopherq:queue)</span>
        </div>
        <div class="summary-item">
          <span class="label">Rate Limiter Synergy</span>
          <span class="val accent">distlimit (Graceful Defer)</span>
        </div>
      </section>

      <!-- Job Explorer Table -->
      <JobExplorerTable
        :jobs="telemetryStore.jobs"
        @select-job="selectedJob = $event"
      />
    </main>

    <!-- Job Detail Modal -->
    <JobDetailModal
      :job="selectedJob"
      @close="selectedJob = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Clock, Activity, Calendar, AlertTriangle } from '@lucide/vue'
import Navbar from './components/Navbar.vue'
import MetricCard from './components/MetricCard.vue'
import JobExplorerTable from './components/JobExplorerTable.vue'
import JobDetailModal from './components/JobDetailModal.vue'
import { useTelemetryStore, type JobItem } from './stores/telemetry'

const telemetryStore = useTelemetryStore()
const selectedJob = ref<JobItem | null>(null)

onMounted(() => {
  telemetryStore.connectSSE()
})

onUnmounted(() => {
  telemetryStore.disconnectSSE()
})
</script>

<style scoped>
.app-container {
  max-width: 1280px;
  margin: 0 auto;
  padding: 24px;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

@media (min-width: 640px) {
  .metrics-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .metrics-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

.overview-bar {
  padding: 16px 24px;
  display: flex;
  justify-content: space-around;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.summary-item .label {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.summary-item .val {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text-primary);
}

.summary-item .val.accent {
  color: var(--accent-brand);
}
</style>
