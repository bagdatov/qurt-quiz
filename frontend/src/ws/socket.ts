// Singleton WebSocket connection shared across the app.
// The store calls connect() once on startup; components call send().

type Listener = (type: string, payload: unknown) => void

let socket: WebSocket | null = null
const listeners: Listener[] = []

function getWsUrl() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  return `${proto}://${location.host}/ws`
}

export function connect(onOpen?: () => void): void {
  if (socket && socket.readyState <= WebSocket.OPEN) return

  socket = new WebSocket(getWsUrl())

  socket.onopen = () => onOpen?.()

  socket.onmessage = (e: MessageEvent) => {
    try {
      const { type, payload } = JSON.parse(e.data as string)
      listeners.forEach(l => l(type, payload))
    } catch {
      // malformed message — ignore
    }
  }

  socket.onclose = () => {
    socket = null
    // Attempt reconnect after 2 seconds.
    setTimeout(() => connect(onOpen), 2000)
  }
}

export function send(type: string, payload: unknown): void {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    console.warn('ws not ready, dropping', type)
    return
  }
  socket.send(JSON.stringify({ type, payload }))
}

export function addListener(l: Listener): () => void {
  listeners.push(l)
  return () => {
    const idx = listeners.indexOf(l)
    if (idx !== -1) listeners.splice(idx, 1)
  }
}

export function isConnected(): boolean {
  return socket?.readyState === WebSocket.OPEN
}
