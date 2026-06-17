<script setup lang="ts">
import { ref, nextTick, watch } from 'vue'

type Page = 'chat' | 'admin'

const page = ref<Page>('chat')
const userId = ref('user_a')
const targetId = ref('user_b')
const mode = ref<'ws' | 'http'>('ws')
const connected = ref(false)
const messages = ref<{ from: string; body: string; time: string; self: boolean }[]>([])
const host = '127.0.0.1:8084'

let ws: WebSocket | null = null
let pollCtrl: AbortController | null = null

function addMsg(from: string, body: string, self: boolean) {
  messages.value.push({ from, body, time: new Date().toLocaleTimeString(), self })
  nextTick(() => {
    const el = document.getElementById('chatList')
    if (el) el.scrollTop = el.scrollHeight
  })
}

async function connect() {
  if (connected.value) return
  messages.value = []
  if (mode.value === 'ws') {
    ws = new WebSocket(`ws://${host}/ws?id=${userId.value}`)
    ws.onopen = () => { connected.value = true; addMsg('system', `已连接 (${mode.value.toUpperCase()})`, false) }
    ws.onmessage = (e) => {
      try { JSON.parse(e.data).forEach((m: any) => addMsg(m.from || m.From, m.body || m.Body || m.msg || m.Msg, false)) } catch { /* */ }
    }
    ws.onclose = () => { connected.value = false; addMsg('system', '已断开', false) }
  } else {
    connected.value = true
    addMsg('system', `已连接 (HTTP 轮询)`, false)
    poll()
  }
}

async function poll() {
  while (connected.value && mode.value === 'http') {
    pollCtrl = new AbortController()
    try {
      const r = await fetch(`http://${host}/ex?id=${userId.value}&liveTime=15`, { signal: pollCtrl.signal })
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
  const body = el?.value?.trim()
  if (!body) return
  el.value = ''
  addMsg(userId.value, body, true)
  if (mode.value === 'ws' && ws) {
    ws.send(JSON.stringify({ to: targetId.value, msg: body }))
  } else {
    fetch(`http://${host}/sendmsg?username=${targetId.value}&msg=${encodeURIComponent(body)}`)
  }
}

// ---- admin ----
interface Endpoint { path: string; desc: string; params?: string[] }
const endpoints: Endpoint[] = [
  { path: '/root_version', desc: '版本信息' },
  { path: '/info_user', desc: '集群信息' },
  { path: '/onlineUser', desc: '在线用户', params: ['size'] },
  { path: '/queryUser', desc: '查询用户', params: ['id'] },
  { path: '/queryClusterUserNum', desc: '集群用户数' },
  { path: '/queryGroupInfo', desc: '群组信息' },
  { path: '/queryVersion', desc: '集群版本' },
]
const apiResult = ref('')
const apiLoading = ref('')
const paramValues = ref<Record<string, string>>({})
const activeEp = ref<Endpoint | null>(null)

function openForm(ep: Endpoint) {
  if (ep.params && ep.params.length > 0) {
    activeEp.value = ep
    paramValues.value = {}
  } else {
    callApi(ep, '')
  }
}

async function callApi(ep: Endpoint, qs: string) {
  activeEp.value = null
  apiLoading.value = ep.path
  const url = `http://${host}${ep.path}${qs}`
  try {
    const r = await fetch(url)
    apiResult.value = JSON.stringify(await r.json(), null, 2)
  } catch (e: any) {
    apiResult.value = `Error: ${e.message}`
  } finally { apiLoading.value = '' }
}

function submitForm() {
  if (!activeEp.value) return
  const qs = '?' + (activeEp.value.params || []).map(k => `${k}=${encodeURIComponent(paramValues.value[k] || '')}`).join('&')
  callApi(activeEp.value, qs)
}

watch(mode, () => { if (connected.value) { disconnect(); connect() } })
</script>

<template>
  <div class="layout">
    <!-- 左侧菜单 -->
    <nav class="menu">
      <div class="logo">httpPush</div>
      <button :class="{ sel: page === 'chat' }" @click="page = 'chat'">💬 聊天</button>
      <button :class="{ sel: page === 'admin' }" @click="page = 'admin'">⚙ 管理</button>
    </nav>

    <!-- 右侧内容 -->
    <div class="right">
      <!-- 聊天区域 -->
      <div v-show="page === 'chat'" class="chat-panel">
        <div class="chat-topbar">
          <input v-model="userId" placeholder="你的 ID" class="sm" />
          <span class="arr">→</span>
          <input v-model="targetId" placeholder="目标 ID" class="sm" />
          <div class="mode-switch">
            <button :class="{ a: mode === 'ws' }" @click="mode = 'ws'">WS</button>
            <button :class="{ a: mode === 'http' }" @click="mode = 'http'">HTTP</button>
          </div>
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

      <!-- 管理界面 -->
      <div v-show="page === 'admin'" class="admin-panel">
        <div class="api-grid">
          <button v-for="ep in endpoints" :key="ep.path" class="api-btn" @click="openForm(ep)" :disabled="apiLoading === ep.path">
            <span class="api-path">{{ ep.path }}</span>
            <span class="api-desc">{{ ep.desc }}</span>
          </button>
        </div>

        <!-- 参数表单 -->
        <div v-if="activeEp" class="param-form">
          <span class="param-title">{{ activeEp.path }}</span>
          <div v-for="p in activeEp.params" :key="p" class="param-row">
            <label>{{ p }}</label>
            <input v-model="paramValues[p]" :placeholder="p" @keyup.enter="submitForm" />
          </div>
          <div class="param-actions">
            <button class="btn-on" @click="submitForm">提交</button>
            <button class="btn-off" @click="activeEp = null">取消</button>
          </div>
        </div>

        <pre class="api-result" v-if="apiResult">{{ apiResult }}</pre>
        <div v-else-if="!activeEp" class="hint">点击接口查看结果</div>
      </div>
    </div>
  </div>
</template>

<style>
* { margin:0; padding:0; box-sizing:border-box; }
body { font-family:system-ui,-apple-system,sans-serif; background:#0f172a; color:#e2e8f0; }
#app { height:100vh; }
.layout { display:flex; height:100vh; }
/* menu */
.menu { width:180px; background:#1e293b; border-right:1px solid #334155; display:flex; flex-direction:column; padding:12px; gap:4px; flex-shrink:0; }
.logo { font-size:16px; font-weight:700; color:#38bdf8; padding:8px 12px 16px; }
.menu button { text-align:left; padding:10px 12px; border:none; border-radius:6px; background:transparent; color:#94a3b8; font-size:13px; cursor:pointer; }
.menu button:hover { background:#273548; color:#e2e8f0; }
.menu button.sel { background:#1e3a5f; color:#38bdf8; }
/* right */
.right { flex:1; display:flex; flex-direction:column; min-width:0; }
/* chat */
.chat-panel { flex:1; display:flex; flex-direction:column; }
.chat-topbar { display:flex; align-items:center; gap:8px; padding:10px 16px; background:#1e293b; border-bottom:1px solid #334155; }
.chat-topbar input.sm { width:100px; padding:6px 8px; border:1px solid #334155; border-radius:6px; background:#0f172a; color:#e2e8f0; font-size:12px; outline:none; }
.chat-topbar input.sm:focus { border-color:#1e3a5f; }
.arr { color:#475569; font-size:12px; }
.mode-switch { display:flex; border-radius:6px; overflow:hidden; border:1px solid #334155; }
.mode-switch button { padding:4px 10px; border:none; background:#0f172a; color:#64748b; font-size:11px; cursor:pointer; }
.mode-switch button.a { background:#1e3a5f; color:#38bdf8; }
.btn-on { padding:6px 14px; border:none; border-radius:6px; background:#166534; color:#86efac; font-size:12px; cursor:pointer; }
.btn-off { padding:6px 14px; border:none; border-radius:6px; background:#7f1d1d; color:#fca5a5; font-size:12px; cursor:pointer; }
.dot { font-size:10px; color:#475569; }
.dot.on { color:#86efac; }
/* messages */
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
/* input */
.input-row { display:flex; gap:8px; padding:12px 16px; border-top:1px solid #334155; background:#1e293b; }
.input-row input { flex:1; padding:10px 14px; border:1px solid #334155; border-radius:8px; background:#0f172a; color:#e2e8f0; font-size:14px; outline:none; }
.input-row input:focus { border-color:#1e3a5f; }
.input-row input:disabled { opacity:.4; }
.input-row button { padding:10px 24px; border:none; border-radius:8px; background:#1e3a5f; color:#38bdf8; font-size:14px; cursor:pointer; }
.input-row button:disabled { opacity:.4; }
/* admin */
.admin-panel { flex:1; display:flex; flex-direction:column; padding:16px; overflow-y:auto; gap:12px; }
.api-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(180px,1fr)); gap:6px; }
.api-btn { display:flex; flex-direction:column; gap:2px; padding:10px 12px; border:1px solid #334155; border-radius:6px; background:#0f172a; color:#e2e8f0; cursor:pointer; text-align:left; }
.api-btn:hover { border-color:#1e3a5f; }
.api-btn:disabled { opacity:.4; }
.api-path { font-family:monospace; font-size:13px; color:#38bdf8; }
.api-desc { font-size:11px; color:#64748b; }
.param-form { background:#0f172a; border:1px solid #334155; border-radius:8px; padding:16px; display:flex; flex-direction:column; gap:10px; }
.param-title { font-family:monospace; font-size:14px; color:#38bdf8; }
.param-row { display:flex; align-items:center; gap:8px; }
.param-row label { font-size:12px; color:#94a3b8; min-width:60px; }
.param-row input { flex:1; padding:8px 10px; border:1px solid #334155; border-radius:6px; background:#1e293b; color:#e2e8f0; font-size:13px; outline:none; }
.param-row input:focus { border-color:#1e3a5f; }
.param-actions { display:flex; gap:8px; }
.api-result { background:#0f172a; border:1px solid #334155; border-radius:8px; padding:16px; font-family:monospace; font-size:12px; color:#94a3b8; white-space:pre-wrap; max-height:300px; overflow-y:auto; }
</style>
