import { requestClient } from '#/api/request';

export namespace SystemGroupApi {
  export interface SystemGroup {
    [key: string]: any;
    children?: SystemGroup[];
    id: number;
    name: string;
    pid?: number;
    roleIds?: number[];
    remark?: string;
    status: 0 | 1;
  }
}

/**
 * 获取分组列表数据
 */
async function getGroupList() {
  return requestClient.get<Array<SystemGroupApi.SystemGroup>>(
    '/system/group/list',
  );
}

/**
 * 创建分组
 * @param data 分组数据
 */
async function createGroup(
  data: Omit<SystemGroupApi.SystemGroup, 'children' | 'id'>,
) {
  return requestClient.post('/system/group', data);
}

/**
 * 更新分组
 *
 * @param id 分组 ID
 * @param data 分组数据
 */
async function updateGroup(
  id: number,
  data: Omit<SystemGroupApi.SystemGroup, 'children' | 'id'>,
) {
  return requestClient.put(`/system/group/${id}`, data);
}

/**
 * 删除分组
 * @param id 分组 ID
 */
async function deleteGroup(id: number) {
  return requestClient.delete(`/system/group/${id}`);
}

export { createGroup, deleteGroup, getGroupList, updateGroup };
