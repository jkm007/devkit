/**
 * 该文件可自行根据业务逻辑进行调整
 */
import type { RequestClientOptions } from '@vben/request';

import { useAppConfig } from '@vben/hooks';
import { preferences } from '@vben/preferences';
import {
  authenticateResponseInterceptor,
  defaultResponseInterceptor,
  errorMessageResponseInterceptor,
  RequestClient,
} from '@vben/request';
import { useAccessStore } from '@vben/stores';

import axios from 'axios';
import { message } from 'ant-design-vue';

import { useAuthStore } from '#/store';

import { getDeviceId } from '#/utils/device-id';
import { showCaptchaVerify } from '#/utils/captcha-verify';
import { refreshTokenApi } from './core';

/**
 * 登录页面自己的 CaptchaModal 是否正在显示
 * 由 login.vue 设置，request.ts 的 403001 拦截器检查
 * 当为 true 时，拦截器跳过全局弹窗，由 login.vue 自行处理
 */
export let loginCaptchaModalActive = false;

export function setLoginCaptchaModalActive(active: boolean) {
  loginCaptchaModalActive = active;
}

const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);

/**
 * 读取指定 Cookie 的值
 */
function getCookie(name: string): string | undefined {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
  return match ? match[2] : undefined;
}

// 干净的 axios 实例（无拦截器），专门用于验证码重试请求
const captchaRetryAxios = axios.create({ baseURL: apiURL, timeout: 10_000 });

function createRequestClient(baseURL: string, options?: RequestClientOptions) {
  const client = new RequestClient({
    ...options,
    baseURL,
  });

  /**
   * 重新认证逻辑
   */
  async function doReAuthenticate() {
    console.warn('Access token or refresh token is invalid or expired. ');
    const accessStore = useAccessStore();
    const authStore = useAuthStore();
    accessStore.setAccessToken(null);
    if (
      preferences.app.loginExpiredMode === 'modal' &&
      accessStore.isAccessChecked
    ) {
      accessStore.setLoginExpired(true);
    } else {
      await authStore.logout();
    }
  }

  /**
   * 刷新token逻辑
   */
  async function doRefreshToken() {
    const accessStore = useAccessStore();
    // baseRequestClient 无拦截器，返回原始 Axios response（框架 request() 内部 as T 转换导致类型不精确）
    // 运行时 resp 实际是 AxiosResponse，resp.data = { code, data: { accessToken, refreshToken }, message }
    const resp = await refreshTokenApi();
    const { accessToken } = (resp as any).data.data;
    accessStore.setAccessToken(accessToken);
    return accessToken;
  }

  function formatToken(token: null | string) {
    return token ? `Bearer ${token}` : null;
  }

  // 请求头处理
  client.addRequestInterceptor({
    fulfilled: async (config) => {
      const accessStore = useAccessStore();

      // 登录/注册/验证码等接口不需要携带token
      const skipAuthUrls = [
        '/auth/login',
        '/auth/login-by-email',
        '/auth/login-by-phone',
        '/auth/register',
        '/auth/captcha',
        '/auth/refresh',
        '/auth/send-code',
        '/auth/verify-code',
        '/auth/send-sms-code',
        '/auth/reset-password',
        '/auth/oauth/authorize',
        '/auth/oauth/callback',
        '/auth/wechat/',
        '/share/',
        '/system/settings/public',
      ];
      const shouldSkipAuth = skipAuthUrls.some((url) =>
        config.url?.includes(url),
      );
      if (!shouldSkipAuth && accessStore.accessToken) {
        config.headers.Authorization = formatToken(accessStore.accessToken);
      }
      config.headers['Accept-Language'] = preferences.app.locale;
      config.headers['X-Device-ID'] = getDeviceId();

      // CSRF Token：从 Cookie 读取并放入请求头（Double Submit Cookie 模式）
      // 非安全方法（POST/PUT/DELETE/PATCH）必须携带
      const method = config.method?.toUpperCase();
      if (method && !['GET', 'HEAD', 'OPTIONS'].includes(method)) {
        const csrfToken = getCookie('csrf_token');
        if (csrfToken) {
          config.headers['X-CSRF-Token'] = csrfToken;
        }
      }

      return config;
    },
  });

  // 风险评分拦截：处理 403001 需要验证码的情况（在数据解包之前拦截，保留原始 config）
  // 验证码错误时后端仍返回 403001，需要重新弹出验证框，最多重试 5 次
  client.addResponseInterceptor({
    fulfilled: async (response: any) => {
      const data = response?.data;
      if (data?.code === 403001) {
        // 登录页面有自己的 CaptchaModal 正在显示时，通知 CaptchaModal 刷新
        if (loginCaptchaModalActive) {
          window.dispatchEvent(
            new CustomEvent('captcha:login-verify-error', {
              detail: { message: data?.message || '验证码错误，请重试' },
            }),
          );
          // 拒绝请求，__captchaError 标记让 errorMessageResponseInterceptor 跳过提示
          const err: any = new Error(data?.message || '验证码错误');
          err.__captchaError = true;
          err.data = data;
          err.response = response;
          throw err;
        }
        try {
          // 从后端响应中读取指定的验证码类型（随机类型）
          let currentCaptchaType = data?.data?.captcha_type || '';
          // 保留原始请求头（含 Authorization、Accept-Language 等）
          const originalHeaders = { ...response.config.headers };
          const maxRetries = 5;
          // 循环：验证码错误时复用已有弹窗刷新验证码，直到成功或达到上限
          for (let retry = 0; retry < maxRetries; retry++) {
            // 第一次弹出弹窗，后续复用已有弹窗（弹窗自行刷新验证码）
            const result = await showCaptchaVerify(currentCaptchaType, retry > 0);
            // 用干净的 axios 实例重试（无拦截器），避免响应被二次处理
            const retryResp = await captchaRetryAxios.request({
              url: response.config.url,
              method: response.config.method,
              params: response.config.params,
              data: response.config.data,
              headers: {
                ...originalHeaders,
                'X-Captcha-Id': result.captchaId,
                'X-Captcha-Code': result.captchaCode,
                'X-Captcha-Start-Time': result.startTime
                  ? String(result.startTime)
                  : '',
              },
            });
            const retryBody = retryResp?.data;
            if (retryBody?.code === 0) {
              // 验证成功：通知弹窗关闭
              window.dispatchEvent(new CustomEvent('captcha:verify-success'));
              // 构造一个符合 axios 响应格式的伪响应对象
              // 传给后续的 defaultResponseInterceptor 进行统一解包
              return {
                data: retryBody,
                status: 200,
                statusText: 'OK',
                headers: retryResp.headers,
                config: response.config,
              };
            }
            if (retryBody?.code === 403001) {
              // 验证码错误：通知弹窗刷新验证码并显示错误
              window.dispatchEvent(
                new CustomEvent('captcha:verify-error', {
                  detail: { message: retryBody?.message || '验证码错误，请重试' },
                }),
              );
              currentCaptchaType =
                retryBody?.data?.captcha_type || currentCaptchaType;
              continue;
            }
            // 其他业务错误
            window.dispatchEvent(new CustomEvent('captcha:verify-success'));
            throw Object.assign({}, retryResp, { response: retryResp });
          }
          // 达到重试上限
          window.dispatchEvent(new CustomEvent('captcha:verify-success'));
          message.warning('验证码验证次数过多，请稍后再试');
          return Promise.reject(new Error('验证码验证次数超限'));
        } catch (e: any) {
          // 区分用户取消和真实错误
          window.dispatchEvent(new CustomEvent('captcha:verify-success'));
          if (e?.message === '用户取消验证码验证') {
            message.warning('操作已取消');
          } else {
            console.error('验证码验证异常:', e);
            message.warning('操作已取消');
          }
          return Promise.reject(new Error('验证码验证已取消'));
        }
      }
      return response;
    },
    rejected: async (error: any) => {
      return Promise.reject(error);
    },
  });

  // 处理返回的响应数据格式
  client.addResponseInterceptor(
    defaultResponseInterceptor({
      codeField: 'code',
      dataField: 'data',
      successCode: 0,
    }),
  );

  // token过期的处理
  client.addResponseInterceptor(
    authenticateResponseInterceptor({
      client,
      doReAuthenticate,
      doRefreshToken,
      enableRefreshToken: preferences.app.enableRefreshToken,
      formatToken,
    }),
  );

  // 通用的错误处理,如果没有进入上面的错误处理逻辑，就会进入这里
  client.addResponseInterceptor(
    errorMessageResponseInterceptor((msg: string, error) => {
      // 403001 验证码错误由页面自行处理，不显示全局错误提示
      if (error?.__captchaError) return;
      // 这里可以根据业务进行定制,你可以拿到 error 内的信息进行定制化处理，根据不同的 code 做不同的提示，而不是直接使用 message.error 提示 msg
      // 当前mock接口返回的错误字段是 error 或者 message
      const responseData = error?.response?.data ?? {};
      const errorMessage = responseData?.error ?? responseData?.message ?? '';
      // 如果没有错误信息，则会根据状态码进行提示
      message.error(errorMessage || msg);
    }),
  );

  return client;
}

export const requestClient = createRequestClient(apiURL, {
  responseReturn: 'data',
});

export const baseRequestClient = new RequestClient({ baseURL: apiURL });
