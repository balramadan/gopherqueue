import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface MetricsData {
  pending: number
  processing: number
  delayed: number
  dlq: number
  timestamp: number
}

export interface JobItem {
  id: string
  queue: string
  type: string
  payload: Record<string, any>
  status: string
  current_retry: number
  max_retries: number
  last_error?: string
  created_at: number
  updated_at: number
}

export type ConnectionStatus = 'LIVE' | 'RECONNECTING' | 'DISCONNECTED'

export const useTelemetryStore = defineStore('telemetry', () => {
  const metrics = ref<MetricsData>({
    pending: 0,
    processing: 0,
    delayed: 0,
    dlq: 0,
    timestamp: Date.now() / 1000,
  })

  const connectionStatus = ref<ConnectionStatus>('DISCONNECTED')
  const history = ref<MetricsData[]>([])
  const jobs = ref<JobItem[]>([])

  let eventSource: EventSource | null = null

  const totalActive = computed(() => metrics.value.pending + metrics.value.processing + metrics.value.delayed)

  function connectSSE() {
    if (eventSource) {
      eventSource.close()
    }

    connectionStatus.value = 'RECONNECTING'
    const isDev = import.meta.env.DEV
    const streamUrl = isDev ? 'http://localhost:8080/api/v1/stream' : '/api/v1/stream'

    eventSource = new EventSource(streamUrl)

    eventSource.onopen = () => {
      connectionStatus.value = 'LIVE'
    }

    eventSource.onmessage = (event) => {
      try {
        const data: MetricsData = JSON.parse(event.data)
        metrics.value = data
        connectionStatus.value = 'LIVE'

        // Maintain up to 30 historical samples
        history.value.push({ ...data })
        if (history.value.length > 30) {
          history.value.shift()
        }
      } catch (err) {
        console.error('[Telemetry Store] Failed to parse SSE JSON:', err)
      }
    }

    eventSource.onerror = (err) => {
      console.warn('[Telemetry Store] SSE Error:', err)
      connectionStatus.value = 'RECONNECTING'
    }
  }

  function disconnectSSE() {
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
    connectionStatus.value = 'DISCONNECTED'
  }

  return {
    metrics,
    connectionStatus,
    history,
    jobs,
    totalActive,
    connectSSE,
    disconnectSSE,
  }
})
