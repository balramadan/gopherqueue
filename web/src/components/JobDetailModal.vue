<template>
  <div v-if="job" class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content glass-card">
      <div class="modal-header">
        <div>
          <span class="modal-subtitle">Job Details</span>
          <h2 class="modal-title font-mono">{{ job.id }}</h2>
        </div>
        <button class="close-btn" @click="$emit('close')">&times;</button>
      </div>

      <div class="modal-body">
        <div class="info-grid">
          <div class="info-item">
            <span class="label">Status</span>
            <StatusBadge :status="job.status" />
          </div>
          <div class="info-item">
            <span class="label">Job Type</span>
            <span class="val font-mono">{{ job.type }}</span>
          </div>
          <div class="info-item">
            <span class="label">Queue</span>
            <span class="val">{{ job.queue || 'default' }}</span>
          </div>
          <div class="info-item">
            <span class="label">Retries</span>
            <span class="val">{{ job.current_retry }} / {{ job.max_retries }}</span>
          </div>
        </div>

        <div v-if="job.last_error" class="error-box">
          <span class="error-title">Last Execution Error</span>
          <pre class="error-text">{{ job.last_error }}</pre>
        </div>

        <div class="payload-box">
          <span class="box-title">Raw Payload (JSON)</span>
          <pre class="json-code">{{ JSON.stringify(job.payload, null, 2) }}</pre>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" @click="$emit('close')">Close</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import StatusBadge from './StatusBadge.vue'
import type { JobItem } from '../stores/telemetry'

defineProps<{
  job: JobItem | null
}>()

defineEmits(['close'])
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.8);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal-content {
  width: 100%;
  max-width: 600px;
  background: #1e293b;
  border: 1px solid var(--border-glass-hover);
  padding: 24px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.modal-subtitle {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.modal-title {
  font-size: 1.2rem;
  color: var(--text-primary);
  margin-top: 2px;
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-secondary);
  font-size: 1.5rem;
  cursor: pointer;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.val {
  font-size: 0.9rem;
  color: var(--text-primary);
}

.error-box {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 20px;
}

.error-title {
  font-size: 0.75rem;
  font-weight: 700;
  color: #ef4444;
  display: block;
  margin-bottom: 6px;
}

.error-text {
  font-size: 0.8rem;
  color: #fca5a5;
  white-space: pre-wrap;
  word-break: break-all;
}

.payload-box {
  background: #0f172a;
  border: 1px solid var(--border-glass);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 20px;
}

.box-title {
  font-size: 0.75rem;
  color: var(--text-secondary);
  display: block;
  margin-bottom: 8px;
}

.json-code {
  font-size: 0.8rem;
  color: #a5b4fc;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
}

.btn {
  padding: 8px 16px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--border-glass);
  color: var(--text-primary);
}
</style>
