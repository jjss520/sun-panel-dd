import path from 'path'
import type { PluginOption } from 'vite'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'
import { createSvgIconsPlugin } from 'vite-plugin-svg-icons'

function setupPlugins(env: ImportMetaEnv): PluginOption[] {
  return [
    vue(),
    env.VITE_GLOB_APP_PWA === 'true' && VitePWA({
      injectRegister: 'auto',
      manifest: {
        name: 'Sun-Panel',
        short_name: 'Sun-Panel',
        icons: [
          { src: 'pwa-192x192.png', sizes: '192x192', type: 'image/png' },
          { src: 'pwa-512x512.png', sizes: '512x512', type: 'image/png' },
        ],
      },
    }),
    createSvgIconsPlugin({
      iconDirs: [path.resolve(process.cwd(), 'src/assets/svg-icons')],
      symbolId: '[name]',
    }),
  ]
}

export default defineConfig((env) => {
  const viteEnv = loadEnv(env.mode, process.cwd()) as unknown as ImportMetaEnv

  return {
    resolve: {
      alias: {
        '@': path.resolve(process.cwd(), 'src'),
      },
    },
    plugins: setupPlugins(viteEnv),
    server: {
      host: '0.0.0.0',
      port: 1002,
      open: false,
      proxy: {
        '/api': {
          target: 'http://localhost:3002',
          changeOrigin: true, // 允许跨域
          rewrite: path => path.replace(/^\/api\//, '/api/'),
        },
        '/uploads': {
          target: 'http://localhost:3002',
          changeOrigin: true, // 允许跨域
          rewrite: path => path.replace(/^\/uploads\//, '/uploads/'),
        },
      },
    },
    build: {
      target: 'es2015', // 兼容 Android 7+ 和旧版浏览器
      reportCompressedSize: false,
      sourcemap: false,
      commonjsOptions: {
        ignoreTryCatch: false,
      },
      rollupOptions: {
        output: {
          // 确保文件名包含哈希，打破浏览器缓存
          entryFileNames: 'assets/[name].[hash].js',
          chunkFileNames: 'assets/[name].[hash].js',
          assetFileNames: 'assets/[name].[hash].[ext]',
        },
      },
      terserOptions: {
        compress: {
          // 生产环境移除 console.log，减小包体积
          drop_console: true,
          // 移除 debugger
          drop_debugger: true,
        },
      },
    },
    css: {
      preprocessorOptions: {
        less: {
          javascriptEnabled: true,
        },
      },
    },
  }
})
