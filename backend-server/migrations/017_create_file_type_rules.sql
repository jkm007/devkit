-- ============================================================
-- 迁移: 017
-- 描述: 创建文件类型检测规则表
-- 作者: Claude Code
-- 日期: 2026-06-09
-- ============================================================

CREATE TABLE IF NOT EXISTS `sys_file_type_rules` (
    `id`          BIGINT NOT NULL AUTO_INCREMENT,
    `extension`   VARCHAR(20) NOT NULL COMMENT '文件扩展名，如 .jpg',
    `file_type`   VARCHAR(50) NOT NULL COMMENT '文件类型，如 image',
    `description` VARCHAR(200) DEFAULT '' COMMENT '描述',
    `status`      TINYINT DEFAULT 1 COMMENT '状态: 1=启用, 0=禁用',
    `created_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_extension` (`extension`),
    KEY `idx_file_type` (`file_type`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件类型检测规则表';

-- ============================================================
-- 插入默认文件类型规则（与 auto_tagger.go 硬编码保持一致）
-- ============================================================

-- 图片类型
INSERT INTO `sys_file_type_rules` (`extension`, `file_type`, `description`) VALUES
('.jpg',    'image',    'JPEG 图片'),
('.jpeg',   'image',    'JPEG 图片'),
('.png',    'image',    'PNG 图片'),
('.gif',    'image',    'GIF 图片'),
('.bmp',    'image',    'BMP 位图'),
('.webp',   'image',    'WebP 图片'),
('.svg',    'image',    'SVG 矢量图'),
('.ico',    'image',    'ICO 图标'),
('.tiff',   'image',    'TIFF 图片'),
('.tif',    'image',    'TIFF 图片'),
('.heic',   'image',    'HEIC 图片'),
('.heif',   'image',    'HEIF 图片');

-- 视频类型
INSERT INTO `sys_file_type_rules` (`extension`, `file_type`, `description`) VALUES
('.mp4',    'video',    'MP4 视频'),
('.avi',    'video',    'AVI 视频'),
('.mov',    'video',    'QuickTime 视频'),
('.wmv',    'video',    'Windows Media 视频'),
('.flv',    'video',    'Flash 视频'),
('.mkv',    'video',    'Matroska 视频'),
('.webm',   'video',    'WebM 视频'),
('.m4v',    'video',    'MPEG-4 视频'),
('.mpeg',   'video',    'MPEG 视频'),
('.mpg',    'video',    'MPEG 视频'),
('.3gp',    'video',    '3GP 视频');

-- 音频类型
INSERT INTO `sys_file_type_rules` (`extension`, `file_type`, `description`) VALUES
('.mp3',    'audio',    'MP3 音频'),
('.wav',    'audio',    'WAV 音频'),
('.flac',   'audio',    'FLAC 无损音频'),
('.aac',    'audio',    'AAC 音频'),
('.ogg',    'audio',    'OGG 音频'),
('.wma',    'audio',    'Windows Media 音频'),
('.m4a',    'audio',    'MPEG-4 音频'),
('.opus',   'audio',    'Opus 音频');

-- 文档类型
INSERT INTO `sys_file_type_rules` (`extension`, `file_type`, `description`) VALUES
('.pdf',    'document', 'PDF 文档'),
('.doc',    'document', 'Word 文档'),
('.docx',   'document', 'Word 文档'),
('.xls',    'document', 'Excel 表格'),
('.xlsx',   'document', 'Excel 表格'),
('.ppt',    'document', 'PowerPoint 演示'),
('.pptx',   'document', 'PowerPoint 演示'),
('.txt',    'document', '纯文本'),
('.md',     'document', 'Markdown 文档'),
('.csv',    'document', 'CSV 数据'),
('.rtf',    'document', '富文本格式');

-- 压缩包类型
INSERT INTO `sys_file_type_rules` (`extension`, `file_type`, `description`) VALUES
('.zip',    'archive',  'ZIP 压缩包'),
('.rar',    'archive',  'RAR 压缩包'),
('.7z',     'archive',  '7-Zip 压缩包'),
('.tar',    'archive',  'TAR 归档'),
('.gz',     'archive',  'Gzip 压缩'),
('.bz2',    'archive',  'Bzip2 压缩'),
('.xz',     'archive',  'XZ 压缩');
