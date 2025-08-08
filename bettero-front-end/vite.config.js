import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'; 

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { 
      '@Pages': path.resolve(__dirname, './src/components/Pages'),
      '@Forms': path.resolve(__dirname, './src/components/Forms'),
      '@Tables': path.resolve(__dirname, './src/components/Tables'),
      '@Charts': path.resolve(__dirname, './src/components/Charts'),
      '@components': path.resolve(__dirname, './src/components'),
      '@provider': path.resolve(__dirname, './src/provider'),
      '@utils': path.resolve(__dirname, './src/utils'),
      '@interface': path.resolve(__dirname, './src/interface')
    }
  }
})
