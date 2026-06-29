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
 * 题目类型（与后端 /questions/types 接口保持一致）
 */
export type QuestionType =
  | 'single_choice'
  | 'multiple_choice'
  | 'indefinite_choice'
  | 'true_false'
  | 'fill_blank'
  | 'cloze'
  | 'term_explanation'
  | 'short_answer'
  | 'essay_question'
  | 'composition'
  | 'material'
  | 'case_analysis'
  | 'reading'
  | 'matching'
  | 'ordering'
  | 'classification'
  | 'listening'
  | 'speaking'
  | 'video'
  | 'document'
  | 'calculation'
  | 'proof'
  | 'operation'
  | 'programming'
  | 'sql'
  | 'code_reading'
  | 'debugging';

/**
 * 题型标签映射
 */
export const QUESTION_TYPE_LABELS: Record<QuestionType, string> = {
  single_choice: '单选题',
  multiple_choice: '多选题',
  indefinite_choice: '不定项选择',
  true_false: '判断题',
  fill_blank: '填空题',
  cloze: '完形填空',
  term_explanation: '名词解释',
  short_answer: '简答题',
  essay_question: '论述题',
  composition: '作文题',
  material: '材料题',
  case_analysis: '案例分析题',
  reading: '阅读理解题',
  matching: '匹配题',
  ordering: '排序题',
  classification: '分类题',
  listening: '听力题',
  speaking: '口语题',
  video: '视频题',
  document: '文档题',
  calculation: '计算题',
  proof: '证明题',
  operation: '操作题',
  programming: '编程题',
  sql: 'SQL题',
  code_reading: '代码阅读题',
  debugging: '调试改错题',
};

/**
 * 选择题类型
 */
export const CHOICE_TYPES: QuestionType[] = [
  'single_choice',
  'multiple_choice',
  'indefinite_choice',
  'true_false',
];

/**
 * 填空题类型
 */
export const FILL_TYPES: QuestionType[] = ['fill_blank', 'cloze'];

/**
 * 主观文字题类型
 */
export const ESSAY_TYPES: QuestionType[] = [
  'short_answer',
  'essay_question',
  'composition',
  'term_explanation',
];

/**
 * 题目接口
 */
export interface Question {
  id: number;
  title: string;
  questionType: QuestionType;
  difficulty: 1 | 2 | 3;
  stem: { blocks: ContentBlock[] };
  options?: Array<{ label: string; content: { blocks: ContentBlock[] }; isCorrect?: boolean }>;
  correctAnswer?: string;  // 正确答案，如 "A" 或 "AB"
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
