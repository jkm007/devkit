# 前端 Admin

基于 vue-vben-admin (web-antd) 的后台管理系统。

## 技术栈

- Vue 3
- Vite 8
- TypeScript
- Ant Design Vue 4
- Pinia
- Tailwind CSS 4

## 安装

```bash
pnpm install
```

## 开发

```bash
pnpm dev
```

## 构建

```bash
pnpm build
```

## 预览

```bash
pnpm preview
```

## 项目结构

```
frontend-admin/
├── apps/
│   ├── web-antd/          # 主开发目录
│   └── backend-mock/      # Mock 服务
├── packages/              # 共享包
│   ├── @core/             # 核心组件
│   ├── utils/             # 工具函数
│   ├── stores/            # 状态管理
│   ├── icons/             # 图标库
│   ├── locales/           # 国际化
│   └── styles/            # 样式
├── internal/              # 构建配置
├── scripts/               # 脚本工具
└── package.json
```

## 开发目录

业务开发主要在 `apps/web-antd/src/`:

- `views/` - 页面组件
- `api/` - API 接口
- `router/` - 路由配置
- `store/` - 业务状态
- `layouts/` - 布局组件
- `locales/` - 语言包

## 文档

原项目文档: https://doc.vben.pro

---

## 图标离线配置

项目默认使用本地图标，不依赖远程 API。

### 安装本地图标集

```bash
pnpm add @iconify/json
```

### 已配置的图标集

`packages/@core/base/icons/src/index.ts`:

| 图标集      | 文件             | 用途                  |
| ----------- | ---------------- | --------------------- |
| mdi         | mdi.json         | Material Design Icons |
| lucide      | lucide.json      | Lucide Icons          |
| carbon      | carbon.json      | Carbon Icons          |
| ant-design  | ant-design.json  | Ant Design Icons      |
| ion         | ion.json         | Ionicons              |
| ep          | ep.json          | Element Plus Icons    |
| charm       | charm.json       | Charm Icons           |
| fluent-mdl2 | fluent-mdl2.json | Fluent UI Icons       |

### 添加新图标集

如需其他图标集，修改 `index.ts`:

```typescript
import newIcons from '@iconify/json/json/{图标集名}.json';
addCollection(newIcons);
```

### 禁用远程 API

启动时调用 `initLocalIcons()` 后，不再请求 `api.iconify.design`。
