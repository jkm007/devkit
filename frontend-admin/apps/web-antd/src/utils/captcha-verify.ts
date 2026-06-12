/**
 * 验证码验证工具
 * 用于 API 拦截器中弹出验证码弹窗，获取验证结果
 */

export interface CaptchaVerifyResult {
  captchaId: string;
  captchaCode: string;
  startTime?: number;
}

/**
 * 弹出验证码弹窗，等待用户完成验证
 * 通过全局事件触发 captcha-verify-modal 组件
 * @param captchaType 后端指定的验证码类型（可选，如 numeric/slider/puzzle/rotation/point）
 * @param reuse 为 true 时不重新触发弹窗（用于重试场景，弹窗已打开且会自行刷新）
 * 返回 captchaId 和 captchaCode，用于请求头
 */
export function showCaptchaVerify(
  captchaType?: string,
  reuse?: boolean,
): Promise<CaptchaVerifyResult> {
  return new Promise((resolve, reject) => {
    // 验证完成或取消时，手动移除所有相关监听，避免冗余
    function cleanup() {
      window.removeEventListener(
        'captcha:verify-result',
        onResult as EventListener,
      );
      window.removeEventListener(
        'captcha:verify-cancel',
        onCancel as EventListener,
      );
    }

    // 监听验证结果
    function onResult(event: CustomEvent) {
      cleanup();
      if (event.detail) {
        resolve(event.detail);
      } else {
        reject(new Error('验证码验证失败'));
      }
    }

    function onCancel() {
      cleanup();
      reject(new Error('用户取消验证码验证'));
    }

    window.addEventListener('captcha:verify-result', onResult as EventListener);
    window.addEventListener('captcha:verify-cancel', onCancel as EventListener);

    // reuse 模式下不触发弹窗（弹窗已打开，会自行刷新验证码）
    if (!reuse) {
      window.dispatchEvent(
        new CustomEvent('captcha:show-verify', {
          detail: { captchaType: captchaType || '' },
        }),
      );
    }
  });
}
