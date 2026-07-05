import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      // FEAT-07: switched from the default generateSW strategy to injectManifest — generateSW
      // only produces an offline precache with no hook point for a `push` event handler
      // (LOGIC-004 needs the SW to actually show a system notification on push). srcDir/filename
      // point at our own service worker; the precache manifest is still injected automatically.
      strategies: 'injectManifest',
      srcDir: 'src',
      filename: 'sw.ts',
      manifest: {
        name: 'Шеф-стол',
        short_name: 'Шеф-стол',
        description: 'Запись на кулинарные классы',
        // Нейтральная тема — бренд не зафиксирован (3-design-brief/00-foundations.md
        // "Объём визуальных требований"), значения ниже — временная заглушка.
        theme_color: '#ffffff',
        background_color: '#ffffff',
        display: 'standalone',
        lang: 'ru',
        icons: [],
      },
    }),
  ],
})
