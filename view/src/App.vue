<script setup lang="ts">
import { ref } from 'vue'
import ExTest from './components/ExTest.vue'
import WsTest from './components/WsTest.vue'
import ApiExplorer from './components/ApiExplorer.vue'

type Tab = 'HTTP 轮询' | 'WebSocket' | 'API 查询'
const tabs: Tab[] = ['HTTP 轮询', 'WebSocket', 'API 查询']
const active = ref<Tab>('HTTP 轮询')
</script>

<template>
  <div class="app">
    <header>
      <h1>httpPush Admin</h1>
      <nav>
        <button v-for="t in tabs" :key="t" :class="{ active: active === t }" @click="active = t">
          {{ t }}
        </button>
      </nav>
    </header>
    <main>
      <ExTest v-if="active === 'HTTP 轮询'" />
      <WsTest v-if="active === 'WebSocket'" />
      <ApiExplorer v-if="active === 'API 查询'" />
    </main>
  </div>
</template>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: system-ui, sans-serif; background: #0f172a; color: #e2e8f0; min-height: 100vh; }
.app { max-width: 960px; margin: 0 auto; padding: 0 20px; }
header { padding: 24px 0 16px; }
h1 { font-size: 22px; margin-bottom: 16px; color: #f1f5f9; }
nav { display: flex; gap: 4px; margin-bottom: 24px; }
nav button {
  padding: 8px 20px; border: 1px solid #334155; border-radius: 6px 6px 0 0;
  background: #1e293b; color: #94a3b8; cursor: pointer; font-size: 14px;
  transition: all .15s;
}
nav button:hover { color: #e2e8f0; background: #273548; }
nav button.active { background: #1e3a5f; color: #38bdf8; border-color: #1e3a5f; }
main { background: #1e293b; border-radius: 0 8px 8px 8px; padding: 24px; min-height: 400px; }
</style>
