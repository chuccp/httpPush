<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ host: string }>()

const groupId = ref('all')
const msg = ref('')
const result = ref('')
const sending = ref(false)

async function send() {
  if (!msg.value) return
  sending.value = true
  try {
    const r = await fetch(`http://${props.host}/sendGroupMsg?groupId=${groupId.value}&msg=${encodeURIComponent(msg.value)}`)
    const j = await r.json()
    result.value = `已发送 ${j?.data?.num || j?.num || 0} 人`
    msg.value = ''
  } catch (e: any) { result.value = '错误: ' + e.message }
  finally { sending.value = false }
}
</script>

<template>
  <div class="panel">
    <h2>📢 群发消息</h2>
    <div class="form">
      <label>群组 ID</label>
      <input v-model="groupId" placeholder="all" />
      <label>消息内容</label>
      <textarea v-model="msg" placeholder="输入群发消息..." rows="4" @keyup.ctrl.enter="send"></textarea>
      <div class="actions">
        <button class="btn-send" :disabled="sending" @click="send">{{ sending ? '发送中...' : '群发' }}</button>
        <span v-if="result" class="result">{{ result }}</span>
      </div>
      <div class="tip">Ctrl+Enter 快捷发送</div>
    </div>
  </div>
</template>

<style scoped>
.panel { flex:1; padding:24px; display:flex; flex-direction:column; }
h2 { font-size:18px; color:#f1f5f9; margin-bottom:20px; }
.form { display:flex; flex-direction:column; gap:8px; max-width:480px; }
label { font-size:12px; color:#64748b; }
input, textarea {
  padding:10px 14px; border:1px solid #334155; border-radius:8px;
  background:#0f172a; color:#e2e8f0; font-size:14px; outline:none; resize:vertical;
}
input:focus, textarea:focus { border-color:#1e3a5f; }
.actions { display:flex; align-items:center; gap:12px; }
.btn-send { padding:10px 32px; border:none; border-radius:8px; background:#1e3a5f; color:#38bdf8; font-size:14px; cursor:pointer; }
.btn-send:disabled { opacity:.4; }
.result { font-size:13px; color:#86efac; }
.tip { font-size:11px; color:#475569; }
</style>
