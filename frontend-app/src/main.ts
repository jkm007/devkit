import { createSSRApp } from "vue";
import App from "./App.vue";
import uviewPlus from 'uview-plus';
import { createPinia } from 'pinia';
import { setupRouteInterceptor } from './utils/routeInterceptor';

export function createApp() {
  const app = createSSRApp(App);
  app.use(createPinia());
  app.use(uviewPlus);

  // 注册路由拦截器
  setupRouteInterceptor();

  return {
    app,
  };
}
