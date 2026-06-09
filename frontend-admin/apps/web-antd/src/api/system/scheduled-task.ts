import { requestClient } from '#/api/request';

/** 定时任务 */
export interface ScheduledTask {
  id: number;
  name: string;
  taskType: string;
  cronExpr: string;
  config: Record<string, any>;
  enabled: boolean;
  status: string;
  lastRunAt?: string;
  lastResult?: string;
  nextRunAt?: string;
  runCount: number;
  createdAt: string;
  updatedAt: string;
}

/** 获取所有定时任务 */
export function getScheduledTasks() {
  return requestClient.get<ScheduledTask[]>('/system/scheduled-tasks');
}

/** 创建定时任务 */
export function createScheduledTask(data: {
  name: string;
  taskType: string;
  cronExpr: string;
  config?: Record<string, any>;
}) {
  return requestClient.post<ScheduledTask>('/system/scheduled-tasks', data);
}

/** 获取任务详情 */
export function getScheduledTaskById(id: number) {
  return requestClient.get<ScheduledTask>(`/system/scheduled-tasks/${id}`);
}

/** 更新任务 */
export function updateScheduledTask(id: number, data: {
  name: string;
  cronExpr: string;
  config?: Record<string, any>;
  enabled?: boolean;
}) {
  return requestClient.put(`/system/scheduled-tasks/${id}`, data);
}

/** 更新任务启用状态 */
export function updateScheduledTaskEnabled(id: number, enabled: boolean) {
  return requestClient.put(`/system/scheduled-tasks/${id}/enabled`, { enabled });
}

/** 删除任务 */
export function deleteScheduledTask(id: number) {
  return requestClient.delete(`/system/scheduled-tasks/${id}`);
}

/** 手动执行任务 */
export function runScheduledTask(id: number) {
  return requestClient.post(`/system/scheduled-tasks/${id}/run`);
}
