import path from 'path'
import { defineConfig } from 'vite'
import reactRefresh from '@vitejs/plugin-react-refresh'
import mpa from 'vite-plugin-mpa'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [reactRefresh(), mpa()],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
        },
    },
    // server: {
    //     proxy: {
    //         '/api': {
    //             target: 'http://localhost:8000',
    //             changeOrigin: true,
    //             secure: false,
    //             ws: true,
    //         }
    //     }
    // }
})
