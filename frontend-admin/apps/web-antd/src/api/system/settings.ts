import { requestClient } from '#/api/request';

/** 验证码响应数据类型 */
export type CaptchaResponse = {
  captcha_id: string;
  image: string;
  thumb?: string;
  thumb_x?: number;
  thumb_y?: number;
  type: string;
  hint_text?: string;
  chars?: string[];
  width?: number;
  height?: number;
  length?: number;
  start_time?: number;
};

export namespace SystemSettingsApi {
  export interface SettingItem {
    key: string;
    value: any;
    label: string;
    type: 'array' | 'boolean' | 'json' | 'number' | 'select' | 'string';
    options: Array<{ label: string; value: string }> | null;
    tip: string;
    sort: number;
    isPublic: boolean;
    isSensitive: boolean;
  }

  export type SettingsGroup = Record<string, SettingItem[]>;

  export interface UpdateResult {
    updated: number;
    restartRequired: boolean;
    restartItems: string[];
  }
}

/**
 * 获取公开配置（无需登录）
 */
async function getPublicSettings() {
  return requestClient.get<Record<string, any>>('/system/settings/public');
}

/**
 * 获取所有配置（按分组）
 */
async function getAllSettings() {
  return requestClient.get<SystemSettingsApi.SettingsGroup>('/system/settings');
}

/**
 * 获取指定分组配置
 */
async function getSettingsByGroup(group: string) {
  return requestClient.get<SystemSettingsApi.SettingItem[]>(
    `/system/settings/${group}`,
  );
}

/**
 * 批量更新配置
 */
async function updateSettings(settings: Record<string, Record<string, any>>) {
  return requestClient.put<SystemSettingsApi.UpdateResult>('/system/settings', {
    settings,
  });
}

/**
 * 更新指定分组配置
 */
async function updateSettingsByGroup(
  group: string,
  settings: Record<string, any>,
) {
  return requestClient.put<SystemSettingsApi.UpdateResult>(
    `/system/settings/${group}`,
    { settings },
  );
}

/**
 * 测试邮件发送
 */
async function testEmail(to: string) {
  return requestClient.post<{ sent: boolean }>('/system/settings/test-email', {
    to,
  });
}

/**
 * 测试短信发送
 */
async function testSms(phone: string) {
  return requestClient.post<{ sent: boolean }>('/system/settings/test-sms', {
    phone,
  });
}

/**
 * 获取验证码（无需登录）
 * @param type 验证码类型: numeric/slider/puzzle/point
 */
async function getCaptcha(type: string = 'numeric') {
  return requestClient.get<CaptchaResponse>('/auth/captcha', {
    params: { type },
  });
}

/**
 * 测试验证码生成（管理员用）
 * @param type 验证码类型: numeric/slider/puzzle/point
 */
async function testCaptcha(type: string = 'numeric') {
  return requestClient.get<CaptchaResponse>('/system/captcha/test', {
    params: { type },
  });
}

/**
 * 验证验证码（管理员测试用）
 */
async function verifyCaptcha(data: {
  captchaId: string;
  captchaCode: string;
  startTime: number;
  points?: Array<{ x: number; y: number }>;
}) {
  return requestClient.post<{ valid: boolean; message: string }>(
    '/system/captcha/verify',
    data,
  );
}

export {
  getAllSettings,
  getCaptcha,
  getPublicSettings,
  getSettingsByGroup,
  testCaptcha,
  testEmail,
  testSms,
  updateSettings,
  updateSettingsByGroup,
  verifyCaptcha,
};
