<script setup lang="ts">
import { ref } from 'vue'
import ChatView from './ChatView.vue'
import BroadcastView from './BroadcastView.vue'
import AdminView from './AdminView.vue'

type Page = 'chat' | 'broadcast' | 'admin'
type ChatTab = 'ws' | 'http'
type AdminTab = 'users' | 'info'

const page = ref<Page>('chat')
const chatTab = ref<ChatTab>('ws')
const adminTab = ref<AdminTab>('users')
const host = ref('127.0.0.1:8084')
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
      <button :class="{ sel: page === 'broadcast' }" @click="page = 'broadcast'">📢 群发</button>
      <div class="menu-group">
        <button :class="{ sel: page === 'admin' }" @click="page = 'admin'">⚙ 管理</button>
        <div v-show="page === 'admin'" class="sub">
          <button :class="{ a: adminTab === 'users' }" @click="adminTab = 'users'">👥 用户</button>
          <button :class="{ a: adminTab === 'info' }" @click="adminTab = 'info'">📊 基本信息</button>
        </div>
      </div>
    </nav>

    <div class="right">
      <ChatView v-show="page === 'chat'" :host="host" :tab="chatTab" />
      <BroadcastView v-show="page === 'broadcast'" :host="host" />
      <AdminView v-show="page === 'admin'" :host="host" :tab="adminTab" />
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
</style>
