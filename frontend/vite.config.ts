import {defineConfig, searchForWorkspaceRoot} from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    fs: {
      allow: [
        '..'
      ],
    },
  },
  plugins: [vue()]
})
