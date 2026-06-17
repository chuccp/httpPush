<script setup lang="ts">
import { ref, onUnmounted } from 'vue'

const userId = ref('test_user')
const liveTime = ref('15')
const connected = ref(false)
const logs = ref<string[]>([])
let controller: AbortController | null = null

function addLog(msg: string) {
  const t = new Date().toLocaleTimeString()
  logs.value.push(`[${t}] ${msg}`)
}

async function connect() {
  if (connected.value) return
  connected.value = true
  addLog(`HTTP 轮询连接中... userId=${userId.value}`)
  poll()
}

function disconnect() {
  connected.value = false
  controller?.abort()
  addLog('已断开')
}

async function poll() {
  while (connected.value) {
    controller = new AbortController()
    try {
      const url = `/ex?id=${encodeURIComponent(userId.value)}&liveTime=${liveTime.value}`
      const resp = await fetch(url, { signal: controller.signal })
      const text = await resp.text()
      if (text && text !== '[]') {
        addLog(`收到: ${text}`)
      }
    } catch (e: any) {
      if (e.name !== 'AbortError') {
        addLog(`连接错误: ${e.message}`)
        connected.value = false
      }
      break
    }
  }
}

async function sendMsg() {
  const to = prompt('发送给谁?', 'user_b')
  const msg = prompt('消息内容?', 'hello from ex test')
  if (!to || !msg) return
  try {
    const resp = await fetch(`/sendmsg?username=${encodeURIComponent(to)}&msg=${encodeURIComponent(msg)}`)
    addLog(`发送 → ${to}: ${await resp.text()}`)
  } catch (e: any) {
    addLog(`发送失败: ${e.message}`)
  }
}

onUnmounted(disconnect)
</script>

<template>
  <div>
    <h2>HTTP 长轮询测试</h2>
    <div class="controls">
      <label>用户ID <input v-model="userId" /></label>
      <label>liveTime <input v-model="liveTime" style="width:60px" /></label>
      <button :class="connected ? 'btn-off' : 'btn-on'" @click="connected ? disconnect() : connect()">
        {{ connected ? '断开' : '连接' }}
      </button>
      <button v-if="connected" class="btn-send" @click="sendMsg">发送消息</button>
    </div>
    <div class="log-box">
      <div v-for="(l, i) in logs" :key="i">{{ l }}</div>
      <div v-if="logs.length === 0" class="empty">点击"连接"开始</div>
    </div>
  </div>
</template>

<style scoped>
h2 { font-size: 18px; margin-bottom: 16px; color: #f1f5f9; }
.controls { display: flex; gap: 12px; align-items: center; margin-bottom: 16px; flex-wrap: wrap; }
.controls label { font-size: 13px; color: #94a3b8; }
.controls input {
  margin-left: 6px; padding: 4px 8px; border: 1px solid #334155; border-radius: 4px;
  background: #0f172a; color: #e2e8f0; font-size: 13px; width: 120px;
}
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
