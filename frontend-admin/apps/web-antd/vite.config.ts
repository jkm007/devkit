import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  return {
    application: {
      // 启用预压缩，生成 .gz 和 .br 文件
      compress: true,
      compressTypes: ['brotli', 'gzip'],
    },
    vite: {
      build: {
        rolldownOptions: {
          output: {
            // 手动 chunk 拆分策略，减少主 chunk 体积
            manualChunks(id: string) {
              // === 第三方库（node_modules）===
              if (id.includes('node_modules')) {
                // Vue 核心生态
                if (
                  id.includes('/vue/') ||
                  id.includes('/@vue/') ||
                  id.includes('/pinia/') ||
                  id.includes('/vue-router/')
                ) {
                  return 'vendor-vue';
                }
                // Ant Design Vue + Icons
                if (
                  id.includes('/ant-design-vue/') ||
                  id.includes('/@ant-design/') ||
                  id.includes('/ant-design') ||
                  id.includes('/antv/')
                ) {
                  return 'vendor-antd';
                }
                // ECharts 图表库
                if (id.includes('/echarts/') || id.includes('/zrender/')) {
                  return 'vendor-echarts';
                }
                // VXE Table
                if (id.includes('/vxe-')) {
                  return 'vendor-vxe';
                }
                // Day.js 日期库
                if (id.includes('/dayjs/') || id.includes('/dayjs.')) {
                  return 'vendor-dayjs';
                }
                // Tiptap 富文本编辑器
                if (id.includes('/@tiptap/') || id.includes('/tiptap/')) {
                  return 'vendor-tiptap';
                }
                // ProseMirror (Tiptap 底层依赖)
                if (id.includes('/prosemirror-')) {
                  return 'vendor-prosemirror';
                }
                // VueUse Vue 组合函数（匹配 pnpm 路径）
                if (
                  id.includes('/@vueuse/') ||
                  id.includes('@vueuse+core') ||
                  id.includes('@vueuse+integrations') ||
                  id.includes('@vueuse+metadata')
                ) {
                  return 'vendor-vueuse';
                }
                // Popper.js / Tippy.js
                if (
                  id.includes('/@popperjs/') ||
                  id.includes('/tippy.js') ||
                  id.includes('@popperjs+core') ||
                  id.includes('tippy.js@')
                ) {
                  return 'vendor-popper';
                }
                // SortableJS 拖拽
                if (id.includes('/sortablejs/') || id.includes('sortablejs@')) {
                  return 'vendor-sortable';
                }
                // Lodash/加密/二维码等工具库
                if (
                  id.includes('/lodash/') ||
                  id.includes('/lodash-es/') ||
                  id.includes('/crypto-js/') ||
                  id.includes('/jsencrypt/') ||
                  id.includes('/qrcode/') ||
                  id.includes('/md5/') ||
                  id.includes('/pinyin/') ||
                  id.includes('/js-cookie/') ||
                  id.includes('/nprogress/') ||
                  id.includes('/pako/')
                ) {
                  return 'vendor-utils';
                }
                // 其他第三方库
                return 'vendor-common';
              }

              // === 本地 workspace 包（@vben/* 通过 symlink）===
              if (id.includes('/packages/')) {
                // @vben/layouts 布局组件（较大）
                if (id.includes('/packages/effects/layouts/') || id.includes('/packages/@core/')) {
                  return 'vendor-vben-layout';
                }
                // @vben/access 权限
                if (id.includes('/packages/effects/access/')) {
                  return 'vendor-vben-access';
                }
                // @vben/request 请求
                if (id.includes('/packages/effects/request/')) {
                  return 'vendor-vben-request';
                }
                // @vben/plugins 插件
                if (id.includes('/packages/effects/plugins/')) {
                  return 'vendor-vben-plugins';
                }
                // @vben/common-ui 通用UI
                if (id.includes('/packages/effects/common-ui/')) {
                  return 'vendor-vben-ui';
                }
                // hooks/constants/locales/utils/stores 等
                if (
                  id.includes('/packages/hooks/') ||
                  id.includes('/packages/effects/hooks/')
                ) {
                  return 'vendor-vben-hooks';
                }
                if (
                  id.includes('/packages/constants/') ||
                  id.includes('/packages/locales/') ||
                  id.includes('/packages/utils/') ||
                  id.includes('/packages/stores/') ||
                  id.includes('/packages/icons/') ||
                  id.includes('/packages/types/') ||
                  id.includes('/packages/styles/') ||
                  id.includes('/packages/preferences/')
                ) {
                  return 'vendor-vben-base';
                }
                // 其他 @vben 包
                return 'vendor-vben-common';
              }
            },
          },
        },
      },
      server: {
        proxy: {
          '/api/v1': {
            changeOrigin: true,
            target: 'http://localhost:8080',
            ws: true,
          },
        },
      },
    },
  };
});
