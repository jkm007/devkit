-- 添加 Ceph RGW 存储配置
-- 先检查是否已存在，避免重复插入

-- 插入存储连接配置
INSERT INTO sys_storage_config (name, driver, endpoint, access_key, secret_key, bucket, region, use_ssl, is_default, presigned_url_expiry, status, description, created_at, updated_at)
SELECT 'Ceph RGW', 'minio', '10.0.50.1:8080', '688VOVPD0CJ1RN9N6GSX', '5yUu2nDTeREL1yKqqpRIPpb9cA7q5XG8a4IyXDCk', 'devkit', '', 0, 0, 3600, 1, 'Ceph Object Gateway (S3 兼容)', NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_storage_config WHERE name = 'Ceph RGW'
);

-- 插入存储桶配置
INSERT INTO sys_storage_bucket (name, driver, endpoint, bucket, access_key, secret_key, region, use_ssl, purpose, is_default, status, description, created_at, updated_at)
SELECT 'Ceph RGW - devkit', 'minio', '10.0.50.1:8080', 'devkit', '688VOVPD0CJ1RN9N6GSX', '5yUu2nDTeREL1yKqqpRIPpb9cA7q5XG8a4IyXDCk', '', 0, 'general', 0, 1, 'Ceph RGW 通用存储桶', NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_storage_bucket WHERE name = 'Ceph RGW - devkit'
);
