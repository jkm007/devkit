import { message } from 'ant-design-vue';

/**
 * 根据文件 MIME 类型返回对应的图标类名
 */
export function getFileIcon(type: string | undefined): string {
  if (!type) return 'i-ant-design:file-outlined';
  if (type.startsWith('image/')) return 'i-ant-design:file-image-outlined';
  if (type.startsWith('video/')) return 'i-ant-design:file-video-outlined';
  if (type.startsWith('audio/')) return 'i-ant-design:sound-outlined';
  if (type.includes('pdf')) return 'i-ant-design:file-pdf-outlined';
  if (type.includes('word') || type.includes('document')) return 'i-ant-design:file-word-outlined';
  if (type.includes('excel') || type.includes('spreadsheet')) return 'i-ant-design:file-excel-outlined';
  if (type.includes('zip') || type.includes('rar')) return 'i-ant-design:file-zip-outlined';
  return 'i-ant-design:file-outlined';
}

/**
 * 格式化文件大小为人类可读字符串
 */
export function formatFileSize(size: number | undefined | null): string {
  if (size === undefined || size === null || isNaN(size)) return '-';
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`;
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`;
}

/**
 * 复制文本到剪贴板，支持降级方案
 */
export function fallbackCopy(text: string, successMsg = '链接已复制到剪贴板'): void {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(text).then(
      () => message.success(successMsg),
      () => {
        // 降级方案
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.cssText = 'position:fixed;left:0;top:0;opacity:0';
        document.body.appendChild(textarea);
        textarea.focus();
        textarea.select();
        try {
          document.execCommand('copy');
          message.success(successMsg);
        } catch {
          message.error('复制失败，请手动复制');
        }
        document.body.removeChild(textarea);
      },
    );
  } else {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.cssText = 'position:fixed;left:0;top:0;opacity:0';
    document.body.appendChild(textarea);
    textarea.focus();
    textarea.select();
    try {
      document.execCommand('copy');
      message.success(successMsg);
    } catch {
      message.error('复制失败，请手动复制');
    }
    document.body.removeChild(textarea);
  }
}
