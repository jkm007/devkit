import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api': {
            changeOrigin: true,
            rewrite: (path) => path.replace(/^\/api/, ''),
            // 真实后端服务地址
            target: 'http://10.0.50.207:8080',
            ws: true,
          },
        },
      },
    },
  };
});
