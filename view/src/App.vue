<script setup lang="ts">
import { ref, nextTick, watch } from 'vue'

type Page = 'chat' | 'admin'
type ChatTab = 'ws' | 'http'
type AdminTab = 'users' | 'info'

const page = ref<Page>('chat')
const chatTab = ref<ChatTab>('ws')
const adminTab = ref<AdminTab>('users')
const userId = ref('user_a')
const targetId = ref('user_b')
const sendMode = ref<'single' | 'group'>('single')
const groupId = ref('all')
const connected = ref(false)
const messages = ref<{ from: string; body: string; time: string; self: boolean }[]>([])
const host = ref('127.0.0.1:8084')

let ws: WebSocket | null = null
let pollCtrl: AbortController | null = null

function addMsg(from: string, body: string, self: boolean) {
  messages.value.push({ from, body, time: new Date().toLocaleTimeString(), self })
  nextTick(() => { const el = document.getElementById('chatList'); if (el) el.scrollTop = el.scrollHeight })
}

async function connect() {
  if (connected.value) return
  messages.value = []
  if (chatTab.value === 'ws') {
    ws = new WebSocket(`ws://${host.value}/ws?id=${userId.value}`)
    ws.onopen = () => { connected.value = true; addMsg('system', `已连接 (WS)`, false) }
    ws.onmessage = (e) => {
      try { JSON.parse(e.data).forEach((m: any) => addMsg(m.from || m.From, m.body || m.Body || m.msg || m.Msg, false)) } catch { /* */ }
    }
    ws.onclose = () => { connected.value = false; addMsg('system', '已断开', false) }
  } else {
    connected.value = true; addMsg('system', '已连接 (HTTP)', false); poll()
  }
}

async function poll() {
  while (connected.value && chatTab.value === 'http') {
    pollCtrl = new AbortController()
    try {
      const r = await fetch(`http://${host.value}/ex?id=${userId.value}&liveTime=15`, { signal: pollCtrl.signal })
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
  if (sendMode.value === 'group') {
    fetch(`http://${host.value}/sendGroupMsg?groupId=${groupId.value}&msg=${encodeURIComponent(body)}`)
  } else if (chatTab.value === 'ws' && ws) {
    ws.send(JSON.stringify({ to: targetId.value, msg: body }))
  } else {
    fetch(`http://${host.value}/sendmsg?username=${targetId.value}&msg=${encodeURIComponent(body)}`)
  }
}

// ---- admin ----
const onlineUsers = ref<any[]>([])
const userDetail = ref<any>(null)
const userQueryId = ref('')
const infoCards = ref([
  { path: '/root_version', desc: '版本信息', data: null as any, loading: false },
  { path: '/info_user', desc: '集群信息', data: null as any, loading: false },
  { path: '/queryClusterUserNum', desc: '集群用户数', data: null as any, loading: false },
  { path: '/queryGroupInfo', desc: '群组信息', data: null as any, loading: false },
  { path: '/queryVersion', desc: '集群版本', data: null as any, loading: false },
])

async function loadOnlineUsers() {
  try { const r = await fetch(`http://${host.value}/onlineUser`); onlineUsers.value = (await r.json())?.data || (await r.json()) || [] } catch { /* */ }
}
async function queryUser() {
  if (!userQueryId.value) return
  try { const r = await fetch(`http://${host.value}/queryUser?id=${userQueryId.value}`); userDetail.value = await r.json() } catch { /* */ }
}
async function loadInfoCard(c: typeof infoCards.value[0]) {
  c.loading = true
  try { const r = await fetch(`http://${host.value}${c.path}`); c.data = await r.json() } catch { /* */ } finally { c.loading = false }
}
function loadAllInfo() { infoCards.value.forEach(c => loadInfoCard(c)) }

watch(page, p => { if (p === 'admin') loadOnlineUsers() })
watch(adminTab, t => { if (t === 'info') loadAllInfo(); if (t === 'users') loadOnlineUsers() })
watch(chatTab, () => { if (connected.value) { disconnect(); connect() } })
</script>

<template>
  <div class="layout">
    <nav class="menu">
      <div class="logo">httpPush</div>
      <div class="menu-group">
        <button :class="{ sel: page === 'chat' }" @click="page = 'chat'">💬 聊天</button>
        <div v-show="page === 'chat'" class="sub">
          <button :class="{ a: chatTab === 'ws' }" @click="chatTab = 'ws'">WebSocket</button>
          <button :class="{ a: chatTab === 'http' }" @click="chatTab = 'http'">HTTP 轮询</button>
        </div>
      </div>
      <div class="menu-group">
        <button :class="{ sel: page === 'admin' }" @click="page = 'admin'">⚙ 管理</button>
        <div v-show="page === 'admin'" class="sub">
          <button :class="{ a: adminTab === 'users' }" @click="adminTab = 'users'">👥 用户</button>
          <button :class="{ a: adminTab === 'info' }" @click="adminTab = 'info'">📊 基本信息</button>
        </div>
      </div>
    </nav>

    <div class="right">
      <!-- ============ 聊天 ============ -->
      <div v-show="page === 'chat'" class="chat-panel">
        <div class="chat-topbar">
          <span class="proto-tag">{{ chatTab === 'ws' ? 'WebSocket' : 'HTTP 轮询' }}</span>
          <input v-model="host" placeholder="地址" class="sm" style="width:140px" />
          <input v-model="userId" placeholder="你的 ID" class="sm" />
          <span class="arr">→</span>
          <div class="mode-switch">
            <button :class="{ a: sendMode === 'single' }" @click="sendMode = 'single'">单发</button>
            <button :class="{ a: sendMode === 'group' }" @click="sendMode = 'group'">群发</button>
          </div>
          <input v-if="sendMode === 'single'" v-model="targetId" placeholder="目标 ID" class="sm" />
          <input v-else v-model="groupId" placeholder="群组 ID" class="sm" />
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

      <!-- ============ 管理 ============ -->
      <div v-show="page === 'admin'" class="admin-panel">
        <!-- 用户 -->
        <div v-if="adminTab === 'users'" class="admin-body">
          <div class="user-search">
            <input v-model="userQueryId" placeholder="输入用户 ID 查询" @keyup.enter="queryUser" />
            <button @click="queryUser">查询</button>
          </div>
          <pre v-if="userDetail" class="json-box">{{ JSON.stringify(userDetail, null, 2) }}</pre>

          <div class="section-title">在线用户 ({{ onlineUsers.length }})</div>
          <div class="user-list" v-if="onlineUsers.length">
            <div v-for="u in onlineUsers" :key="u.userName || u.UserName" class="user-row">
              <span class="uname">{{ u.userName || u.UserName }}</span>
              <span class="umeta" v-if="u.machineId || u.MachineId">节点: {{ u.machineId || u.MachineId }}</span>
              <span class="umeta" v-if="u.createTime || u.CreateTime">{{ u.createTime || u.CreateTime }}</span>
            </div>
          </div>
          <div v-else class="hint">暂无在线用户</div>
        </div>

        <!-- 基本信息 -->
        <div v-if="adminTab === 'info'" class="admin-body">
          <div class="info-cards">
            <div v-for="c in infoCards" :key="c.path" class="info-card" @click="loadInfoCard(c)">
              <div class="ic-path">{{ c.path }}</div>
              <div class="ic-desc">{{ c.desc }}</div>
              <pre v-if="c.data" class="ic-data">{{ JSON.stringify(c.data, null, 2) }}</pre>
              <div v-else class="ic-hint">{{ c.loading ? '加载中...' : '点击加载' }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
* { margin:0; padding:0; box-sizing:border-box; }
body { font-family:system-ui,-apple-system,sans-serif; background:#0f172a; color:#e2e8f0; }
#app { height:100vh; }
.layout { display:flex; height:100vh; }
.menu { width:180px; background:#1e293b; border-right:1px solid #334155; display:flex; flex-direction:column; padding:12px; gap:4px; flex-shrink:0; }
.logo { font-size:16px; font-weight:700; color:#38bdf8; padding:8px 12px 16px; }
.menu button { text-align:left; padding:10px 12px; border:none; border-radius:6px; background:transparent; color:#94a3b8; font-size:13px; cursor:pointer; }
.menu button:hover { background:#273548; color:#e2e8f0; }
.menu button.sel { background:#1e3a5f; color:#38bdf8; }
.menu-group { display:flex; flex-direction:column; }
.sub { display:flex; flex-direction:column; padding-left:20px; gap:2px; margin-bottom:4px; }
.sub button { font-size:12px; padding:6px 10px; }
.sub button.a { color:#38bdf8; }
.right { flex:1; display:flex; flex-direction:column; min-width:0; }

/* chat */
.chat-panel { flex:1; display:flex; flex-direction:column; }
.chat-topbar { display:flex; align-items:center; gap:8px; padding:10px 16px; background:#1e293b; border-bottom:1px solid #334155; }
.chat-topbar input.sm { width:100px; padding:6px 8px; border:1px solid #334155; border-radius:6px; background:#0f172a; color:#e2e8f0; font-size:12px; outline:none; }
.proto-tag { font-size:11px; font-weight:600; padding:4px 8px; border-radius:4px; background:#1e3a5f; color:#38bdf8; white-space:nowrap; }
.chat-topbar input.sm:focus { border-color:#1e3a5f; }
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

/* admin */
.admin-panel { flex:1; display:flex; flex-direction:column; }
.admin-body { flex:1; overflow-y:auto; padding:16px; }
/* user */
.user-search { display:flex; gap:8px; margin-bottom:16px; }
.user-search input { flex:1; padding:8px 12px; border:1px solid #334155; border-radius:6px; background:#0f172a; color:#e2e8f0; font-size:13px; outline:none; }
.user-search input:focus { border-color:#1e3a5f; }
.user-search button { padding:8px 16px; border:none; border-radius:6px; background:#1e3a5f; color:#38bdf8; font-size:13px; cursor:pointer; }
.json-box { background:#0f172a; border:1px solid #334155; border-radius:8px; padding:14px; font-family:monospace; font-size:12px; color:#94a3b8; white-space:pre-wrap; margin-bottom:16px; max-height:240px; overflow-y:auto; }
.section-title { font-size:13px; font-weight:600; color:#94a3b8; margin-bottom:8px; }
.user-list { display:flex; flex-direction:column; gap:6px; }
.user-row { display:flex; align-items:center; gap:12px; padding:10px 14px; background:#0f172a; border:1px solid #334155; border-radius:6px; font-size:13px; }
.uname { color:#e2e8f0; font-weight:500; }
.umeta { color:#64748b; font-size:11px; }
/* info */
.info-cards { display:grid; grid-template-columns:repeat(auto-fill,minmax(280px,1fr)); gap:8px; }
.info-card { background:#0f172a; border:1px solid #334155; border-radius:8px; padding:14px; cursor:pointer; transition:border-color .15s; }
.info-card:hover { border-color:#1e3a5f; }
.ic-path { font-family:monospace; font-size:13px; color:#38bdf8; margin-bottom:2px; }
.ic-desc { font-size:11px; color:#64748b; margin-bottom:8px; }
.ic-data { font-family:monospace; font-size:11px; color:#94a3b8; white-space:pre-wrap; max-height:200px; overflow-y:auto; }
.ic-hint { font-size:11px; color:#475569; }
</style>
