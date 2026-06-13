/**
 * 请求配置类型
 */
export interface RequestConfig {
  url: string;
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  data?: any;
  params?: Record<string, any>;
  headers?: Record<string, string>;
  skipAuth?: boolean;
}

/**
 * 统一响应结构
 */
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

/**
 * 分页响应结构
 */
export interface PageResponse<T = any> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

/**
 * 题目内容块类型
 */
export type BlockType = 'text' | 'image' | 'video' | 'audio' | 'document' | 'formula' | 'table' | 'code' | 'quote' | 'divider';

export interface ContentBlock {
  type: BlockType;
  content?: string;
  fileId?: number;
  url?: string;
}

/**
 * 题目类型
 */
export type QuestionType =
  | 'single_choice'
  | 'multiple_choice'
  | 'indefinite_choice'
  | 'true_false'
  | 'fill_blank'
  | 'short_answer'
  | 'essay'
  | 'material'
  | 'case_analysis'
  | 'reading_comprehension'
  | 'matching'
  | 'sorting'
  | 'classification'
  | 'listening'
  | 'speaking'
  | 'video_question'
  | 'document_question'
  | 'calculation'
  | 'proof'
  | 'operation'
  | 'programming'
  | 'sql'
  | 'code_review'
  | 'debugging';

/**
 * 题目接口
 */
export interface Question {
  id: number;
  title: string;
  questionType: QuestionType;
  difficulty: 1 | 2 | 3;
  stem: { blocks: ContentBlock[] };
  options?: Array<{ label: string; content: { blocks: ContentBlock[] } }>;
  answerVisible: boolean;
  analysisVisible: boolean;
  isFavorited: boolean;
  categoryName?: string;
  knowledgePoints?: string[];
  tags?: string[];
}

/**
 * 通知接口
 */
export interface Notification {
  id: number;
  type: string;
  title: string;
  content: string;
  link: string;
  isRead: boolean;
  senderId: number;
  createdAt: string;
}

/**
 * 收藏接口
 */
export interface Favorite {
  id: number;
  title: string;
  questionType: QuestionType;
  difficulty: number;
  categoryName: string;
  favoritedAt: string;
}

/**
 * 笔记接口
 */
export interface Note {
  id: number;
  questionId: number;
  questionTitle: string;
  content: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * 文件信息接口
 */
export interface FileInfo {
  id: number;
  name: string;
  originalName: string;
  fileType: string;          // image/jpeg, video/mp4, application/pdf, etc.
  fileSize: number;          // bytes
  url: string;               // 访问 URL
  thumbnailUrl?: string;     // 缩略图 URL（图片/视频）
  duration?: number;         // 音视频时长（秒）
  width?: number;
  height?: number;
  folder?: string;           // 所属文件夹
  uploadedAt: string;
  uploadedBy: number;
}

/**
 * 分片上传初始化响应
 */
export interface UploadInitResponse {
  uploadId: string;
  chunkSize: number;
  totalChunks: number;
  uploadedChunks: number[];  // 已上传的分片索引（断点续传）
}

/**
 * 分片上传响应
 */
export interface UploadChunkResponse {
  uploadId: string;
  chunkIndex: number;
  uploaded: boolean;
}

/**
 * 文件分享接口
 */
export interface FileShare {
  id: number;
  fileId: number;
  fileName: string;
  shareCode: string;
  shareUrl: string;
  expireAt: string;
  viewCount: number;
  maxViews?: number;
  password?: string;
  createdAt: string;
  createdBy: number;
}

