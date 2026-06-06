-- =============================================
-- 系统设置初始数据
-- =============================================

USE backend_db;

-- 基础设置（string/boolean 类型）
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('basic', 'site_name',         '"管理系统"',  '站点名称',     'string',  '显示在浏览器标题和登录页',                   1, 1, 0, NOW(3), NOW(3)),
('basic', 'site_logo',         '""',          '站点 Logo',    'string',  'Logo 图片 URL',                            2, 1, 0, NOW(3), NOW(3)),
('basic', 'site_description',  '""',          '站点描述',     'string',  '用于 SEO 和登录页展示',                     3, 1, 0, NOW(3), NOW(3)),
('basic', 'copyright',         '""',          '版权信息',     'string',  '页面底部版权文字',                          4, 1, 0, NOW(3), NOW(3)),
('basic', 'watermark_enabled', 'true',        '启用水印',     'boolean', '页面是否显示水印（用户名+时间戳）',           5, 1, 0, NOW(3), NOW(3)),
('basic', 'watermark_content', '""',          '水印内容',     'string',  '自定义水印文字，为空则使用"用户名 时间"',     6, 1, 0, NOW(3), NOW(3));

-- 基础设置（select 类型，需要 options）
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, options, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('basic', 'default_theme', '"auto"',  '默认主题', 'select',
 '[{"label":"跟随系统","value":"auto"},{"label":"浅色","value":"light"},{"label":"深色","value":"dark"}]',
 '用户首次访问时的主题，后续跟随用户偏好', 7, 1, 0, NOW(3), NOW(3)),
('basic', 'default_lang', '"zh-CN"', '默认语言', 'select',
 '[{"label":"简体中文","value":"zh-CN"},{"label":"English","value":"en-US"}]',
 '用户首次访问时的界面语言', 8, 1, 0, NOW(3), NOW(3));

-- 邮箱设置
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('email', 'smtp_enabled',   'false', '启用邮件服务', 'boolean', '开启后可发送验证邮件和通知邮件', 1, 0, 0, NOW(3), NOW(3)),
('email', 'smtp_host',      '""',    'SMTP 服务器',  'string',  '如 smtp.qq.com',                2, 0, 0, NOW(3), NOW(3)),
('email', 'smtp_port',      '465',   'SMTP 端口',    'number',  '通常为 465(SSL) 或 587(TLS)',    3, 0, 0, NOW(3), NOW(3)),
('email', 'smtp_username',  '""',    'SMTP 用户名',  'string',  '登录邮箱账号',                   4, 0, 0, NOW(3), NOW(3)),
('email', 'smtp_password',  '""',    'SMTP 密码',    'string',  '登录邮箱密码或授权码',            5, 0, 1, NOW(3), NOW(3)),
('email', 'smtp_from',      '""',    '发件人地址',   'string',  '如 noreply@example.com',        6, 0, 0, NOW(3), NOW(3)),
('email', 'smtp_from_name', '""',    '发件人名称',   'string',  '显示的发件人名称',               7, 0, 0, NOW(3), NOW(3));

-- 短信设置
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, options, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('sms', 'sms_enabled',      'false',    '启用短信服务', 'boolean', NULL, '开启后可发送短信验证码',             1, 0, 0, NOW(3), NOW(3)),
('sms', 'sms_driver',       '"aliyun"', '短信服务商',   'select',
 '[{"label":"阿里云","value":"aliyun"},{"label":"腾讯云","value":"tencent"}]',
 '选择短信服务商', 2, 0, 0, NOW(3), NOW(3)),
('sms', 'sms_access_key',   '""',       'AccessKey',    'string',  NULL, '短信服务商的 AccessKey',             3, 0, 0, NOW(3), NOW(3)),
('sms', 'sms_secret_key',   '""',       'SecretKey',    'string',  NULL, '短信服务商的 SecretKey',             4, 0, 1, NOW(3), NOW(3)),
('sms', 'sms_sign_name',    '""',       '短信签名',     'string',  NULL, '短信模板中使用的签名',               5, 0, 0, NOW(3), NOW(3)),
('sms', 'sms_template_code','""',       '短信模板码',   'string',  NULL, '验证码短信的模板编码',               6, 0, 0, NOW(3), NOW(3));

-- 验证码设置（通用）
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, options, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('captcha', 'captcha_enabled',      'true',     '启用验证码',   'boolean', NULL, '登录时是否需要验证码',                                1, 1, 0, NOW(3), NOW(3)),
('captcha', 'captcha_type',         '"slider"', '验证码类型',   'select',
 '[{"label":"滑块","value":"slider"},{"label":"拼图","value":"puzzle"},{"label":"旋转","value":"rotation"},{"label":"点选","value":"point"},{"label":"数字","value":"numeric"}]',
 '验证码展示形式', 2, 1, 0, NOW(3), NOW(3)),
('captcha', 'captcha_expire',       '120',      '验证码有效期(秒)', 'number',  NULL, '验证码有效时间',                                 3, 0, 0, NOW(3), NOW(3)),
('captcha', 'captcha_max_fail',     '5',        '最大失败次数', 'number',  NULL, '连续验证失败后刷新验证码',                             4, 0, 0, NOW(3), NOW(3)),
('captcha', 'captcha_login_trigger','3',        '触发阈值',     'number',  NULL, '登录失败几次后开始要求验证码（0=始终开启）',            5, 0, 0, NOW(3), NOW(3)),
('captcha', 'captcha_min_duration', '500',      '最短操作时间(ms)', 'number',  NULL, '操作时间小于此值判定为机器人',                         6, 0, 0, NOW(3), NOW(3));

-- 验证码设置（数字验证码）
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('captcha', 'numeric_length',  '4',   '数字位数', 'number', '验证码数字长度（4-6位）',     10, 0, 0, NOW(3), NOW(3)),
('captcha', 'numeric_width',   '160', '图片宽度', 'number', '验证码图片宽度',             11, 0, 0, NOW(3), NOW(3)),
('captcha', 'numeric_height',  '60',  '图片高度', 'number', '验证码图片高度',             12, 0, 0, NOW(3), NOW(3));

-- 验证码设置（滑块验证码）
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('captcha', 'slider_width',      '320', '图片宽度', 'number', '背景图片宽度',           20, 0, 0, NOW(3), NOW(3)),
('captcha', 'slider_height',     '200', '图片高度', 'number', '背景图片高度',           21, 0, 0, NOW(3), NOW(3)),
('captcha', 'slider_tolerance',  '5',   '验证容差(px)', 'number', 'X坐标允许误差',         22, 0, 0, NOW(3), NOW(3));

-- 验证码设置（拼图验证码）
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('captcha', 'puzzle_width',           '320', '图片宽度', 'number', '背景图片宽度',           30, 0, 0, NOW(3), NOW(3)),
('captcha', 'puzzle_height',          '200', '图片高度', 'number', '背景图片高度',           31, 0, 0, NOW(3), NOW(3)),
('captcha', 'puzzle_tolerance',       '5',   '验证容差(px)', 'number', '坐标允许误差',           32, 0, 0, NOW(3), NOW(3)),
('captcha', 'puzzle_vertical_random', 'true', '垂直随机', 'boolean', '是否随机垂直位置',        33, 0, 0, NOW(3), NOW(3));

-- 验证码设置（旋转验证码）
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('captcha', 'rotation_size',        '220', '图片尺寸', 'number', '正方形图片边长',         40, 0, 0, NOW(3), NOW(3)),
('captcha', 'rotation_thumb_size',  '80',  '缩略图尺寸', 'number', '旋转缩略图边长（80-100）', 41, 0, 0, NOW(3), NOW(3)),
('captcha', 'rotation_tolerance',   '10',  '角度容差(度)', 'number', '角度允许误差',           42, 0, 0, NOW(3), NOW(3));

-- 验证码设置（点选验证码）
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('captcha', 'point_width',      '320', '图片宽度', 'number', '背景图片宽度',           50, 0, 0, NOW(3), NOW(3)),
('captcha', 'point_height',     '220', '图片高度', 'number', '背景图片高度',           51, 0, 0, NOW(3), NOW(3)),
('captcha', 'point_count',      '4',   '点击数量', 'number', '需点击的文字数量（4-6）', 52, 0, 0, NOW(3), NOW(3)),
('captcha', 'point_tolerance',  '30',  '验证容差(px)', 'number', '点击位置允许误差',        53, 0, 0, NOW(3), NOW(3));

-- 存储设置
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, options, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('storage', 'storage_driver',      '"local"', '存储驱动',       'select',
 '[{"label":"本地存储","value":"local"},{"label":"MinIO","value":"minio"},{"label":"阿里云 OSS","value":"oss"},{"label":"腾讯云 COS","value":"cos"}]',
 '文件存储方式', 1, 0, 0, NOW(3), NOW(3)),
('storage', 'storage_max_size',    '10',      '单文件上限(MB)', 'number',  NULL, '单个文件上传大小限制',                                 2, 0, 0, NOW(3), NOW(3)),
('storage', 'storage_allowed_ext', '["jpg","png","gif","pdf","doc","docx","xls","xlsx","mp4","zip"]',
 '允许的文件类型', 'json', NULL, '允许上传的文件扩展名', 3, 0, 0, NOW(3), NOW(3));

-- MinIO 配置（开发环境默认值，生产环境请通过管理后台修改）
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('storage', 'minio_endpoint',    '"10.0.50.108:9000"',  'MinIO Endpoint',  'string',  '如 localhost:9000',  4, 0, 0, NOW(3), NOW(3)),
('storage', 'minio_access_key',  '"minioadmin"',        'MinIO AccessKey', 'string',  '',                   5, 0, 0, NOW(3), NOW(3)),
('storage', 'minio_secret_key',  '"minioadmin123"',     'MinIO SecretKey', 'string',  '',                   6, 0, 1, NOW(3), NOW(3)),
('storage', 'minio_bucket',      '"backend"',           'MinIO Bucket',    'string',  '存储桶名称',          7, 0, 0, NOW(3), NOW(3)),
('storage', 'minio_use_ssl',     'false',               'MinIO SSL',       'boolean', '是否使用 HTTPS',      8, 0, 0, NOW(3), NOW(3));

-- 阿里云 OSS 配置
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('storage', 'oss_endpoint',        '""',    'OSS Endpoint',        'string',  '如 oss-cn-shenzhen.aliyuncs.com',  9,  0, 0, NOW(3), NOW(3)),
('storage', 'oss_access_key',      '""',    'OSS AccessKey',       'string',  '',                                10, 0, 0, NOW(3), NOW(3)),
('storage', 'oss_secret_key',      '""',    'OSS SecretKey',       'string',  '',                                11, 0, 1, NOW(3), NOW(3)),
('storage', 'oss_bucket',          '""',    'OSS Bucket',          'string',  '存储桶名称',                       12, 0, 0, NOW(3), NOW(3)),
('storage', 'oss_cdn_domain',      '""',    'OSS CDN 域名',        'string',  '可选，配置后使用 CDN 地址访问',     13, 0, 0, NOW(3), NOW(3));

-- 腾讯云 COS 配置
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('storage', 'cos_region',          '""',    'COS Region',          'string',  '如 ap-guangzhou',                  14, 0, 0, NOW(3), NOW(3)),
('storage', 'cos_secret_id',       '""',    'COS SecretId',        'string',  '',                                 15, 0, 0, NOW(3), NOW(3)),
('storage', 'cos_secret_key',      '""',    'COS SecretKey',       'string',  '',                                 16, 0, 1, NOW(3), NOW(3)),
('storage', 'cos_bucket',          '""',    'COS Bucket',          'string',  '存储桶名称',                       17, 0, 0, NOW(3), NOW(3));

-- 微信设置
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('wechat', 'wechat_miniapp_enabled',  'false', '启用小程序登录',   'boolean', '开启微信小程序登录功能',        1, 0, 0, NOW(3), NOW(3)),
('wechat', 'wechat_miniapp_appid',    '""',    '小程序 AppID',     'string',  '',                              2, 0, 0, NOW(3), NOW(3)),
('wechat', 'wechat_miniapp_secret',   '""',    '小程序 AppSecret', 'string',  '',                              3, 0, 1, NOW(3), NOW(3)),
('wechat', 'wechat_official_enabled', 'false', '启用公众号登录',   'boolean', '开启微信公众号网页授权登录',     4, 0, 0, NOW(3), NOW(3)),
('wechat', 'wechat_official_appid',   '""',    '公众号 AppID',     'string',  '',                              5, 0, 0, NOW(3), NOW(3)),
('wechat', 'wechat_official_secret',  '""',    '公众号 AppSecret', 'string',  '',                              6, 0, 1, NOW(3), NOW(3)),
('wechat', 'wechat_oauth_enabled',    'false', '启用微信扫码登录', 'boolean', 'PC 端微信扫码登录',              7, 0, 0, NOW(3), NOW(3)),
('wechat', 'wechat_oauth_appid',     '""',    '网站应用 AppID',   'string',  '微信开放平台-网站应用的 AppID',   8, 0, 0, NOW(3), NOW(3)),
('wechat', 'wechat_oauth_secret',    '""',    '网站应用 AppSecret','string',  '微信开放平台-网站应用的 AppSecret',9, 0, 1, NOW(3), NOW(3)),
('wechat', 'wechat_oauth_redirect_url','""',  '网站扫码回调地址', 'string',  '如 https://your-domain/auth/wechat/web-callback', 10, 0, 0, NOW(3), NOW(3)),
('wechat', 'wechat_official_redirect_url','""','公众号回调地址',  'string',  '如 https://your-domain/auth/wechat/official-callback', 11, 0, 0, NOW(3), NOW(3));

-- 安全设置
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('security', 'login_fail_lock',        'true', '登录失败锁定',     'boolean', '连续登录失败后是否锁定账号',     1, 0, 0, NOW(3), NOW(3)),
('security', 'login_fail_max',         '5',    '最大失败次数',     'number',  '连续失败几次后锁定',             2, 0, 0, NOW(3), NOW(3)),
('security', 'login_lock_duration',    '30',   '锁定时长(分钟)',   'number',  '账号锁定持续时间',               3, 0, 0, NOW(3), NOW(3)),
('security', 'password_min_length',    '6',    '密码最小长度',     'number',  '',                               4, 0, 0, NOW(3), NOW(3)),
('security', 'password_history_count', '5',    '密码历史检查次数', 'number',  '修改密码时检查最近N次是否重复',   5, 0, 0, NOW(3), NOW(3)),
('security', 'session_timeout',        '60',   '会话超时(分钟)',   'number',  '前端无操作自动登出时间',          6, 0, 0, NOW(3), NOW(3)),
('security', 'allow_multi_device',     'true', '允许多设备登录',   'boolean', '关闭后新设备登录会踢出旧设备',    7, 0, 0, NOW(3), NOW(3));

-- 风险评分配置
INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, is_sensitive, created_at, updated_at) VALUES
('risk_score', 'risk_enabled',        'true',  '启用风险评分',         'boolean', '开启后对敏感接口进行风险评估，高风险请求需要验证码', 1, 0, 0, NOW(3), NOW(3)),
('risk_score', 'risk_trigger_score',  '50',    '触发验证码阈值',       'number',  '风险分达到此值时要求验证码',                         2, 0, 0, NOW(3), NOW(3)),
('risk_score', 'risk_block_score',    '80',    '直接拦截阈值',         'number',  '风险分达到此值时直接拒绝请求（0=不拦截）',            3, 0, 0, NOW(3), NOW(3)),
('risk_score', 'risk_decay_minutes',  '30',    '分数衰减周期(分钟)',    'number',  '无请求后风险分开始衰减的时间',                       4, 0, 0, NOW(3), NOW(3)),
('risk_score', 'risk_decay_rate',     '0.5',   '衰减比例',             'number',  '每个衰减周期风险分减少的比例（0-1）',                 5, 0, 0, NOW(3), NOW(3)),
('risk_score', 'risk_protected_paths','"/system/user,/system/role,/system/group,/system/settings"', '受保护路径前缀', 'string', '需要风险评估的API路径前缀（逗号分隔）', 6, 0, 0, NOW(3), NOW(3)),

('risk_score', 'rule_frequency_enabled',   'true',  '频率检测-启用',           'boolean', '',  10, 0, 0, NOW(3), NOW(3)),
('risk_score', 'rule_frequency_threshold', '30',    '频率检测-每分钟请求阈值',  'number',  '',  11, 0, 0, NOW(3), NOW(3)),
('risk_score', 'rule_frequency_score',     '30',    '频率检测-超限加分',        'number',  '',  12, 0, 0, NOW(3), NOW(3)),

('risk_score', 'rule_no_referer_enabled',  'true',  '无Referer检测-启用',      'boolean', '直接调用API而非从页面发起的请求通常没有Referer', 20, 0, 0, NOW(3), NOW(3)),
('risk_score', 'rule_no_referer_score',    '20',    '无Referer检测-加分',      'number',  '',  21, 0, 0, NOW(3), NOW(3)),

('risk_score', 'rule_no_lang_enabled',     'true',  '无Accept-Language-启用',  'boolean', '机器请求通常不携带Accept-Language头', 30, 0, 0, NOW(3), NOW(3)),
('risk_score', 'rule_no_lang_score',       '15',    '无Accept-Language-加分',  'number',  '',  31, 0, 0, NOW(3), NOW(3)),

('risk_score', 'rule_ua_enabled',          'true',  'UA异常检测-启用',         'boolean', '',  40, 0, 0, NOW(3), NOW(3)),
('risk_score', 'rule_ua_keywords',         '"curl,python,java,go-http,postman,scrapy,requests,httpie"', 'UA异常关键词', 'string', '包含这些关键词的User-Agent视为异常（逗号分隔）', 41, 0, 0, NOW(3), NOW(3)),
('risk_score', 'rule_ua_score',            '25',    'UA异常检测-加分',         'number',  '',  42, 0, 0, NOW(3), NOW(3)),

('risk_score', 'rule_interval_enabled',    'true',  '请求间隔检测-启用',       'boolean', '短时间内大量请求视为异常', 60, 0, 0, NOW(3), NOW(3)),
('risk_score', 'rule_interval_min_ms',     '500',   '最小请求间隔(毫秒)',      'number',  '',  61, 0, 0, NOW(3), NOW(3)),
('risk_score', 'rule_interval_score',      '20',    '间隔过短-加分',           'number',  '',  62, 0, 0, NOW(3), NOW(3));
