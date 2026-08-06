<template>
  <div class="metric-card glass-card" :class="themeClass">
    <div class="card-header">
      <span class="card-title">{{ title }}</span>
      <div class="icon-wrapper">
        <component :is="icon" v-if="icon" class="icon" />
      </div>
    </div>
    <div class="card-body">
      <div class="value">{{ count.toLocaleString() }}</div>
      <span class="subtext">{{ subtext }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  title: string
  count: number
  theme: 'pending' | 'processing' | 'delayed' | 'dlq'
  subtext: string
  icon?: any
}>()

const themeClass = computed(() => `theme-${props.theme}`)
</script>

<style scoped>
.metric-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  position: relative;
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.card-title {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.icon-wrapper {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.05);
}

.icon {
  width: 18px;
  height: 18px;
}

.value {
  font-size: 2.25rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  margin-bottom: 4px;
}

.subtext {
  font-size: 0.75rem;
  color: var(--text-muted);
}

/* Status Themes */
.theme-pending {
  border-left: 4px solid var(--status-pending);
}
.theme-pending .value { color: var(--status-pending); }
.theme-pending .icon-wrapper { color: var(--status-pending); background: rgba(245, 158, 11, 0.1); }

.theme-processing {
  border-left: 4px solid var(--status-processing);
}
.theme-processing .value { color: var(--status-processing); }
.theme-processing .icon-wrapper { color: var(--status-processing); background: rgba(6, 182, 212, 0.1); }

.theme-delayed {
  border-left: 4px solid var(--status-delayed);
}
.theme-delayed .value { color: var(--status-delayed); }
.theme-delayed .icon-wrapper { color: var(--status-delayed); background: rgba(139, 92, 246, 0.1); }

.theme-dlq {
  border-left: 4px solid var(--status-dlq);
}
.theme-dlq .value { color: var(--status-dlq); }
.theme-dlq .icon-wrapper { color: var(--status-dlq); background: rgba(239, 68, 68, 0.1); }
</style>
