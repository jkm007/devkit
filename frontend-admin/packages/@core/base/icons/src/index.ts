import { addCollection } from '@iconify/vue';

// 常用图标集（按需加载）
import antDesignIcons from '@iconify/json/json/ant-design.json';
import carbonIcons from '@iconify/json/json/carbon.json';
import charmIcons from '@iconify/json/json/charm.json';
import epIcons from '@iconify/json/json/ep.json';
import fluentMdl2Icons from '@iconify/json/json/fluent-mdl2.json';
import ionIcons from '@iconify/json/json/ion.json';
import lucideIcons from '@iconify/json/json/lucide.json';
import mdiIcons from '@iconify/json/json/mdi.json';

/**
 * 初始化本地图标
 * 禁用远程 API，使用本地 JSON 数据
 */
function initLocalIcons() {
  // 添加常用图标集
  addCollection(antDesignIcons);
  addCollection(carbonIcons);
  addCollection(charmIcons);
  addCollection(epIcons);
  addCollection(fluentMdl2Icons);
  addCollection(ionIcons);
  addCollection(lucideIcons);
  addCollection(mdiIcons);
}

export * from './create-icon';

export * from './lucide';

export type { IconifyIcon as IconifyIconStructure } from '@iconify/vue';
export {
  addCollection,
  addIcon,
  Icon as IconifyIcon,
  listIcons,
} from '@iconify/vue';

export { initLocalIcons };