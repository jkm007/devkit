/**
 * 验证码验证工具
 * 用于 API 拦截器中弹出验证码弹窗，获取验证结果
 */

export interface CaptchaVerifyResult {
  captchaId: string;
  captchaCode: string;
}

/**
 * 弹出验证码弹窗，等待用户完成验证
 * 通过全局事件触发 captcha-verify-modal 组件
 * 返回 captchaId 和 captchaCode，用于请求头
 */
export function showCaptchaVerify(): Promise<CaptchaVerifyResult> {
  return new Promise((resolve, reject) => {
    // 监听验证结果
    function onResult(event: CustomEvent) {
      window.removeEventListener('captcha:verify-result', onResult as EventListener);
      window.removeEventListener('captcha:verify-cancel', onCancel as EventListener);
      if (event.detail) {
        resolve(event.detail);
      } else {
        reject(new Error('验证码验证失败'));
      }
    }

    function onCancel() {
      window.removeEventListener('captcha:verify-result', onResult as EventListener);
      window.removeEventListener('captcha:verify-cancel', onCancel as EventListener);
      reject(new Error('用户取消验证码验证'));
    }

    window.addEventListener('captcha:verify-result', onResult as EventListener, { once: true });
    window.addEventListener('captcha:verify-cancel', onCancel as EventListener, { once: true });

    // 触发弹窗显示
    window.dispatchEvent(new CustomEvent('captcha:show-verify'));
  });
}
