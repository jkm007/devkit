import { createSSRApp } from "vue";
import App from "./App.vue";
import uviewPlus from 'uview-plus';
import { createPinia } from 'pinia';
import { setupRouteInterceptor } from './utils/routeInterceptor';
import { initNetworkListener, loadQueue } from './utils/offline';

export function createApp() {
  const app = createSSRApp(App);
  app.use(createPinia());
  app.use(uviewPlus);

  // 注册路由拦截器
  setupRouteInterceptor();

  // 初始化网络状态监听 + 加载离线队列
  initNetworkListener();
  loadQueue();

  return {
    app,
  };
}
