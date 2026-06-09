import { requestClient } from '#/api/request';

export namespace RateLimitRuleApi {
  export interface Rule {
    id: number;
    pathPattern: string;
    method: string;
    rate: number;
    burst: number;
    cooldown: number;
    blockDuration: number;
    maxViolations: number;
    violationScore: number;
    description: string;
    enabled: boolean;
    priority: number;
    createdAt: string;
    updatedAt: string;
  }

  export interface RuleForm {
    pathPattern: string;
    method: string;
    rate: number;
    burst: number;
    cooldown: number;
    blockDuration: number;
    maxViolations: number;
    violationScore: number;
    description: string;
    enabled: boolean;
    priority: number;
  }
}

/** 获取所有限流规则 */
export function getRateLimitRules() {
  return requestClient.get<RateLimitRuleApi.Rule[]>(
    '/system/rate-limit-rules',
  );
}

/** 获取单个规则 */
export function getRateLimitRule(id: number) {
  return requestClient.get<RateLimitRuleApi.Rule>(
    `/system/rate-limit-rules/${id}`,
  );
}

/** 创建规则 */
export function createRateLimitRule(data: RateLimitRuleApi.RuleForm) {
  return requestClient.post<RateLimitRuleApi.Rule>(
    '/system/rate-limit-rules',
    data,
  );
}

/** 更新规则 */
export function updateRateLimitRule(
  id: number,
  data: RateLimitRuleApi.RuleForm,
) {
  return requestClient.put<RateLimitRuleApi.Rule>(
    `/system/rate-limit-rules/${id}`,
    data,
  );
}

/** 删除规则 */
export function deleteRateLimitRule(id: number) {
  return requestClient.delete(`/system/rate-limit-rules/${id}`);
}

/** 更新规则启用状态 */
export function updateRateLimitRuleStatus(id: number, enabled: boolean) {
  return requestClient.put(`/system/rate-limit-rules/${id}/status`, {
    enabled,
  });
}
