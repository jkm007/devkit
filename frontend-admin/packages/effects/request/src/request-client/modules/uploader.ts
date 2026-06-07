import type { RequestClient } from '../request-client';
import type { RequestClientConfig } from '../types';

import { isUndefined } from '@vben/utils';

export interface UploadProgressEvent {
  loaded: number;
  total: number;
  percent: number;
}

export type UploadProgressCallback = (event: UploadProgressEvent) => void;

class FileUploader {
  private client: RequestClient;

  constructor(client: RequestClient) {
    this.client = client;
  }

  public async upload<T = any>(
    url: string,
    data: Record<string, any> & { file: Blob | File },
    config?: RequestClientConfig & { onProgress?: UploadProgressCallback },
  ): Promise<T> {
    const formData = new FormData();

    Object.entries(data).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        value.forEach((item, index) => {
          !isUndefined(item) && formData.append(`${key}[${index}]`, item);
        });
      } else {
        !isUndefined(value) && formData.append(key, value);
      }
    });

    const { onProgress, ...restConfig } = config || {};

    // 如果有进度回调，使用 XMLHttpRequest
    if (onProgress) {
      return this.uploadWithProgress<T>(url, formData, onProgress, restConfig);
    }

    const finalConfig: RequestClientConfig = {
      ...restConfig,
      headers: {
        'Content-Type': 'multipart/form-data',
        ...restConfig?.headers,
      },
    };

    return this.client.post(url, formData, finalConfig);
  }

  private uploadWithProgress<T>(
    url: string,
    formData: FormData,
    onProgress: UploadProgressCallback,
    config?: RequestClientConfig,
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();

      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
          onProgress({
            loaded: event.loaded,
            total: event.total,
            percent: Math.round((event.loaded / event.total) * 100),
          });
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const response = JSON.parse(xhr.responseText);
            resolve(response as T);
          } catch {
            resolve(xhr.responseText as T);
          }
        } else {
          reject(new Error(`Upload failed: ${xhr.status}`));
        }
      });

      xhr.addEventListener('error', () => {
        reject(new Error('Upload failed'));
      });

      xhr.addEventListener('abort', () => {
        reject(new Error('Upload aborted'));
      });

      // 构建完整 URL
      const baseURL = this.client.getBaseUrl() || '';
      const fullURL = url.startsWith('http') ? url : `${baseURL}${url}`;

      xhr.open('POST', fullURL);

      // 设置 Authorization header
      const token = this.client.instance.defaults.headers.common?.Authorization;
      if (token) {
        xhr.setRequestHeader('Authorization', token as string);
      }

      // 设置其他 headers
      if (config?.headers) {
        Object.entries(config.headers).forEach(([key, value]) => {
          if (key !== 'Content-Type') {
            xhr.setRequestHeader(key, value as string);
          }
        });
      }

      xhr.send(formData);
    });
  }
}

export { FileUploader };
