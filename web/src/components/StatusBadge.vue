<template>
  <span class="status-badge" :class="badgeClass">
    {{ status }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: string
}>()

const badgeClass = computed(() => {
  switch (props.status.toUpperCase()) {
    case 'PENDING':
    case 'ENQUEUED':
      return 'pending'
    case 'PROCESSING':
      return 'processing'
    case 'DELAYED':
    case 'RATE_LIMITED':
      return 'delayed'
    case 'COMPLETED':
      return 'completed'
    case 'FAILED':
    case 'DEAD_LETTER':
    case 'DLQ':
      return 'dlq'
    default:
      return 'default'
  }
})
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 6px;
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.pending {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.processing {
  background: rgba(6, 182, 212, 0.15);
  color: #06b6d4;
}

.delayed {
  background: rgba(139, 92, 246, 0.15);
  color: #8b5cf6;
}

.completed {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.dlq {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.default {
  background: rgba(148, 163, 184, 0.15);
  color: #94a3b8;
}
</style>
