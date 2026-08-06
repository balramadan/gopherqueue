<template>
  <div class="metrics-grid">
    <div class="card pending">
      <h3>Pending Jobs</h3>
      <p class="value">{{ metrics.pending }}</p>
    </div>
    <div class="card processing">
      <h3>Processing Jobs</h3>
      <p class="value">{{ metrics.processing }}</p>
    </div>
    <div class="card delayed">
      <h3>Delayed Jobs</h3>
      <p class="value">{{ metrics.delayed }}</p>
    </div>
    <div class="card dlq">
      <h3>Dead Letter Queue (DLQ)</h3>
      <p class="value">{{ metrics.dlq }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const metrics = ref({
  pending: 0,
  processing: 0,
  delayed: 0,
  dlq: 0,
  timestamp: 0,
})

let eventSource = null

onMounted(() => {
  // Connect to Go Backend SSE Endpoint
  const streamUrl = window.location.port ? '/api/v1/stream' : 'http://localhost:8080/api/v1/stream'
  eventSource = new EventSource(streamUrl)

  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      metrics.value = data
    } catch (err) {
      console.error('Failed to parse SSE JSON data:', err)
    }
  }

  eventSource.onerror = (error) => {
    console.error('SSE Connection Error:', error)
  }
})

onUnmounted(() => {
  if (eventSource) {
    eventSource.close()
  }
})
</script>

<style scoped>
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.card {
  padding: 20px;
  border-radius: 8px;
  background: #1e1e2e;
  color: #fff;
}
.value {
  font-size: 2rem;
  font-weight: bold;
}
</style>
