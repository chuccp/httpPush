<script setup lang="ts">
import { ref, onUnmounted } from 'vue'

const userId = ref('test_user')
const wsUrl = ref('ws://127.0.0.1:8084/ws')
const connected = ref(false)
const logs = ref<string[]>([])
let ws: WebSocket | null = null

function addLog(msg: string) {
  const t = new Date().toLocaleTimeString()
  logs.value.push(`[${t}] ${msg}`)
}

function connect() {
  if (connected.value) return
  const url = `${wsUrl.value}?id=${encodeURIComponent(userId.value)}`
  addLog(`WebSocket 连接: ${url}`)
  ws = new WebSocket(url)
  ws.onopen = () => {
    connected.value = true
    addLog('已连接')
  }
  ws.onmessage = (e) => {
    addLog(`收到: ${e.data}`)
  }
  ws.onclose = () => {
    connected.value = false
    addLog('连接关闭')
  }
  ws.onerror = (e) => {
    addLog('连接错误')
  }
}

function disconnect() {
  ws?.close()
  connected.value = false
}

function sendMsg() {
  const to = prompt('发送给谁?', 'user_b') ?? ''
  const msg = prompt('消息内容?', 'hello from ws') ?? ''
  if (!to || !msg) return
  const data = JSON.stringify({ to, msg })
  ws?.send(data)
  addLog(`发送: ${data}`)
}

onUnmounted(disconnect)
</script>

<template>
  <div>
    <h2>WebSocket 测试</h2>
    <div class="controls">
      <label>用户ID <input v-model="userId" /></label>
      <label>WS地址 <input v-model="wsUrl" style="width:200px" /></label>
      <button :class="connected ? 'btn-off' : 'btn-on'" @click="connected ? disconnect() : connect()">
        {{ connected ? '断开' : '连接' }}
      </button>
      <button v-if="connected" class="btn-send" @click="sendMsg">发送消息</button>
    </div>
    <div class="status" :class="{ on: connected }">
      {{ connected ? '● 已连接' : '○ 未连接' }}
    </div>
    <div class="log-box">
      <div v-for="(l, i) in logs" :key="i">{{ l }}</div>
      <div v-if="logs.length === 0" class="empty">点击"连接"开始</div>
    </div>
  </div>
</template>

<style scoped>
h2 { font-size: 18px; margin-bottom: 16px; color: #f1f5f9; }
.controls { display: flex; gap: 12px; align-items: center; margin-bottom: 12px; flex-wrap: wrap; }
.controls label { font-size: 13px; color: #94a3b8; }
.controls input {
  margin-left: 6px; padding: 4px 8px; border: 1px solid #334155; border-radius: 4px;
  background: #0f172a; color: #e2e8f0; font-size: 13px; width: 120px;
}
.status { font-size: 13px; padding: 4px 0 12px; }
.status.on { color: #86efac; }
button {
  padding: 6px 16px; border: none; border-radius: 4px; font-size: 13px; cursor: pointer;
}
.btn-on { background: #166534; color: #86efac; }
.btn-off { background: #7f1d1d; color: #fca5a5; }
.btn-send { background: #1e3a5f; color: #7dd3fc; }
.log-box {
  background: #0f172a; border: 1px solid #334155; border-radius: 6px;
  padding: 12px; max-height: 400px; overflow-y: auto; font-family: monospace; font-size: 13px;
}
.log-box div { padding: 2px 0; word-break: break-all; }
.empty { color: #475569; }
</style>
