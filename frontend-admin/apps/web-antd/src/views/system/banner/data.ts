export interface Banner {
  id: number;
  title: string;
  image: string;
  link: string;
  linkType: 'internal' | 'external' | 'none';
  sortOrder: number;
  status: 'enabled' | 'disabled';
  createdAt: string;
  updatedAt: string;
}

export const LINK_TYPE_OPTIONS = [
  { label: '内部链接', value: 'internal' },
  { label: '外部链接', value: 'external' },
  { label: '无链接', value: 'none' },
];

export const STATUS_OPTIONS = [
  { label: '启用', value: 'enabled' },
  { label: '禁用', value: 'disabled' },
];