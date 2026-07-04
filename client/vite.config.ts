import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
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
