<template>
  <header class="navbar glass-card">
    <div class="brand">
      <div class="logo-icon">GQ</div>
      <div class="brand-text">
        <h1 class="title">GopherQueue</h1>
        <span class="version-tag">v0.1</span>
      </div>
    </div>

    <div class="controls">
      <ConnectionPill :status="telemetryStore.connectionStatus" />
      <button class="reconnect-btn" @click="telemetryStore.connectSSE">Refresh SSE</button>
    </div>
  </header>
</template>

<script setup lang="ts">
import ConnectionPill from './ConnectionPill.vue'
import { useTelemetryStore } from '../stores/telemetry'

const telemetryStore = useTelemetryStore()
</script>

<style scoped>
.navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  margin-bottom: 24px;
  border-radius: 12px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 14px;
}

.logo-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff;
  font-weight: 800;
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 16px rgba(99, 102, 241, 0.4);
}

.title {
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: #f8fafc;
}

.version-tag {
  font-size: 0.75rem;
  color: #94a3b8;
  font-weight: 500;
}

.controls {
  display: flex;
  align-items: center;
  gap: 16px;
}

.reconnect-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #cbd5e1;
  padding: 6px 14px;
  border-radius: 8px;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.reconnect-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}
</style>
