import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/sendmsg': 'http://127.0.0.1:8084',
      '/sendMessage': 'http://127.0.0.1:8084',
      '/root_version': 'http://127.0.0.1:8084',
      '/queryUser': 'http://127.0.0.1:8084',
      '/onlineUser': 'http://127.0.0.1:8084',
      '/sendGroupMsg': 'http://127.0.0.1:8084',
      '/info_user': 'http://127.0.0.1:8084',
      '/queryOrderInfo': 'http://127.0.0.1:8084',
      '/queryTimeWheelLog': 'http://127.0.0.1:8084',
      '/queryClusterUserNum': 'http://127.0.0.1:8084',
      '/queryGroupInfo': 'http://127.0.0.1:8084',
      '/queryVersion': 'http://127.0.0.1:8084',
      '/ex': 'http://127.0.0.1:8084',
    }
  }
})
