import { updatePreferences } from '@vben/preferences';

/**
 * 从后端获取公开配置并应用到前端偏好设置
 * 在应用启动时调用，将后端的系统设置同步到前端
 */
export async function applyPublicSettings() {
  try {
    const response = await fetch(
      `${import.meta.env.VITE_GLOB_API_URL || ''}/system/settings/public`,
    );
    const result = await response.json();

    if (result.code !== 0 || !result.data) {
      return;
    }

    const { basic } = result.data;

    if (!basic) return;

    // 必须用嵌套对象格式，不支持点号路径
    const updates: Record<string, any> = {};

    // 站点名称 → app.name
    if (basic.site_name !== undefined && basic.site_name !== '') {
      updates.app = updates.app || {};
      updates.app.name = basic.site_name;
    }

    // 站点 Logo → logo.source
    if (basic.site_logo !== undefined && basic.site_logo !== '') {
      updates.logo = updates.logo || {};
      updates.logo.source = basic.site_logo;
    }

    // 版权公司名 → copyright.companyName
    if (basic.copyright !== undefined && basic.copyright !== '') {
      updates.copyright = updates.copyright || {};
      updates.copyright.companyName = basic.copyright;
    }

    // 版权开关 → copyright.enable
    if (basic.copyright_enabled !== undefined) {
      updates.copyright = updates.copyright || {};
      updates.copyright.enable = basic.copyright_enabled;
    }

    // 公司网站 → copyright.companySiteLink
    if (basic.copyright_company_site !== undefined) {
      updates.copyright = updates.copyright || {};
      updates.copyright.companySiteLink = basic.copyright_company_site;
    }

    // 版权年份 → copyright.date
    if (basic.copyright_date !== undefined && basic.copyright_date !== '') {
      updates.copyright = updates.copyright || {};
      updates.copyright.date = basic.copyright_date;
    }

    // ICP 备案 → copyright.icp
    if (basic.copyright_icp !== undefined) {
      updates.copyright = updates.copyright || {};
      updates.copyright.icp = basic.copyright_icp;
    }

    // ICP 链接 → copyright.icpLink
    if (basic.copyright_icp_link !== undefined) {
      updates.copyright = updates.copyright || {};
      updates.copyright.icpLink = basic.copyright_icp_link;
    }

    // 水印开关 → app.watermark
    if (basic.watermark_enabled !== undefined) {
      updates.app = updates.app || {};
      updates.app.watermark = basic.watermark_enabled;
    }

    // 水印内容 → app.watermarkContent
    if (basic.watermark_content !== undefined) {
      updates.app = updates.app || {};
      updates.app.watermarkContent = basic.watermark_content;
    }

    // 水印透明度 → app.watermarkOpacity
    if (basic.watermark_opacity !== undefined) {
      updates.app = updates.app || {};
      updates.app.watermarkOpacity = basic.watermark_opacity;
    }

    // 页脚开关 → footer.enable
    if (basic.footer_enabled !== undefined) {
      updates.footer = updates.footer || {};
      updates.footer.enable = basic.footer_enabled;
    }

    // 固定页脚 → footer.fixed
    if (basic.footer_fixed !== undefined) {
      updates.footer = updates.footer || {};
      updates.footer.fixed = basic.footer_fixed;
    }

    // 默认主题 → theme.mode
    if (basic.default_theme !== undefined && basic.default_theme !== '') {
      const themeMode = basic.default_theme;
      if (
        themeMode === 'auto' ||
        themeMode === 'light' ||
        themeMode === 'dark'
      ) {
        updates.theme = updates.theme || {};
        updates.theme.mode = themeMode;
      }
    }

    // 默认语言 → app.locale
    if (basic.default_lang !== undefined && basic.default_lang !== '') {
      updates.app = updates.app || {};
      updates.app.locale = basic.default_lang;
    }

    // 应用更新
    if (Object.keys(updates).length > 0) {
      updatePreferences(updates);
    }
  } catch {
    // 静默失败，不影响应用启动
    console.warn('[Settings] Failed to load public settings from backend');
  }
}
