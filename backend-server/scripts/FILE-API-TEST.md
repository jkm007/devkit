# 文件管理功能测试指南

## 测试环境

- 前端地址: http://localhost:5667/
- 后端地址: http://localhost:8080/
- Swagger 文档: http://localhost:8080/swagger/index.html

## 测试步骤

### 1. 登录系统

访问 http://localhost:5667/，使用账号登录：
- 用户名: admin
- 密码: admin123
- 验证码: 从界面获取

### 2. 测试文件管理菜单

登录后，在左侧菜单找到「文件管理」：
- 菜单图标: 📁 folder
- 菜单名称: 文件管理 > 文件列表

### 3. 测试文件夹操作

在文件管理页面左侧：

#### 创建文件夹
1. 点击「新建文件夹」按钮
2. 输入文件夹名称
3. 选择父文件夹（可选）
4. 点击确认创建

#### 重命名文件夹
1. 鼠标悬停在文件夹上
2. 点击更多操作按钮(...)
3. 选择「重命名」
4. 输入新名称，确认

#### 删除文件夹
1. 鼠标悬停在文件夹上
2. 点击更多操作按钮(...)
3. 选择「删除」
4. 确认删除（注意：会删除文件夹内所有文件）

### 4. 测试文件操作

在文件管理页面右侧：

#### 上传文件
1. 点击「上传文件」按钮
2. 选择文件（支持多文件）
3. 查看上传进度

**秒传测试**: 上传相同文件第二次，应立即完成（秒传）

#### 文件列表
- 查看文件名称、大小、类型、创建时间
- 点击文件夹筛选文件

#### 预览文件
1. 点击文件行的「眼睛」图标
2. 图片文件：显示图片预览 Modal
3. 视频文件：显示视频播放器（开发中）
4. 其他文件：提示不支持预览

#### 移动文件
1. 点击文件行的「文件夹」图标
2. 选择目标文件夹
3. 确认移动

#### 删除文件
1. 点击文件行的「删除」图标
2. 确认删除

### 5. 测试头像上传

#### 进入个人设置
1. 点击右上角用户头像
2. 选择「个人设置」

#### 更换头像
1. 在个人设置页面点击头像区域
2. 弹出头像上传 Modal
3. 选择图片文件
4. 调整裁剪区域（1:1 比例）
5. 确认裁剪并上传
6. 查看头像是否更新

### 6. API 测试（通过 Swagger）

访问 http://localhost:8080/swagger/index.html

#### 获取 Token
1. 先在前端登录
2. 在浏览器开发者工具获取 Authorization header 的 token
3. 点击 Swagger 页面的「Authorize」按钮
4. 输入 token（格式：Bearer <token>）

#### 测试接口

**文件夹管理**:
- POST /files/folder - 创建文件夹
- GET /files/tree - 获取文件夹树
- PUT /files/folder/:id - 重命名文件夹
- DELETE /files/folder/:id - 删除文件夹

**文件管理**:
- GET /files/list - 获取文件列表
- POST /files/move - 移动文件
- DELETE /files/:id - 删除文件

**分片上传**:
- POST /files/upload/check - 检查上传状态（秒传检测）
- POST /files/upload/init - 初始化上传
- POST /files/upload/part - 上传分片
- POST /files/upload/complete - 完成上传
- POST /files/upload/abort - 取消上传
- GET /files/upload/status - 获取上传进度

**媒体文件**:
- GET /files/:id/metadata - 获取媒体信息
- GET /files/:id/stream - 获取视频流
- GET /files/:id/download - 下载文件

## 测试要点

### 秒传功能
1. 上传一个新文件
2. 再次上传相同文件
3. 第二次应立即完成（显示「秒传」提示）

### 分片上传（大文件）
1. 上传超过 5MB 的文件
2. 系统自动分片上传
3. 可暂停/继续上传（前端开发中）

### 权限验证
1. 未登录访问 API 应返回 401
2. 登录后可正常访问

### 文件类型图标
- 图片: 🖼️ file-image
- 视频: 🎬 file-video
- 音频: 🎵 sound
- PDF: 📄 file-pdf
- Word: 📝 file-word
- Excel: 📊 file-excel
- 压缩包: 📦 file-zip
- 其他: 📄 file