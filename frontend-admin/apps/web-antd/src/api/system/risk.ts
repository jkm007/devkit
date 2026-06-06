import { requestClient } from '#/api/request';

export interface RiskScoreItem {
  ip: string;
  score: number;
  updatedAt: string;
  expireAt: string;
}

export interface RiskStats {
  totalCount: number;
  triggerCount: number;
  blockCount: number;
  highRiskCount: number;
  triggerScore: number;
  blockScore: number;
  enabled: boolean;
}

/**
 * 获取风险评分列表
 */
export async function getRiskScores(limit: number = 100): Promise<RiskScoreItem[]> {
  return requestClient.get('/system/risk/scores', { params: { limit } });
}

/**
 * 获取指定 IP 的风险评分
 */
export async function getRiskScoreByIP(ip: string): Promise<RiskScoreItem> {
  return requestClient.get('/system/risk/score', { params: { ip } });
}

/**
 * 获取风险评分统计
 */
export async function getRiskStats(): Promise<RiskStats> {
  return requestClient.get('/system/risk/stats');
}

/**
 * 清除指定 IP 的风险评分
 */
export async function clearRiskScore(ip: string): Promise<void> {
  return requestClient.post('/system/risk/clear', null, { params: { ip } });
}