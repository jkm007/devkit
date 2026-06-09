-- 回收站：给文件条目添加软删除和过期时间字段
ALTER TABLE sys_file_entries
  ADD COLUMN deleted_at DATETIME(3) DEFAULT NULL COMMENT '软删除时间' AFTER user_id,
  ADD COLUMN recycle_expire_at DATETIME DEFAULT NULL COMMENT '回收站过期时间（永久删除时间）' AFTER deleted_at;

-- 添加索引
CREATE INDEX idx_file_entries_deleted_at ON sys_file_entries(deleted_at);
CREATE INDEX idx_file_entries_recycle_expire ON sys_file_entries(recycle_expire_at);

-- 定时任务配置表
CREATE TABLE IF NOT EXISTS sys_scheduled_tasks (
  id            BIGINT AUTO_INCREMENT PRIMARY KEY,
  name          VARCHAR(100) NOT NULL COMMENT '任务名称',
  task_type     VARCHAR(50)  NOT NULL COMMENT '任务类型: recycle_cleanup, backup, sync',
  cron_expr     VARCHAR(50)  NOT NULL DEFAULT '0 3 * * *' COMMENT 'Cron 表达式',
  config        JSON         DEFAULT NULL COMMENT '任务配置（JSON）',
  enabled       TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '是否启用',
  status        VARCHAR(20)  NOT NULL DEFAULT 'idle' COMMENT '状态: idle, running, success, failed',
  last_run_at   DATETIME     DEFAULT NULL COMMENT '最后执行时间',
  last_result   TEXT         DEFAULT NULL COMMENT '最后执行结果',
  next_run_at   DATETIME     DEFAULT NULL COMMENT '下次执行时间',
  run_count     INT          NOT NULL DEFAULT 0 COMMENT '执行次数',
  created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at    DATETIME(3)  DEFAULT NULL,
  INDEX idx_scheduled_tasks_type (task_type),
  INDEX idx_scheduled_tasks_enabled (enabled),
  INDEX idx_scheduled_tasks_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务配置';

-- 插入默认的回收站清理任务
INSERT INTO sys_scheduled_tasks (name, task_type, cron_expr, config, enabled) VALUES
('回收站清理', 'recycle_cleanup', '0 3 * * *', '{"retention_days": 7}', 1);

-- 添加回收站菜单
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta) VALUES
(40, 'FileRecycle', '/file/recycle', '/file/recycle/index', 'menu', 1, 'file:recycle:list', 'lucide:trash-2', 3, '{"title":"回收站","icon":"lucide:trash-2","order":3}');

-- 回收站操作按钮
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta) VALUES
((SELECT id FROM (SELECT id FROM sys_menus WHERE name='FileRecycle' AND deleted_at IS NULL LIMIT 1) AS t), 'FileRecycleRestore', '', '', 'button', 1, 'file:recycle:restore', '', 0, '{"title":"恢复文件"}'),
((SELECT id FROM (SELECT id FROM sys_menus WHERE name='FileRecycle' AND deleted_at IS NULL LIMIT 1) AS t), 'FileRecycleDelete', '', '', 'button', 1, 'file:recycle:delete', '', 0, '{"title":"永久删除"}'),
((SELECT id FROM (SELECT id FROM sys_menus WHERE name='FileRecycle' AND deleted_at IS NULL LIMIT 1) AS t), 'FileRecycleEmpty', '', '', 'button', 1, 'file:recycle:empty', '', 0, '{"title":"清空回收站"}');

-- 添加定时任务管理菜单（在系统管理下）
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta) VALUES
(4, 'SystemScheduledTask', '/system/scheduled-task', '/system/scheduled-task/index', 'menu', 1, 'system:task:list', 'lucide:clock', 12, '{"order":12,"title":"定时任务"}');

-- 定时任务操作按钮
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta) VALUES
((SELECT id FROM (SELECT id FROM sys_menus WHERE name='SystemScheduledTask' AND deleted_at IS NULL LIMIT 1) AS t), 'SystemScheduledTaskEdit', '', '', 'button', 1, 'system:task:edit', '', 0, '{"title":"编辑任务"}'),
((SELECT id FROM (SELECT id FROM sys_menus WHERE name='SystemScheduledTask' AND deleted_at IS NULL LIMIT 1) AS t), 'SystemScheduledTaskRun', '', '', 'button', 1, 'system:task:run', '', 0, '{"title":"执行任务"}');
