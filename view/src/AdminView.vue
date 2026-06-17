<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{ host: string; tab: 'users' | 'info' }>()

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
  try { const r = await fetch(`http://${props.host}/onlineUser`); onlineUsers.value = (await r.json())?.data || [] } catch { /* */ }
}
async function queryUser() {
  if (!userQueryId.value) return
  try { const r = await fetch(`http://${props.host}/queryUser?id=${userQueryId.value}`); userDetail.value = await r.json() } catch { /* */ }
}
async function loadInfoCard(c: typeof infoCards.value[0]) {
  c.loading = true
  try { const r = await fetch(`http://${props.host}${c.path}`); c.data = await r.json() } catch { /* */ } finally { c.loading = false }
}
function loadAllInfo() { infoCards.value.forEach(c => loadInfoCard(c)) }

watch(() => props.tab, t => { if (t === 'info') loadAllInfo(); if (t === 'users') loadOnlineUsers() }, { immediate: true })
</script>

<template>
  <div class="admin-panel">
    <!-- 用户 -->
    <div v-if="tab === 'users'" class="admin-body">
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
    <div v-if="tab === 'info'" class="admin-body">
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
</template>

<style scoped>
.admin-panel { flex:1; display:flex; flex-direction:column; }
.admin-body { flex:1; overflow-y:auto; padding:16px; }
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
.hint { color:#475569; text-align:center; margin-top:40px; }
.info-cards { display:grid; grid-template-columns:repeat(auto-fill,minmax(280px,1fr)); gap:8px; }
.info-card { background:#0f172a; border:1px solid #334155; border-radius:8px; padding:14px; cursor:pointer; transition:border-color .15s; }
.info-card:hover { border-color:#1e3a5f; }
.ic-path { font-family:monospace; font-size:13px; color:#38bdf8; margin-bottom:2px; }
.ic-desc { font-size:11px; color:#64748b; margin-bottom:8px; }
.ic-data { font-family:monospace; font-size:11px; color:#94a3b8; white-space:pre-wrap; max-height:200px; overflow-y:auto; }
.ic-hint { font-size:11px; color:#475569; }
</style>
