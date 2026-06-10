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

import { message } from 'ant-design-vue';

import { useAuthStore } from '#/store';

import { getDeviceId } from '#/utils/device-id';
import { showCaptchaVerify } from '#/utils/captcha-verify';
import { refreshTokenApi } from './core';

const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);

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
    const resp = await refreshTokenApi() as any;
    // baseRequestClient 返回原始 axios response，resp.data = { code, data, message }
    const result = resp?.data?.data ?? resp?.data;
    const newAccessToken = result?.accessToken || result;
    const newRefreshToken = result?.refreshToken;
    accessStore.setAccessToken(newAccessToken);
    if (newRefreshToken) {
      accessStore.setRefreshToken(newRefreshToken);
    }
    return newAccessToken;
  }

  function formatToken(token: null | string) {
    return token ? `Bearer ${token}` : null;
  }

  // 请求头处理
  client.addRequestInterceptor({
    fulfilled: async (config) => {
      const accessStore = useAccessStore();

      // 登录/注册/验证码等接口不需要携带token
      const skipAuthUrls = ['/auth/login', '/auth/register', '/auth/captcha', '/system/settings/public'];
      if (!skipAuthUrls.some((url) => config.url?.includes(url))) {
        config.headers.Authorization = formatToken(accessStore.accessToken);
      }
      config.headers['Accept-Language'] = preferences.app.locale;
      config.headers['X-Device-ID'] = getDeviceId();
      return config;
    },
  });

  // 风险评分拦截：处理 403001 需要验证码的情况（在数据解包之前拦截，保留原始 config）
  // 验证码错误时后端仍返回 403001，需要重新弹出验证框，不设重试次数限制
  client.addResponseInterceptor({
    fulfilled: async (response: any) => {
      const data = response?.data;
      if (data?.code === 403001) {
        try {
          // 循环：验证码错误时重新弹出验证框，直到成功或用户取消
          let currentConfig = { ...response.config };
          while (true) {
            const result = await showCaptchaVerify();
            currentConfig = {
              ...currentConfig,
              headers: {
                ...currentConfig.headers,
                'X-Captcha-Id': result.captchaId,
                'X-Captcha-Code': result.captchaCode,
                'X-Captcha-Start-Time': result.startTime ? String(result.startTime) : '',
              },
            };
            // 用原始 axios 实例发起重试，请求头已包含验证码信息
            // 注意：重试响应会再次经过拦截器链，defaultResponseInterceptor
            // 可能已将数据解包，需要同时处理两种情况
            let retryResult: any;
            try {
              retryResult = await client.instance.request(currentConfig);
            } catch (retryErr: any) {
              // defaultResponseInterceptor 在 code≠0 时 throw，检查是否是 403001
              const errData = retryErr?.response?.data ?? retryErr?.data;
              if (errData?.code === 403001) {
                continue; // 验证码错误，重新弹框
              }
              throw retryErr; // 其他错误，抛出
            }

            // retryResult 可能是完整的 axios response，也可能是被解包后的数据
            const retryResponseData = retryResult?.data ?? retryResult;
            const code = retryResponseData?.code;
            if (code === 0) {
              // 验证成功，返回业务数据
              return retryResponseData?.data ?? retryResult;
            }
            if (code === 403001) {
              continue; // 验证码错误，重新弹框
            }
            // 其他错误
            throw Object.assign({}, retryResult, { response: retryResult });
          }
        } catch {
          message.warning('操作已取消');
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
