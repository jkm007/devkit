-- 20260629: 为 MyPageMenu 表添加 link_type 字段
ALTER TABLE mobile_my_page_menus ADD COLUMN IF NOT EXISTS link_type VARCHAR(20) DEFAULT 'page' NOT NULL;
