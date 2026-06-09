-- 021: 为 sys_menus 表添加 sort 排序字段
-- 这个字段用于控制菜单在同级中的显示顺序，值越小越靠前

-- 1. 添加 sort 列
ALTER TABLE sys_menus ADD COLUMN sort INT NOT NULL DEFAULT 0 COMMENT '排序值(越小越靠前)' AFTER icon;

-- 2. 从 meta JSON 中的 order 字段迁移数据到 sort 列
UPDATE sys_menus SET sort = CAST(JSON_EXTRACT(meta, '$.order') AS UNSIGNED) WHERE JSON_EXTRACT(meta, '$.order') IS NOT NULL;

-- 3. 为 sort 列添加索引（配合 pid 用于排序查询）
ALTER TABLE sys_menus ADD INDEX idx_pid_sort (pid, sort);
