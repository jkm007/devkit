-- 文件分享安全增强迁移
-- 1. file_entries 表添加 is_public 字段（AutoMigrate 自动处理）
-- 2. file_shares 表添加 password 和 has_password 字段（AutoMigrate 自动处理）
-- 3. 将轮播图关联的文件设为公开

-- 将轮播图关联的文件条目设为公开
UPDATE sys_file_entries fe
INNER JOIN banners b ON fe.id = b.file_id
SET fe.is_public = true
WHERE b.file_id IS NOT NULL AND b.status = 'enabled';
