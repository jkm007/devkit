# 轮播图URL永久化方案

## 问题分析

当前轮播图数据库直接存储MinIO预签名URL，有效期7天，过期后图片无法显示。

## 方案设计：存储文件ID，动态获取预签名URL

### 1. 数据库改造

```sql
-- 添加文件ID字段
ALTER TABLE banners ADD COLUMN file_id INT NULL COMMENT '文件ID（关联file_entries表）';

-- 更新现有数据（需要迁移）
-- 将现有MinIO URL对应的文件ID提取出来
```

### 2. 后端接口改造

#### 2.1 新增公开接口（无需认证）

```go
// 路由：GET /api/v1/files/{id}/public-url
// 用途：轮播图等公开资源获取预签名URL
// 特点：无需认证，带缓存

func (h *MediaHandler) GetPublicURL(c *gin.Context) {
    id := c.Param("id")
    
    // 1. 先查Redis缓存
    cacheKey := fmt.Sprintf("file:public-url:%s", id)
    cached, _ := redis.Get(ctx, cacheKey).Result()
    if cached != "" {
        c.JSON(200, gin.H{"url": cached})
        return
    }
    
    // 2. 查数据库获取文件信息
    asset := fileService.GetAssetByFileID(id)
    
    // 3. 生成预签名URL
    var url string
    if asset.StorageType == "local" {
        url = fmt.Sprintf("/files/%s/view", id)  // 本地存储直接代理
    } else {
        url = storage.GetPresignedURL(asset.ObjectKey, 7*24*3600)  // 7天
    }
    
    // 4. 缓存到Redis（5天，比URL有效期提前2天过期）
    redis.Set(ctx, cacheKey, url, 5*24*time.Hour)
    
    c.JSON(200, gin.H{"url": url})
}
```

#### 2.2 修改轮播图接口

```go
// 轮播图列表接口返回file_id
type BannerResponse struct {
    ID       uint   `json:"id"`
    Title    string `json:"title"`
    FileID   uint   `json:"fileId"`   // 新增
    Image    string `json:"image"`    // 保留兼容，但不再是预签名URL
    Link     string `json:"link"`
    LinkType string `json:"linkType"`
}

// 启动时或定时任务：刷新即将过期的缓存
func RefreshPublicURLCache() {
    // 扫描所有 file:public-url:* 键
    // 对TTL<2天的键重新生成预签名URL
}
```

### 3. 前端改造

#### 3.1 轮播图组件

```typescript
// api/banner.ts
export interface BannerItem {
  id: number;
  title: string;
  fileId: number;   // 新增
  image: string;    // 备用，可能是相对路径
  link: string;
  linkType: string;
}

// 获取公开文件URL（无需认证）
export function getPublicFileURL(fileId: number) {
  return request.get<{ url: string }>(`/files/${fileId}/public-url`);
}

// 批量获取URL
export function batchGetPublicURLs(fileIds: number[]) {
  return request.post<{ urls: Record<number, string> }>('/files/batch-public-url', { fileIds });
}
```

#### 3.2 首页轮播图渲染

```vue
<script setup>
import { getBanners, getPublicFileURL } from '@/api/banner';

const banners = ref([]);

onMounted(async () => {
  const bannerList = await getBanners();
  
  // 并行获取所有轮播图的URL
  const urlPromises = bannerList.map(async (banner) => {
    if (banner.fileId) {
      const { url } = await getPublicFileURL(banner.fileId);
      banner.resolvedUrl = url;
    } else {
      banner.resolvedUrl = banner.image;  // 兼容旧数据
    }
    return banner;
  });
  
  banners.value = await Promise.all(urlPromises);
});
</script>

<template>
  <swiper>
    <swiper-item v-for="banner in banners" :key="banner.id">
      <image :src="banner.resolvedUrl" />
    </swiper-item>
  </swiper>
</template>
```

### 4. 缓存策略

```
┌─────────────────────────────────────────────────────────┐
│                    缓存层次设计                          │
├─────────────────────────────────────────────────────────┤
│  Layer 1: 浏览器缓存                                     │
│  - HTTP Cache-Control: max-age=3600 (1小时)             │
│  - 减少重复请求                                          │
├─────────────────────────────────────────────────────────┤
│  Layer 2: Redis缓存                                      │
│  - Key: file:public-url:{fileId}                        │
│  - TTL: 5天 (预签名URL 7天，提前2天刷新)                 │
│  - 定时任务扫描即将过期的缓存并刷新                        │
├─────────────────────────────────────────────────────────┤
│  Layer 3: MinIO预签名URL                                 │
│  - 有效期: 7天                                           │
│  - 由后端生成，前端不直接访问                              │
└─────────────────────────────────────────────────────────┘
```

### 5. 定时刷新任务

```go
// 每天凌晨3点执行
func (s *ScheduledTaskService) RefreshFileURLs() {
    // 1. 扫描所有公开资源（轮播图、快捷菜单等）
    publicFiles := getAllPublicFileIDs()
    
    // 2. 为每个文件生成新的预签名URL
    for _, fileID := range publicFiles {
        url := generatePresignedURL(fileID, 7*24*3600)
        redis.Set(ctx, fmt.Sprintf("file:public-url:%d", fileID), url, 5*24*time.Hour)
    }
    
    // 3. 清理过期缓存
    cleanExpiredCache()
}
```

### 6. 迁移步骤

1. **阶段1：添加file_id字段**
   - 修改Banner model
   - 运行数据库迁移
   
2. **阶段2：实现公开URL接口**
   - 新增 `/files/{id}/public-url` 接口
   - 实现Redis缓存
   
3. **阶段3：前端改造**
   - 修改轮播图组件使用fileId
   - 动态获取URL
   
4. **阶段4：数据迁移**
   - 将现有MinIO URL对应的文件ID提取出来
   - 更新banners表的file_id字段

5. **阶段5：定时任务**
   - 添加URL刷新定时任务
   - 监控缓存命中率

### 7. 兼容性处理

```go
// 后端兼容：同时返回fileId和image
func (h *BannerHandler) GetBanners(c *gin.Context) {
    banners := h.bannerService.ListEnabled()
    
    // 为每个banner生成当前可用的URL
    for i := range banners {
        if banners[i].FileID > 0 {
            // 有fileID，从缓存获取URL
            banners[i].Image = getPublicURL(banners[i].FileID)
        }
        // 否则保持原有image字段（可能是旧的预签名URL）
    }
    
    response.Success(c, banners)
}
```

## 优势

1. **安全性**：bucket保持私有，不暴露其他资源
2. **永久性**：URL动态生成，永不过期
3. **性能**：Redis缓存，减少MinIO压力
4. **兼容性**：支持旧数据平滑迁移
