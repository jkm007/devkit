import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api/v1': {
            changeOrigin: true,
            // 真实后端服务地址
            target: 'http://localhost:8080',
            ws: true,
          },
        },
      },
    },
  };
});
