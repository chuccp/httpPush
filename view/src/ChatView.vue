<script setup lang="ts">
import { ref, nextTick, watch } from 'vue'

const props = defineProps<{ host: string; tab: 'ws' | 'http' }>()

const userId = ref('user_a')
const targetId = ref('user_b')
const groupId = ref('111')
const connected = ref(false)
const messages = ref<{ from: string; body: string; time: string; self: boolean }[]>([])

let ws: WebSocket | null = null
let pollCtrl: AbortController | null = null

function addMsg(from: string, body: string, self: boolean) {
  messages.value.push({ from, body, time: new Date().toLocaleTimeString(), self })
  nextTick(() => { const el = document.getElementById('chatList'); if (el) el.scrollTop = el.scrollHeight })
}

function gid() { return groupId.value ? `&groupId=${groupId.value}` : '' }

async function connect() {
  if (connected.value) return
  messages.value = []
  if (props.tab === 'ws') {
    ws = new WebSocket(`ws://${props.host}/ws?id=${userId.value}${gid()}`)
    ws.onopen = () => { connected.value = true; addMsg('system', '已连接 (WS)', false) }
    ws.onmessage = (e) => {
      try { JSON.parse(e.data).forEach((m: any) => addMsg(m.from || m.From, m.body || m.Body || m.msg || m.Msg, false)) } catch { /* */ }
    }
    ws.onclose = () => { connected.value = false; addMsg('system', '已断开', false) }
  } else {
    connected.value = true; addMsg('system', '已连接 (HTTP)', false); poll()
  }
}

async function poll() {
  while (connected.value && props.tab === 'http') {
    pollCtrl = new AbortController()
    try {
      const r = await fetch(`http://${props.host}/ex?id=${userId.value}&liveTime=15${gid()}`, { signal: pollCtrl.signal })
      const t = await r.text()
      if (t && t !== '[]') {
        try { JSON.parse(t).forEach((m: any) => addMsg(m.from || m.From, m.body || m.Body || m.msg || m.Msg, false)) } catch { /* */ }
      }
    } catch (e: any) { if (e.name !== 'AbortError') { connected.value = false; addMsg('system', '连接错误', false); break } }
  }
}

function disconnect() { ws?.close(); pollCtrl?.abort(); connected.value = false; ws = null }

function send() {
  const el = document.getElementById('msgInput') as HTMLInputElement
  const body = el?.value?.trim(); if (!body) return
  el.value = ''; addMsg(userId.value, body, true)
  if (props.tab === 'ws' && ws) { ws.send(JSON.stringify({ to: targetId.value, msg: body })) }
  else { fetch(`http://${props.host}/sendmsg?username=${targetId.value}&msg=${encodeURIComponent(body)}`) }
}

watch(() => props.tab, () => { if (connected.value) { disconnect(); connect() } })
</script>

<template>
  <div class="chat-panel">
    <div class="chat-topbar">
      <span class="proto-tag">{{ tab === 'ws' ? 'WebSocket' : 'HTTP 轮询' }}</span>
      <input v-model="userId" placeholder="你的 ID" class="sm" />
      <span class="arr">→</span>
      <input v-model="targetId" placeholder="目标 ID" class="sm" />
      <input v-model="groupId" placeholder="群组(可选)" class="sm" style="width:100px" />
      <button v-if="!connected" class="btn-on" @click="connect">连接</button>
      <button v-else class="btn-off" @click="disconnect">断开</button>
      <span class="dot" :class="{ on: connected }">{{ connected ? '●' : '○' }}</span>
    </div>
    <div id="chatList" class="msg-area">
      <div v-for="(m, i) in messages" :key="i" :class="['row', m.self ? 'me' : 'you']">
        <div class="bubble">
          <div class="who">{{ m.self ? '你' : m.from }}</div>
          <div class="text">{{ m.body }}</div>
          <div class="ts">{{ m.time }}</div>
        </div>
      </div>
      <div v-if="messages.length === 0" class="hint">连接后收发消息</div>
    </div>
    <div class="input-row">
      <input id="msgInput" placeholder="输入消息，回车发送..." @keyup.enter="send" :disabled="!connected" />
      <button @click="send" :disabled="!connected">发送</button>
    </div>
  </div>
</template>

<style scoped>
.chat-panel { flex:1; display:flex; flex-direction:column; }
.chat-topbar { display:flex; align-items:center; gap:8px; padding:10px 16px; background:#1e293b; border-bottom:1px solid #334155; flex-wrap:wrap; }
.chat-topbar input.sm { width:100px; padding:6px 8px; border:1px solid #334155; border-radius:6px; background:#0f172a; color:#e2e8f0; font-size:12px; outline:none; }
.chat-topbar input.sm:focus { border-color:#1e3a5f; }
.proto-tag { font-size:11px; font-weight:600; padding:4px 8px; border-radius:4px; background:#1e3a5f; color:#38bdf8; }
.arr { color:#475569; font-size:12px; }
.btn-on { padding:6px 14px; border:none; border-radius:6px; background:#166534; color:#86efac; font-size:12px; cursor:pointer; }
.btn-off { padding:6px 14px; border:none; border-radius:6px; background:#7f1d1d; color:#fca5a5; font-size:12px; cursor:pointer; }
.dot { font-size:10px; color:#475569; }
.dot.on { color:#86efac; }
.msg-area { flex:1; overflow-y:auto; padding:16px; display:flex; flex-direction:column; gap:10px; }
.hint { color:#475569; text-align:center; margin-top:40px; }
.row { display:flex; max-width:75%; }
.row.me { align-self:flex-end; }
.row.you { align-self:flex-start; }
.bubble { padding:10px 14px; border-radius:12px; font-size:13px; min-width:60px; }
.row.me .bubble { background:#1e3a5f; border-bottom-right-radius:4px; }
.row.you .bubble { background:#1e293b; border:1px solid #334155; border-bottom-left-radius:4px; }
.who { font-size:10px; color:#64748b; margin-bottom:4px; }
.text { word-break:break-word; }
.ts { font-size:9px; color:#475569; margin-top:4px; text-align:right; }
.input-row { display:flex; gap:8px; padding:12px 16px; border-top:1px solid #334155; background:#1e293b; }
.input-row input { flex:1; padding:10px 14px; border:1px solid #334155; border-radius:8px; background:#0f172a; color:#e2e8f0; font-size:14px; outline:none; }
.input-row input:focus { border-color:#1e3a5f; }
.input-row input:disabled { opacity:.4; }
.input-row button { padding:10px 24px; border:none; border-radius:8px; background:#1e3a5f; color:#38bdf8; font-size:14px; cursor:pointer; }
.input-row button:disabled { opacity:.4; }
</style>
