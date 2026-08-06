<template>
  <div class="connection-pill" :class="statusClass">
    <span class="dot"></span>
    <span class="label">{{ statusText }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ConnectionStatus } from '../stores/telemetry'

const props = defineProps<{
  status: ConnectionStatus
}>()

const statusClass = computed(() => {
  switch (props.status) {
    case 'LIVE':
      return 'live'
    case 'RECONNECTING':
      return 'reconnecting'
    case 'DISCONNECTED':
    default:
      return 'disconnected'
  }
})

const statusText = computed(() => {
  switch (props.status) {
    case 'LIVE':
      return 'LIVE SSE STREAM'
    case 'RECONNECTING':
      return 'RECONNECTING...'
    case 'DISCONNECTED':
    default:
      return 'OFFLINE'
  }
})
</script>

<style scoped>
.connection-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  border: 1px solid transparent;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.live {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
  border-color: rgba(16, 185, 129, 0.2);
}
.live .dot {
  background: #10b981;
  box-shadow: 0 0 8px #10b981;
  animation: pulse 2s infinite;
}

.reconnecting {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
  border-color: rgba(245, 158, 11, 0.2);
}
.reconnecting .dot {
  background: #f59e0b;
  animation: blink 1s infinite;
}

.disconnected {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.2);
}
.disconnected .dot {
  background: #ef4444;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(1.1); }
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.2; }
}
</style>
