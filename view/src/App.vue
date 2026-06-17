<script setup lang="ts">
import { ref } from 'vue'

type Mode = 'ws' | 'http'

const userId = ref('user_a')
const targetId = ref('user_b')
const mode = ref<Mode>('ws')
const connected = ref(false)
const messages = ref<{ from: string; body: string; time: string; self: boolean }[]>([])

let ws: WebSocket | null = null
let pollCtrl: AbortController | null = null

const host = '127.0.0.1:8084'

function addMsg(from: string, body: string, self: boolean) {
  messages.value.push({
    from,
    body,
    time: new Date().toLocaleTimeString(),
    self,
  })
}

async function connect() {
  if (connected.value) return
  if (mode.value === 'ws') {
    ws = new WebSocket(`ws://${host}/ws?id=${userId.value}`)
    ws.onopen = () => { connected.value = true; addMsg('system', 'WebSocket 已连接', false) }
    ws.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data)
        if (Array.isArray(data)) {
          data.forEach((m: any) => addMsg(m.from || m.From, m.body || m.Body || m.msg || m.Msg, false))
        }
      } catch { /* ignore */ }
    }
    ws.onclose = () => { connected.value = false; addMsg('system', '连接断开', false) }
  } else {
    connected.value = true
    addMsg('system', 'HTTP 轮询已连接', false)
    poll()
  }
}

async function poll() {
  while (connected.value && mode.value === 'http') {
    pollCtrl = new AbortController()
    try {
      const resp = await fetch(`http://${host}/ex?id=${userId.value}&liveTime=15`, { signal: pollCtrl.signal })
      const text = await resp.text()
      if (text && text !== '[]') {
        try {
          JSON.parse(text).forEach((m: any) => addMsg(m.from || m.From, m.body || m.Body || m.msg || m.Msg, false))
        } catch { /* ignore */ }
      }
    } catch (e: any) {
      if (e.name !== 'AbortError') { connected.value = false; addMsg('system', '连接错误', false); break }
    }
  }
}

function disconnect() {
  ws?.close()
  pollCtrl?.abort()
  connected.value = false
}

function send() {
  const body = (document.getElementById('msgInput') as HTMLInputElement)?.value?.trim()
  if (!body) return;
  (document.getElementById('msgInput') as HTMLInputElement).value = ''
  addMsg(userId.value, body, true)

  if (mode.value === 'ws' && ws) {
    ws.send(JSON.stringify({ to: targetId.value, msg: body }))
  } else {
    fetch(`http://${host}/sendmsg?username=${targetId.value}&msg=${encodeURIComponent(body)}`)
  }
}
</script>

<template>
  <div class="chat">
    <aside class="sidebar">
      <h2>httpPush</h2>
      <div class="field">
        <label>用户ID</label>
        <input v-model="userId" placeholder="你的ID" />
      </div>
      <div class="field">
        <label>发送给</label>
        <input v-model="targetId" placeholder="目标ID" />
      </div>
      <div class="field">
        <label>连接方式</label>
        <div class="tabs">
          <button :class="{ active: mode === 'ws' }" @click="mode = 'ws'">WebSocket</button>
          <button :class="{ active: mode === 'http' }" @click="mode = 'http'">HTTP 轮询</button>
        </div>
      </div>
      <div class="actions">
        <button v-if="!connected" class="btn-connect" @click="connect">连接</button>
        <button v-else class="btn-disconnect" @click="disconnect">断开</button>
      </div>
      <div class="status" :class="{ on: connected }">
        {{ connected ? '● 已连接 (' + mode.toUpperCase() + ')' : '○ 未连接' }}
      </div>
    </aside>

    <main class="main">
      <div class="msg-list" ref="listRef">
        <div v-for="(m, i) in messages" :key="i" :class="['msg', m.self ? 'self' : 'other']">
          <div class="bubble">
            <div class="sender">{{ m.self ? '你' : m.from }}</div>
            <div class="body">{{ m.body }}</div>
            <div class="time">{{ m.time }}</div>
          </div>
        </div>
        <div v-if="messages.length === 0" class="empty">连接后即可收发消息</div>
      </div>
      <div class="input-bar">
        <input id="msgInput" placeholder="输入消息..." @keyup.enter="send" :disabled="!connected" />
        <button @click="send" :disabled="!connected">发送</button>
      </div>
    </main>
  </div>
</template>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: system-ui, -apple-system, sans-serif; background: #0f172a; color: #e2e8f0; }
#app { height: 100vh; }
.chat { display: flex; height: 100vh; }
.sidebar {
  width: 260px; background: #1e293b; padding: 20px; display: flex; flex-direction: column; gap: 16px;
  border-right: 1px solid #334155; flex-shrink: 0;
}
.sidebar h2 { font-size: 18px; color: #38bdf8; }
.field { display: flex; flex-direction: column; gap: 4px; }
.field label { font-size: 11px; color: #64748b; text-transform: uppercase; letter-spacing: .5px; }
.field input {
  padding: 8px 10px; border: 1px solid #334155; border-radius: 6px;
  background: #0f172a; color: #e2e8f0; font-size: 13px; outline: none;
}
.field input:focus { border-color: #1e3a5f; }
.tabs { display: flex; gap: 4px; }
.tabs button {
  flex: 1; padding: 6px 0; border: 1px solid #334155; border-radius: 6px;
  background: #0f172a; color: #64748b; font-size: 12px; cursor: pointer;
}
.tabs button.active { background: #1e3a5f; color: #38bdf8; border-color: #1e3a5f; }
.btn-connect { padding: 10px; border: none; border-radius: 6px; background: #166534; color: #86efac; font-size: 14px; cursor: pointer; }
.btn-disconnect { padding: 10px; border: none; border-radius: 6px; background: #7f1d1d; color: #fca5a5; font-size: 14px; cursor: pointer; }
.status { font-size: 12px; color: #475569; }
.status.on { color: #86efac; }
.main { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.msg-list { flex: 1; overflow-y: auto; padding: 20px; display: flex; flex-direction: column; gap: 12px; }
.empty { color: #475569; text-align: center; margin-top: 40px; }
.msg { display: flex; max-width: 70%; }
.msg.self { align-self: flex-end; }
.msg.other { align-self: flex-start; }
.bubble { padding: 10px 14px; border-radius: 12px; font-size: 14px; }
.msg.self .bubble { background: #1e3a5f; border-bottom-right-radius: 4px; }
.msg.other .bubble { background: #1e293b; border: 1px solid #334155; border-bottom-left-radius: 4px; }
.sender { font-size: 11px; color: #64748b; margin-bottom: 4px; }
.body { color: #e2e8f0; word-break: break-word; }
.time { font-size: 10px; color: #475569; margin-top: 4px; text-align: right; }
.input-bar {
  display: flex; gap: 8px; padding: 16px 20px; border-top: 1px solid #334155; background: #1e293b;
}
.input-bar input {
  flex: 1; padding: 10px 14px; border: 1px solid #334155; border-radius: 8px;
  background: #0f172a; color: #e2e8f0; font-size: 14px; outline: none;
}
.input-bar input:focus { border-color: #1e3a5f; }
.input-bar input:disabled { opacity: .4; }
.input-bar button {
  padding: 10px 24px; border: none; border-radius: 8px; background: #1e3a5f; color: #38bdf8;
  font-size: 14px; cursor: pointer;
}
.input-bar button:disabled { opacity: .4; cursor: default; }
</style>
