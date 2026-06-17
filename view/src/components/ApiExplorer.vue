<script setup lang="ts">
import { ref } from 'vue'

interface Endpoint {
  path: string
  desc: string
  params?: string[]
}

const endpoints: Endpoint[] = [
  { path: '/root_version', desc: '版本信息' },
  { path: '/queryVersion', desc: '集群版本' },
  { path: '/info_user', desc: '集群信息（节点数/用户数）' },
  { path: '/onlineUser', desc: '在线用户列表', params: ['size'] },
  { path: '/queryUser', desc: '查询用户连接', params: ['id'] },
  { path: '/queryOrderInfo', desc: '用户连接排序', params: ['id'] },
  { path: '/queryClusterUserNum', desc: '集群用户数' },
  { path: '/queryGroupInfo', desc: '群组信息' },
  { path: '/queryTimeWheelLog', desc: '时间轮日志' },
]

interface Result {
  path: string
  data: any
  error?: string
}

const results = ref<Result[]>([])
const loading = ref('')

async function query(ep: Endpoint) {
  loading.value = ep.path
  try {
    let url = ep.path
    if (ep.params) {
      const qs = ep.params.map(p => {
        const v = prompt(`参数 ${p}?`)
        return v ? `${p}=${encodeURIComponent(v)}` : ''
      }).filter(Boolean).join('&')
      if (qs) url += '?' + qs
    }
    const resp = await fetch(url)
    const data = await resp.json()
    const idx = results.value.findIndex(r => r.path === ep.path)
    if (idx >= 0) results.value[idx] = { path: ep.path, data }
    else results.value.push({ path: ep.path, data })
  } catch (e: any) {
    const idx = results.value.findIndex(r => r.path === ep.path)
    if (idx >= 0) results.value[idx] = { path: ep.path, data: null, error: e.message }
    else results.value.push({ path: ep.path, data: null, error: e.message })
  } finally {
    loading.value = ''
  }
}
</script>

<template>
  <div>
    <h2>API 查询</h2>
    <div class="grid">
      <button
        v-for="ep in endpoints" :key="ep.path"
        class="ep-btn" :class="{ loading: loading === ep.path }"
        @click="query(ep)"
      >
        <span class="path">{{ ep.path }}</span>
        <span class="desc">{{ ep.desc }}</span>
      </button>
    </div>
    <div v-if="results.length" class="results">
      <div v-for="r in results" :key="r.path" class="result-card">
        <h3>{{ r.path }}</h3>
        <pre v-if="r.error" class="error">{{ r.error }}</pre>
        <pre v-else>{{ JSON.stringify(r.data, null, 2) }}</pre>
      </div>
    </div>
  </div>
</template>

<style scoped>
h2 { font-size: 18px; margin-bottom: 16px; color: #f1f5f9; }
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 8px; margin-bottom: 24px; }
.ep-btn {
  display: flex; flex-direction: column; align-items: flex-start; gap: 4px;
  padding: 10px 14px; border: 1px solid #334155; border-radius: 6px;
  background: #0f172a; color: #e2e8f0; cursor: pointer; text-align: left;
  transition: all .15s;
}
.ep-btn:hover { border-color: #1e3a5f; background: #1a2740; }
.ep-btn.loading { opacity: .6; pointer-events: none; }
.path { font-family: monospace; font-size: 14px; color: #38bdf8; }
.desc { font-size: 12px; color: #64748b; }
.results { display: flex; flex-direction: column; gap: 16px; }
.result-card { background: #0f172a; border: 1px solid #334155; border-radius: 6px; padding: 14px; }
.result-card h3 { font-size: 14px; color: #38bdf8; margin-bottom: 8px; }
pre {
  font-size: 12px; color: #94a3b8; white-space: pre-wrap; word-break: break-all;
  max-height: 300px; overflow-y: auto;
}
.error { color: #fca5a5; }
</style>
