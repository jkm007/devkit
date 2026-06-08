import { requestClient } from '#/api/request';

export namespace StorageConfigApi {
  export interface StorageConfig {
    id: number;
    name: string;
    driver: string;
    endpoint: string;
    accessKey: string;
    secretKey: string;
    bucket: string;
    region: string;
    useSsl: boolean;
    cdnDomain: string;
    isDefault: boolean;
    presignedUrlExpiry: number;
    status: number;
    description: string;
    createdAt: string;
    updatedAt: string;
  }

  export interface CreateStorageConfig {
    name: string;
    driver: string;
    endpoint?: string;
    accessKey?: string;
    secretKey?: string;
    bucket?: string;
    region?: string;
    useSsl?: boolean;
    cdnDomain?: string;
    isDefault?: boolean;
    presignedUrlExpiry?: number;
    status?: number;
    description?: string;
  }

  export interface TestConnectionData {
    driver: string;
    endpoint?: string;
    accessKey?: string;
    secretKey?: string;
    bucket?: string;
    region?: string;
    useSsl?: boolean;
  }
}

/** 获取所有存储配置 */
export async function getAllStorageConfigsApi() {
  return requestClient.get<StorageConfigApi.StorageConfig[]>(
    '/system/storage-configs',
  );
}

/** 根据ID获取存储配置 */
export async function getStorageConfigByIdApi(id: number) {
  return requestClient.get<StorageConfigApi.StorageConfig>(
    `/system/storage-configs/${id}`,
  );
}

/** 创建存储配置 */
export async function createStorageConfigApi(
  data: StorageConfigApi.CreateStorageConfig,
) {
  return requestClient.post<StorageConfigApi.StorageConfig>(
    '/system/storage-configs',
    data,
  );
}

/** 更新存储配置 */
export async function updateStorageConfigApi(
  id: number,
  data: StorageConfigApi.CreateStorageConfig,
) {
  return requestClient.put<StorageConfigApi.StorageConfig>(
    `/system/storage-configs/${id}`,
    data,
  );
}

/** 删除存储配置 */
export async function deleteStorageConfigApi(id: number) {
  return requestClient.delete(`/system/storage-configs/${id}`);
}

/** 设置默认存储配置 */
export async function setDefaultStorageConfigApi(id: number) {
  return requestClient.put(`/system/storage-configs/${id}/default`);
}

/** 测试已有配置的连接 */
export async function testStorageConfigConnectionApi(id: number) {
  return requestClient.post<{ connected: boolean }>(
    `/system/storage-configs/${id}/test`,
  );
}

/** 根据数据测试连接 */
export async function testStorageConfigByDataApi(
  data: StorageConfigApi.TestConnectionData,
) {
  return requestClient.post<{ connected: boolean }>(
    '/system/storage-configs/test-by-data',
    data,
  );
}

/** 获取已启用的驱动列表 */
export async function getStorageConfigEnabledDriversApi() {
  return requestClient.get<
    Array<{ value: string; label: string; icon: string; enabled: boolean }>
  >('/system/storage-configs/enabled-drivers');
}
