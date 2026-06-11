-- ============================================================
-- 迁移: 029
-- 描述: 初始化基础数据（文件类型规则、标签、限速规则、定时任务）
-- 作者: Claude Code
-- 日期: 2026-06-11
-- ============================================================

USE backend_db;

-- ============================================================
-- 第一部分：初始化文件类型规则
-- ============================================================

INSERT INTO sys_file_type_rules (extension, file_type, description, status, created_at, updated_at) VALUES
-- 图片格式
('jpg', 'image', 'JPEG 图片', 1, NOW(), NOW()),
('jpeg', 'image', 'JPEG 图片', 1, NOW(), NOW()),
('png', 'image', 'PNG 图片', 1, NOW(), NOW()),
('gif', 'image', 'GIF 动画', 1, NOW(), NOW()),
('bmp', 'image', 'BMP 图片', 1, NOW(), NOW()),
('webp', 'image', 'WebP 图片', 1, NOW(), NOW()),
('svg', 'image', 'SVG 矢量图', 1, NOW(), NOW()),
('ico', 'image', '图标文件', 1, NOW(), NOW()),
('tif', 'image', 'TIFF 图片', 1, NOW(), NOW()),
('tiff', 'image', 'TIFF 图片', 1, NOW(), NOW()),
('psd', 'image', 'Photoshop 文件', 1, NOW(), NOW()),
('raw', 'image', 'RAW 照片', 1, NOW(), NOW()),
('exif', 'image', 'EXIF 图片', 1, NOW(), NOW()),

-- 文档格式
('doc', 'document', 'Word 文档', 1, NOW(), NOW()),
('docx', 'document', 'Word 文档', 1, NOW(), NOW()),
('xls', 'document', 'Excel 表格', 1, NOW(), NOW()),
('xlsx', 'document', 'Excel 表格', 1, NOW(), NOW()),
('ppt', 'document', 'PowerPoint 演示', 1, NOW(), NOW()),
('pptx', 'document', 'PowerPoint 演示', 1, NOW(), NOW()),
('pdf', 'document', 'PDF 文档', 1, NOW(), NOW()),
('txt', 'document', '纯文本文件', 1, NOW(), NOW()),
('rtf', 'document', 'RTF 富文本', 1, NOW(), NOW()),
('csv', 'document', 'CSV 逗号分隔', 1, NOW(), NOW()),
('json', 'document', 'JSON 数据', 1, NOW(), NOW()),
('xml', 'document', 'XML 数据', 1, NOW(), NOW()),
('md', 'document', 'Markdown 文档', 1, NOW(), NOW()),
('markdown', 'document', 'Markdown 文档', 1, NOW(), NOW()),

-- 压缩格式
('zip', 'archive', 'ZIP 压缩包', 1, NOW(), NOW()),
('rar', 'archive', 'RAR 压缩包', 1, NOW(), NOW()),
('7z', 'archive', '7z 压缩包', 1, NOW(), NOW()),
('tar', 'archive', 'TAR 归档', 1, NOW(), NOW()),
('gz', 'archive', 'GZ 压缩', 1, NOW(), NOW()),
('bz2', 'archive', 'BZ2 压缩', 1, NOW(), NOW()),
('xz', 'archive', 'XZ 压缩', 1, NOW(), NOW()),
('tgz', 'archive', 'TGZ 压缩包', 1, NOW(), NOW()),
('iso', 'archive', 'ISO 光盘镜像', 1, NOW(), NOW()),
('dmg', 'archive', 'DMG 苹果镜像', 1, NOW(), NOW()),

-- 音频格式
('mp3', 'audio', 'MP3 音频', 1, NOW(), NOW()),
('wav', 'audio', 'WAV 音频', 1, NOW(), NOW()),
('wma', 'audio', 'WMA 音频', 1, NOW(), NOW()),
('ogg', 'audio', 'OGG 音频', 1, NOW(), NOW()),
('flac', 'audio', 'FLAC 无损音频', 1, NOW(), NOW()),
('aac', 'audio', 'AAC 音频', 1, NOW(), NOW()),
('m4a', 'audio', 'M4A 音频', 1, NOW(), NOW()),
('ape', 'audio', 'APE 无损音频', 1, NOW(), NOW()),

-- 视频格式
('mp4', 'video', 'MP4 视频', 1, NOW(), NOW()),
('avi', 'video', 'AVI 视频', 1, NOW(), NOW()),
('mov', 'video', 'MOV 视频', 1, NOW(), NOW()),
('wmv', 'video', 'WMV 视频', 1, NOW(), NOW()),
('flv', 'video', 'FLV 视频', 1, NOW(), NOW()),
('mkv', 'video', 'MKV 视频', 1, NOW(), NOW()),
('webm', 'video', 'WebM 视频', 1, NOW(), NOW()),
('mts', 'video', 'MTS 视频', 1, NOW(), NOW()),
('ts', 'video', 'TS 视频流', 1, NOW(), NOW()),
('rmvb', 'video', 'RMVB 视频', 1, NOW(), NOW()),
('3gp', 'video', '3GP 手机视频', 1, NOW(), NOW()),

-- 可执行文件
('exe', 'executable', 'Windows 可执行', 1, NOW(), NOW()),
('msi', 'executable', 'Windows 安装包', 1, NOW(), NOW()),
('dmg', 'executable', 'Mac 安装包', 1, NOW(), NOW()),
('deb', 'executable', 'Debian 安装包', 1, NOW(), NOW()),
('rpm', 'executable', 'RedHat 安装包', 1, NOW(), NOW()),
('apk', 'executable', 'Android 安装包', 1, NOW(), NOW()),
('ipa', 'executable', 'iOS 安装包', 1, NOW(), NOW()),

-- 编程相关
('html', 'code', 'HTML 页面', 1, NOW(), NOW()),
('htm', 'code', 'HTML 页面', 1, NOW(), NOW()),
('css', 'code', 'CSS 样式', 1, NOW(), NOW()),
('js', 'code', 'JavaScript 脚本', 1, NOW(), NOW()),
('ts', 'code', 'TypeScript 脚本', 1, NOW(), NOW()),
('py', 'code', 'Python 脚本', 1, NOW(), NOW()),
('java', 'code', 'Java 源文件', 1, NOW(), NOW()),
('c', 'code', 'C 源文件', 1, NOW(), NOW()),
('cpp', 'code', 'C++ 源文件', 1, NOW(), NOW()),
('h', 'code', 'C/C++ 头文件', 1, NOW(), NOW()),
('go', 'code', 'Go 源文件', 1, NOW(), NOW()),
('rs', 'code', 'Rust 源文件', 1, NOW(), NOW()),
('sql', 'code', 'SQL 脚本', 1, NOW(), NOW()),
('sh', 'code', 'Shell 脚本', 1, NOW(), NOW()),
('bat', 'code', 'Windows 批处理', 1, NOW(), NOW()),
('ps1', 'code', 'PowerShell 脚本', 1, NOW(), NOW()),
('php', 'code', 'PHP 脚本', 1, NOW(), NOW()),
('rb', 'code', 'Ruby 脚本', 1, NOW(), NOW()),
('swift', 'code', 'Swift 源文件', 1, NOW(), NOW()),
('kt', 'code', 'Kotlin 源文件', 1, NOW(), NOW()),
('dart', 'code', 'Dart 源文件', 1, NOW(), NOW()),
('vue', 'code', 'Vue 组件', 1, NOW(), NOW()),
('jsx', 'code', 'React 组件', 1, NOW(), NOW()),
('tsx', 'code', 'React TS 组件', 1, NOW(), NOW()),
('json', 'code', 'JSON 文件', 1, NOW(), NOW()),
('yaml', 'code', 'YAML 配置', 1, NOW(), NOW()),
('yml', 'code', 'YAML 配置', 1, NOW(), NOW()),
('toml', 'code', 'TOML 配置', 1, NOW(), NOW()),
('ini', 'code', 'INI 配置', 1, NOW(), NOW()),
('conf', 'code', '配置文件', 1, NOW(), NOW()),
('log', 'code', '日志文件', 1, NOW(), NOW()),
('patch', 'code', '补丁文件', 1, NOW(), NOW()),
('diff', 'code', '差异文件', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- ============================================================
-- 第二部分：初始化系统标签
-- ============================================================

INSERT INTO sys_tag (tag_key, tag_value, tag_name, icon, color, description, is_system, sort_order, status, created_at, updated_at) VALUES
('type', 'image', '图片', 'lucide:image', '#1890ff', '图片文件', 1, 1, 1, NOW(), NOW()),
('type', 'document', '文档', 'lucide:file-text', '#52c41a', '文档文件', 1, 2, 1, NOW(), NOW()),
('type', 'video', '视频', 'lucide:video', '#fa8c16', '视频文件', 1, 3, 1, NOW(), NOW()),
('type', 'audio', '音频', 'lucide:music', '#722ed1', '音频文件', 1, 4, 1, NOW(), NOW()),
('type', 'archive', '压缩包', 'lucide:folder-plus', '#eb2f96', '压缩文件', 1, 5, 1, NOW(), NOW()),
('type', 'executable', '可执行', 'lucide:play', '#faad14', '可执行文件', 1, 6, 1, NOW(), NOW()),
('type', 'code', '代码', 'lucide:code', '#f5222d', '代码文件', 1, 7, 1, NOW(), NOW()),
('status', 'public', '公开', 'lucide:globe', '#13c2c2', '公开访问', 1, 8, 1, NOW(), NOW()),
('status', 'private', '私有', 'lucide:lock', '#faad14', '仅自己可见', 1, 9, 1, NOW(), NOW()),
('status', 'protected', '受保护', 'lucide:shield-check', '#52c41a', '受保护文件', 1, 10, 1, NOW(), NOW()),
('project', 'design', '设计', 'lucide:palette', '#eb2f96', '设计相关', 1, 11, 1, NOW(), NOW()),
('project', 'document', '文档', 'lucide:book', '#1890ff', '文档项目', 1, 12, 1, NOW(), NOW()),
('project', 'video', '视频', 'lucide:camera', '#fa8c16', '视频项目', 1, 13, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- ============================================================
-- 第三部分：初始化限速规则
-- ============================================================

INSERT INTO sys_rate_limit_rules (
    name, path, method, limit_per_minute, limit_per_hour, limit_per_day,
    is_global, enabled, description, sort_order,
    cooldown, block_duration, max_violations, violation_score,
    created_at, updated_at
) VALUES
-- 登录限速
('登录接口', '/api/v1/auth/login', 'POST', 5, 10, 20, 0, 1, '登录接口限制，防止暴力破解', 1,
 60, 300, 5, 10, NOW(), NOW()),
-- 注册限速
('注册接口', '/api/v1/auth/register', 'POST', 3, 10, 50, 0, 1, '注册接口限制，防止垃圾注册', 2,
 120, 600, 3, 5, NOW(), NOW()),
-- 通用 API 限速
('通用API', '/api/*', 'ANY', 60, 600, 10000, 1, 0, '全局 API 限速，防止滥用', 99,
 0, 0, 0, 0, NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- ============================================================
-- 第四部分：初始化定时任务（示例任务）
-- ============================================================

INSERT INTO sys_scheduled_tasks (
    name, task_type, cron_expression, params,
    status, description, sort_order,
    last_run_at, next_run_at,
    created_at, updated_at
) VALUES
-- 清理过期上传任务
('清理过期上传', 'clean_uploads', '0 3 * * *', '{"days":7}', 1, '每天凌晨3点清理7天前的过期上传任务', 1, NULL, NULL, NOW(), NOW()),
-- 清理旧的安全日志
('清理安全日志', 'clean_security_logs', '0 4 1 * *', '{"days":90}', 1, '每月1号凌晨4点清理90天前的安全日志', 2, NULL, NULL, NOW(), NOW()),
-- 清理回收站
('清理回收站', 'clean_recycle_bin', '0 5 * * 0', '{"days":30}', 1, '每周日凌晨5点清理30天前的回收站文件', 3, NULL, NULL, NOW(), NOW()),
-- 数据统计
('数据统计', 'data_statistics', '0 6 * * *', '{}', 1, '每天凌晨6点统计数据', 4, NULL, NULL, NOW(), NOW()),
-- 检查存储状态
('检查存储状态', 'check_storage', '30 * * * *', '{}', 1, '每小时检查存储状态', 5, NULL, NULL, NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- ============================================================
-- 完成
-- ============================================================
SELECT '基础数据初始化完成' AS status;
