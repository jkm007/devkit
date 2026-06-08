-- ============================================================
-- 历史数据迁移：为已有文件补充标签
-- 说明：根据文件的 content_type 自动为已有文件打标签
-- 执行前请先备份数据库！
-- ============================================================

-- 1. 确保标签数据存在（如果不存在则插入）
INSERT IGNORE INTO sys_tag (tag_key, tag_value, tag_name, icon, color, is_system, sort_order) VALUES
('type', 'image', '图片', '🖼️', '#52c41a', 1, 1),
('type', 'video', '视频', '🎬', '#1890ff', 1, 2),
('type', 'audio', '音频', '🎵', '#722ed1', 1, 3),
('type', 'document', '文档', '📄', '#fa8c16', 1, 4),
('type', 'archive', '压缩包', '📦', '#13c2c2', 1, 5),
('type', 'other', '其他', '📎', '#8c8c8c', 1, 99),
('source', 'user', '用户上传', '👤', '#1890ff', 1, 1),
('source', 'system', '系统生成', '⚙️', '#595959', 1, 2),
('sensitivity', 'public', '公开', '🔓', '#52c41a', 1, 1),
('sensitivity', 'internal', '内部', '🔒', '#faad14', 1, 2),
('sensitivity', 'confidential', '机密', '🔐', '#f5222d', 1, 3);

-- 2. 为图片文件添加标签
INSERT IGNORE INTO sys_file_tag (file_id, tag_id, source)
SELECT
    fe.id AS file_id,
    t.id AS tag_id,
    'auto' AS source
FROM sys_file_entries fe
JOIN sys_tag t ON t.tag_key = 'type' AND t.tag_value = 'image'
WHERE fe.content_type LIKE 'image/%'
  AND fe.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_file_tag ft
    WHERE ft.file_id = fe.id AND ft.tag_id = t.id
  );

-- 3. 为视频文件添加标签
INSERT IGNORE INTO sys_file_tag (file_id, tag_id, source)
SELECT
    fe.id AS file_id,
    t.id AS tag_id,
    'auto' AS source
FROM sys_file_entries fe
JOIN sys_tag t ON t.tag_key = 'type' AND t.tag_value = 'video'
WHERE fe.content_type LIKE 'video/%'
  AND fe.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_file_tag ft
    WHERE ft.file_id = fe.id AND ft.tag_id = t.id
  );

-- 4. 为音频文件添加标签
INSERT IGNORE INTO sys_file_tag (file_id, tag_id, source)
SELECT
    fe.id AS file_id,
    t.id AS tag_id,
    'auto' AS source
FROM sys_file_entries fe
JOIN sys_tag t ON t.tag_key = 'type' AND t.tag_value = 'audio'
WHERE fe.content_type LIKE 'audio/%'
  AND fe.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_file_tag ft
    WHERE ft.file_id = fe.id AND ft.tag_id = t.id
  );

-- 5. 为文档文件添加标签
INSERT IGNORE INTO sys_file_tag (file_id, tag_id, source)
SELECT
    fe.id AS file_id,
    t.id AS tag_id,
    'auto' AS source
FROM sys_file_entries fe
JOIN sys_tag t ON t.tag_key = 'type' AND t.tag_value = 'document'
WHERE (
    fe.content_type LIKE 'application/pdf'
    OR fe.content_type LIKE 'application/msword'
    OR fe.content_type LIKE 'application/vnd.openxmlformats%'
    OR fe.content_type LIKE 'application/vnd.ms-%'
    OR fe.content_type LIKE 'text/plain'
    OR fe.content_type LIKE 'text/csv'
    OR fe.content_type LIKE 'text/markdown'
)
AND fe.deleted_at IS NULL
AND NOT EXISTS (
    SELECT 1 FROM sys_file_tag ft
    WHERE ft.file_id = fe.id AND ft.tag_id = t.id
);

-- 6. 为压缩包文件添加标签
INSERT IGNORE INTO sys_file_tag (file_id, tag_id, source)
SELECT
    fe.id AS file_id,
    t.id AS tag_id,
    'auto' AS source
FROM sys_file_entries fe
JOIN sys_tag t ON t.tag_key = 'type' AND t.tag_value = 'archive'
WHERE (
    fe.content_type LIKE 'application/zip'
    OR fe.content_type LIKE 'application/x-rar%'
    OR fe.content_type LIKE 'application/x-7z%'
    OR fe.content_type LIKE 'application/x-tar%'
    OR fe.content_type LIKE 'application/gzip'
)
AND fe.deleted_at IS NULL
AND NOT EXISTS (
    SELECT 1 FROM sys_file_tag ft
    WHERE ft.file_id = fe.id AND ft.tag_id = t.id
);

-- 7. 为其他文件添加标签
INSERT IGNORE INTO sys_file_tag (file_id, tag_id, source)
SELECT
    fe.id AS file_id,
    t.id AS tag_id,
    'auto' AS source
FROM sys_file_entries fe
JOIN sys_tag t ON t.tag_key = 'type' AND t.tag_value = 'other'
WHERE fe.deleted_at IS NULL
  AND fe.content_type NOT LIKE 'image/%'
  AND fe.content_type NOT LIKE 'video/%'
  AND fe.content_type NOT LIKE 'audio/%'
  AND fe.content_type NOT LIKE 'application/pdf'
  AND fe.content_type NOT LIKE 'application/msword'
  AND fe.content_type NOT LIKE 'application/vnd.openxmlformats%'
  AND fe.content_type NOT LIKE 'application/vnd.ms-%'
  AND fe.content_type NOT LIKE 'text/plain'
  AND fe.content_type NOT LIKE 'text/csv'
  AND fe.content_type NOT LIKE 'text/markdown'
  AND fe.content_type NOT LIKE 'application/zip'
  AND fe.content_type NOT LIKE 'application/x-rar%'
  AND fe.content_type NOT LIKE 'application/x-7z%'
  AND fe.content_type NOT LIKE 'application/x-tar%'
  AND fe.content_type NOT LIKE 'application/gzip'
  AND NOT EXISTS (
    SELECT 1 FROM sys_file_tag ft
    WHERE ft.file_id = fe.id AND ft.tag_id = t.id
  );

-- 8. 为所有文件添加来源标签（默认为用户上传）
INSERT IGNORE INTO sys_file_tag (file_id, tag_id, source)
SELECT
    fe.id AS file_id,
    t.id AS tag_id,
    'auto' AS source
FROM sys_file_entries fe
JOIN sys_tag t ON t.tag_key = 'source' AND t.tag_value = 'user'
WHERE fe.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_file_tag ft
    WHERE ft.file_id = fe.id AND ft.tag_id = t.id
  );

-- 9. 为所有文件添加敏感度标签（默认为公开）
INSERT IGNORE INTO sys_file_tag (file_id, tag_id, source)
SELECT
    fe.id AS file_id,
    t.id AS tag_id,
    'auto' AS source
FROM sys_file_entries fe
JOIN sys_tag t ON t.tag_key = 'sensitivity' AND t.tag_value = 'public'
WHERE fe.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_file_tag ft
    WHERE ft.file_id = fe.id AND ft.tag_id = t.id
  );

-- 10. 统计迁移结果
SELECT
    '迁移完成' AS status,
    (SELECT COUNT(*) FROM sys_file_tag) AS total_tags,
    (SELECT COUNT(DISTINCT file_id) FROM sys_file_tag) AS tagged_files,
    (SELECT COUNT(*) FROM sys_file_entries WHERE deleted_at IS NULL) AS total_files;
