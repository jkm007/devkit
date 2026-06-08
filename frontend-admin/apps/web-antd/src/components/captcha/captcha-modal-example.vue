<script setup lang="ts">
/**
 * CaptchaModal 使用示例
 * 在需要验证码弹框的页面中使用
 */
import { ref } from 'vue';
import { Button, message } from 'ant-design-vue';
import { CaptchaModal } from '#/components/captcha';

// 验证码弹框状态
const captchaModalVisible = ref(false);
const captchaType = ref('slider'); // slider/puzzle/rotation/point

// 验证成功后获得的验证数据
const verifiedCaptchaData = ref<{ captchaId: string; captchaCode: string } | null>(null);

// 打开验证码弹框
function openCaptcha() {
  captchaModalVisible.value = true;
}

// 验证成功回调
function onCaptchaSuccess(data: { captchaId: string; captchaCode: string }) {
  verifiedCaptchaData.value = data;
  message.success('验证通过！');

  // 这里可以继续业务逻辑，比如：
  // - 提交表单时带上 captchaId 和 captchaCode
  // - 调用需要验证码保护的 API
  console.log('验证码数据:', data);
}

// 验证失败回调
function onCaptchaFail(msg: string) {
  message.error(msg);
}

// 提交业务表单（示例）
// @ts-ignore - 示例代码，暂时未使用
async function submitForm() {
  if (!verifiedCaptchaData.value) {
    message.warning('请先完成安全验证');
    openCaptcha();
    return;
  }

  // 提交时带上验证码数据
  // @ts-ignore - 示例代码，暂时未使用
  const params = {
    // ...其他表单数据
    captchaId: verifiedCaptchaData.value.captchaId,
    captchaCode: verifiedCaptchaData.value.captchaCode,
  };

  // 调用 API...
}
</script>

<template>
  <div class="p-4">
    <h3 class="mb-4 text-lg font-bold">验证码弹框使用示例</h3>

    <!-- 触发按钮 -->
    <Button type="primary" @click="openCaptcha">
      安全验证
    </Button>

    <!-- 验证码弹框组件 -->
    <CaptchaModal
      v-model:visible="captchaModalVisible"
      :captcha-type="captchaType"
      title="请完成安全验证"
      @success="onCaptchaSuccess"
      @fail="onCaptchaFail"
    />

    <!-- 显示验证状态 -->
    <div v-if="verifiedCaptchaData" class="mt-4 text-sm text-green-600">
      已验证通过，可继续操作
    </div>
  </div>
</template>