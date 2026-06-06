/**
 * 通用表单验证规则
 * 用于登录、注册、找回密码等认证页面
 */

/**
 * 验证邮箱格式
 */
export function validateEmail(email: string): boolean {
  if (!email) return false;
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

/**
 * 验证手机号格式（中国大陆）
 */
export function validatePhone(phone: string): boolean {
  if (!phone) return false;
  return /^1\d{10}$/.test(phone);
}

/**
 * 验证用户名（至少3个字符）
 */
export function validateUsername(username: string): boolean {
  if (!username) return false;
  return username.length >= 3;
}

/**
 * 验证密码（至少6个字符）
 */
export function validatePassword(password: string): boolean {
  if (!password) return false;
  return password.length >= 6;
}

/**
 * 验证验证码（6位数字）
 */
export function validateCode(code: string): boolean {
  if (!code) return false;
  return code.length === 6 && /^\d+$/.test(code);
}

/**
 * 验证两次密码是否一致
 */
export function validatePasswordMatch(
  password: string,
  confirmPassword: string,
): boolean {
  return password === confirmPassword;
}

/**
 * 表单验证错误消息
 */
export const VALIDATION_MESSAGES = {
  emailRequired: '请输入邮箱',
  emailInvalid: '请输入正确的邮箱格式',
  phoneRequired: '请输入手机号',
  phoneInvalid: '请输入正确的手机号',
  usernameRequired: '请输入用户名',
  usernameMinLength: '用户名至少3个字符',
  passwordRequired: '请输入密码',
  passwordMinLength: '密码至少6个字符',
  codeRequired: '请输入验证码',
  codeInvalid: '请输入6位验证码',
  passwordMismatch: '两次输入的密码不一致',
};