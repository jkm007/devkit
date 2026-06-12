<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';

import {
  computed,
  defineAsyncComponent,
  markRaw,
  onMounted,
  ref,
  watch,
} from 'vue';

import { AuthenticationLogin, VbenIconButton, z } from '@vben/common-ui';
import { SvgGithubIcon, SvgGoogleIcon, SvgWeChatIcon } from '@vben/icons';
import { $t } from '@vben/locales';
import { message } from 'ant-design-vue';

import { getOAuthUrl } from '#/api/core/auth';
import { getCaptcha, getPublicSettings } from '#/api/system/settings';
import { useAuthStore } from '#/store';

// 动态导入验证码组件，避免影响其他 authentication 页面
const CaptchaModal = defineAsyncComponent(
  () => import('#/components/captcha/captcha-modal.vue'),
);
const NumericCaptcha = defineAsyncComponent(() =>
  import('@vben/common-ui').then((m) => m.NumericCaptcha),
);

defineOptions({ name: 'Login' });

const authStore = useAuthStore();

// ==================== 登录方式设置 ====================
const loginEmailEnabled = ref(false);
const loginPhoneEnabled = ref(false);
const loginOauthEnabled = ref(false);
const loginOauthProviders = ref<string[]>([]);

// ==================== 验证码状态 ====================
const captchaEnabled = ref(false);
const captchaType = ref('slider');
const captchaLoginTrigger = ref(999);
const loginFailCount = ref(0);
const settingsLoaded = ref(false);

// 弹框验证码状态
const captchaModalVisible = ref(false);
const captchaVerified = ref(false);
const captchaResult = ref<{
  captchaId: string;
  captchaCode: string;
  startTime?: number;
} | null>(null);
const captchaVerifyResult = ref<boolean | null>(null); // 验证结果
const captchaVerifyMessage = ref(''); // 验证失败消息

// 待登录参数（验证成功后继续登录）
const pendingLoginParams = ref<Record<string, any> | null>(null);

// 数字验证码数据（嵌入表单）
const numericCaptchaImage = ref('');
const numericCaptchaId = ref('');
const numericCaptchaLength = ref(4);
const numericCaptchaLoading = ref(false);

// ==================== 计算属性 ====================
// 是否需要显示验证码
const shouldShowCaptcha = computed(() => {
  if (!settingsLoaded.value) return false;
  if (!captchaEnabled.value) return false;
  if (captchaLoginTrigger.value === 0) return true;
  return loginFailCount.value >= captchaLoginTrigger.value;
});

// 验证码类型判断
const isNumericCaptcha = computed(() => captchaType.value === 'numeric');

// ==================== 加载逻辑 ====================
async function loadSettings() {
  try {
    const settings = await getPublicSettings();
    // 加载验证码设置
    if (settings?.captcha) {
      captchaEnabled.value =
        settings.captcha.captcha_enabled === true ||
        settings.captcha.captcha_enabled === 'true';
      if (settings.captcha.captcha_type) {
        captchaType.value = String(settings.captcha.captcha_type).replace(
          /"/g,
          '',
        );
      }
      if (settings.captcha.captcha_login_trigger !== undefined) {
        captchaLoginTrigger.value = Number(
          settings.captcha.captcha_login_trigger,
        );
      }
    }
    // 加载登录方式设置
    if (settings?.auth) {
      loginEmailEnabled.value =
        settings.auth.login_email_enabled === true ||
        settings.auth.login_email_enabled === 'true';
      loginPhoneEnabled.value =
        settings.auth.login_phone_enabled === true ||
        settings.auth.login_phone_enabled === 'true';
      loginOauthEnabled.value =
        settings.auth.login_oauth_enabled === true ||
        settings.auth.login_oauth_enabled === 'true';
      if (settings.auth.login_oauth_providers) {
        loginOauthProviders.value = Array.isArray(
          settings.auth.login_oauth_providers,
        )
          ? settings.auth.login_oauth_providers
          : [];
      }
    }
  } catch (e) {
    console.error('[Login] 加载设置失败:', e);
  } finally {
    settingsLoaded.value = true;
  }
}

// 获取数字验证码
async function fetchNumericCaptcha() {
  if (numericCaptchaLoading.value) return;
  numericCaptchaLoading.value = true;

  try {
    const data = await getCaptcha('numeric');
    if (data && data.captcha_id && data.image) {
      numericCaptchaImage.value = data.image;
      numericCaptchaId.value = data.captcha_id;
      numericCaptchaLength.value = data.length || 4;
      captchaVerified.value = false;
      captchaResult.value = null;
    }
  } catch (e) {
    console.error('[Login] 获取数字验证码失败:', e);
  } finally {
    numericCaptchaLoading.value = false;
  }
}

// 弹框验证完成（收集到验证码数据，准备提交登录）
function onCaptchaModalSuccess(data: {
  captchaId: string;
  captchaCode: string;
  startTime?: number;
}) {
  captchaVerified.value = true;
  captchaResult.value = data;

  // 如果有待登录参数，自动继续登录
  if (pendingLoginParams.value) {
    doLogin(pendingLoginParams.value, data);
    pendingLoginParams.value = null;
  }
}

// 弹框验证失败
function onCaptchaModalFail(_msg: string) {
  captchaVerified.value = false;
  captchaResult.value = null;
}

// 数字验证码成功
function handleNumericCaptchaSuccess(data: {
  captchaId: string;
  code: string;
}) {
  captchaVerified.value = true;
  captchaResult.value = { captchaId: data.captchaId, captchaCode: data.code };
}

// ==================== 生命周期 ====================
onMounted(async () => {
  await loadSettings();
  // 如果需要数字验证码，加载图片
  if (shouldShowCaptcha.value && isNumericCaptcha.value) {
    await fetchNumericCaptcha();
  }
});

watch(shouldShowCaptcha, async (need) => {
  if (need) {
    if (isNumericCaptcha.value && !numericCaptchaImage.value) {
      await fetchNumericCaptcha();
    }
  } else {
    captchaVerified.value = false;
    captchaResult.value = null;
    numericCaptchaImage.value = '';
    numericCaptchaId.value = '';
  }
});

// ==================== 表单 ====================
const formSchema = computed((): VbenFormSchema[] => {
  const schema: VbenFormSchema[] = [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: $t('authentication.usernameTip'),
      },
      fieldName: 'username',
      label: $t('authentication.username'),
      rules: z.string().min(1, { message: $t('authentication.usernameTip') }),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        placeholder: $t('authentication.password'),
      },
      fieldName: 'password',
      label: $t('authentication.password'),
      rules: z.string().min(1, { message: $t('authentication.passwordTip') }),
    },
  ];

  // 验证码需要显示时添加（数字验证码嵌入表单，其他类型弹框处理）
  if (shouldShowCaptcha.value) {
    if (isNumericCaptcha.value && numericCaptchaImage.value) {
      // 数字验证码嵌入表单
      schema.push({
        component: markRaw(NumericCaptcha),
        componentProps: {
          serverImage: numericCaptchaImage.value,
          serverCaptchaId: numericCaptchaId.value,
          charLength: numericCaptchaLength.value,
          onSuccess: handleNumericCaptchaSuccess,
        },
        fieldName: 'captcha',
        rules: z.boolean().refine((value) => value, {
          message: $t('authentication.verifyRequiredTip'),
        }),
      });
    }
    // 其他验证码类型（slider/puzzle/rotation/point）通过弹框处理，不在表单中显示
  }

  return schema;
});

// ==================== 辅助函数 ====================
// 重置验证码相关状态
function resetCaptchaState() {
  captchaVerified.value = false;
  captchaResult.value = null;
  pendingLoginParams.value = null;
}

// ==================== 登录 ====================
// 执行登录（带验证码数据）
function doLogin(
  params: Record<string, any>,
  captchaData?: { captchaCode: string; captchaId: string; startTime?: number },
) {
  const loginParams: Record<string, any> = {
    username: params.username,
    password: params.password,
  };

  if (captchaData) {
    loginParams.captchaId = captchaData.captchaId;
    loginParams.captchaCode = captchaData.captchaCode;
    loginParams.captchaType = captchaType.value;
    loginParams.startTime = captchaData.startTime || 0;
    if (captchaType.value === 'point' && captchaData.captchaCode) {
      loginParams.points = JSON.parse(captchaData.captchaCode);
    }
  } else if (captchaResult.value) {
    loginParams.captchaId = captchaResult.value.captchaId;
    loginParams.captchaCode = captchaResult.value.captchaCode;
    loginParams.captchaType = captchaType.value;
    loginParams.startTime = captchaResult.value.startTime || 0;
    if (captchaType.value === 'point' && captchaResult.value.captchaCode) {
      loginParams.points = JSON.parse(captchaResult.value.captchaCode);
    }
  }

  authStore.authLogin(loginParams).then((result: any) => {
    if (!result?.userInfo) {
      loginFailCount.value++;
      resetCaptchaState();

      // 关闭弹窗，下次重新打开
      captchaModalVisible.value = false;

      // 数字验证码需要重新获取
      if (isNumericCaptcha.value && shouldShowCaptcha.value) {
        fetchNumericCaptcha();
      }
    } else {
      // 登录成功，关闭弹窗
      captchaModalVisible.value = false;
      resetCaptchaState();
    }
  }).catch((error: any) => {
    // 403001 验证码错误：关闭弹窗，显示错误，下次重新打开
    const responseData = error?.data || error?.response?.data;
    if (responseData?.code === 403001) {
      message.error(responseData?.message || '验证码错误，请重试');
    }
    // 所有错误都关闭弹窗并重置状态
    captchaModalVisible.value = false;
    loginFailCount.value++;
    resetCaptchaState();
  });
}

// 表单提交处理
function handleSubmit(params: Record<string, any>) {
  // 如果不需要验证码，直接登录
  if (!shouldShowCaptcha.value) {
    doLogin(params);
    return;
  }

  // 数字验证码：表单内验证，直接登录
  if (isNumericCaptcha.value) {
    if (!captchaVerified.value) {
      message.warning('请输入验证码');
      return;
    }
    doLogin(params);
    return;
  }

  // 弹框验证码（slider/puzzle/rotation/point）
  if (captchaVerified.value && captchaResult.value) {
    // 已验证，直接登录
    doLogin(params);
  } else {
    // 未验证，弹出验证框，保存登录参数待验证成功后继续
    pendingLoginParams.value = params;
    captchaModalVisible.value = true;
  }
}

// ==================== 第三方登录 ====================
async function handleOAuthLogin(provider: string) {
  try {
    const { url } = await getOAuthUrl(provider);
    // 跳转到第三方授权页面
    window.location.href = url;
  } catch (e: any) {
    message.error(e?.message || '获取授权链接失败');
  }
}
</script>

<template>
  <AuthenticationLogin
    :form-schema="formSchema"
    :loading="authStore.loginLoading"
    :show-code-login="loginEmailEnabled || loginPhoneEnabled"
    :show-third-party-login="
      loginOauthEnabled && loginOauthProviders.length > 0
    "
    @submit="handleSubmit"
  >
    <!-- 自定义第三方登录（绑定实际 OAuth 跳转） -->
    <template #third-party-login>
      <div
        v-if="loginOauthEnabled && loginOauthProviders.length > 0"
        class="w-full sm:mx-auto md:max-w-md"
      >
        <div class="mt-4 flex items-center justify-between">
          <span
            class="w-[35%] border-b border-input dark:border-gray-600"
          ></span>
          <span class="text-center text-xs text-muted-foreground uppercase">
            其他登录方式
          </span>
          <span
            class="w-[35%] border-b border-input dark:border-gray-600"
          ></span>
        </div>
        <div class="mt-4 flex flex-wrap justify-center gap-4">
          <VbenIconButton
            v-if="loginOauthProviders.includes('wechat')"
            tooltip="微信登录"
            tooltip-side="top"
            @click="handleOAuthLogin('wechat')"
          >
            <SvgWeChatIcon />
          </VbenIconButton>
          <VbenIconButton
            v-if="loginOauthProviders.includes('github')"
            tooltip="GitHub 登录"
            tooltip-side="top"
            @click="handleOAuthLogin('github')"
          >
            <SvgGithubIcon />
          </VbenIconButton>
          <VbenIconButton
            v-if="loginOauthProviders.includes('google')"
            tooltip="Google 登录"
            tooltip-side="top"
            @click="handleOAuthLogin('google')"
          >
            <SvgGoogleIcon />
          </VbenIconButton>
        </div>
      </div>
    </template>
  </AuthenticationLogin>

  <!-- 验证码弹框（公开模式，无需认证） -->
  <CaptchaModal
    v-model:visible="captchaModalVisible"
    :captcha-type="captchaType"
    :public="true"
    :verify-result="captchaVerifyResult"
    :verify-message="captchaVerifyMessage"
    title="请完成安全验证"
    @success="onCaptchaModalSuccess"
    @fail="onCaptchaModalFail"
  />
</template>
