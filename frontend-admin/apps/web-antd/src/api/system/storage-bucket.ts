import { requestClient } from '#/api/request';

export namespace StorageBucketApi {
  export interface StorageBucket {
    id: number;
    name: string;
    driver: 'local' | 'minio' | 'oss' | 'cos';
    endpoint: string;
    bucket: string;
    accessKey: string;
    secretKey: string;
    region: string;
    useSsl: boolean;
    cdnDomain: string;
    pathPrefix: string;
    purpose: string;
    isDefault: boolean;
    status: number;
    description: string;
    createdAt: string;
    updatedAt: string;
  }

  export interface CreateStorageBucket {
    name: string;
    driver: string;
    endpoint?: string;
    bucket?: string;
    accessKey?: string;
    secretKey?: string;
    region?: string;
    useSsl?: boolean;
    cdnDomain?: string;
    pathPrefix?: string;
    purpose?: string;
    isDefault?: boolean;
    status?: number;
    description?: string;
  }
}

/**
 * 获取所有存储桶
 */
export function getAllStorageBucketsApi() {
  return requestClient.get<StorageBucketApi.StorageBucket[]>(
    '/system/storage-buckets',
  );
}

/**
 * 根据ID获取存储桶
 */
export function getStorageBucketByIdApi(id: number) {
  return requestClient.get<StorageBucketApi.StorageBucket>(
    `/system/storage-buckets/${id}`,
  );
}

/**
 * 创建存储桶
 */
export function createStorageBucketApi(
  data: StorageBucketApi.CreateStorageBucket,
) {
  return requestClient.post('/system/storage-buckets', data);
}

/**
 * 更新存储桶
 */
export function updateStorageBucketApi(
  id: number,
  data: StorageBucketApi.CreateStorageBucket,
) {
  return requestClient.put(`/system/storage-buckets/${id}`, data);
}

/**
 * 删除存储桶
 */
export function deleteStorageBucketApi(id: number) {
  return requestClient.delete(`/system/storage-buckets/${id}`);
}

/**
 * 设置默认存储桶
 */
export function setDefaultStorageBucketApi(id: number) {
  return requestClient.put(`/system/storage-buckets/${id}/default`);
}

/**
 * 测试存储桶连接
 */
export function testStorageBucketConnectionApi(id: number) {
  return requestClient.post<{ message: string }>(
    `/system/storage-buckets/${id}/test`,
  );
}

/**
 * 获取已启用的存储驱动列表
 */
export function getEnabledDriversApi() {
  return requestClient.get<
    Array<{ value: string; label: string; icon: string; enabled: boolean }>
  >('/system/storage-buckets/enabled-drivers');
}

/**
 * 按驱动和桶名测试连接（无需先保存）
 */
export function testStorageBucketByDriverApi(data: {
  driver: string;
  bucketName: string;
  region?: string;
}) {
  return requestClient.post<{ message: string }>(
    '/system/storage-buckets/test-by-driver',
    data,
  );
}
