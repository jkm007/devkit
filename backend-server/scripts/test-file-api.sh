#!/bin/bash

# 文件管理 API 测试脚本
# 使用方法: ./test-file-api.sh <TOKEN>

BASE_URL="http://localhost:8080"
TOKEN="$1"

if [ -z "$TOKEN" ]; then
    echo "错误: 请提供访问令牌"
    echo "使用方法: ./test-file-api.sh <TOKEN>"
    echo ""
    echo "获取令牌步骤:"
    echo "1. 访问 http://localhost:5667/ 登录"
    echo "2. 打开浏览器开发者工具 (F12)"
    echo "3. 在 Application/Storage > Cookies 中找到 access_token"
    echo "4. 或在 Network 标签中查看请求的 Authorization header"
    exit 1
fi

echo "=========================================="
echo "文件管理 API 测试"
echo "=========================================="
echo ""

# 测试函数
test_api() {
    local name="$1"
    local method="$2"
    local path="$3"
    local data="$4"

    echo "--- 测试: $name ---"
    if [ -n "$data" ]; then
        curl -s -X "$method" "$BASE_URL$path" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data" | jq .
    else
        curl -s -X "$method" "$BASE_URL$path" \
            -H "Authorization: Bearer $TOKEN" | jq .
    fi
    echo ""
}

# 1. 获取文件夹树
test_api "获取文件夹树" "GET" "/files/tree"

# 2. 创建文件夹
test_api "创建文件夹 (测试目录)" "POST" "/files/folder" '{"name":"测试目录","parentId":null}'

# 3. 获取文件夹树 (查看新创建的文件夹)
test_api "获取文件夹树 (查看新文件夹)" "GET" "/files/tree"

# 4. 获取文件列表
test_api "获取文件列表" "GET" "/files/list"

# 5. 检查上传状态 (秒传检测)
test_api "检查上传状态 (秒传检测)" "POST" "/files/upload/check" '{"fileHash":"test123456789","fileSize":1024,"fileName":"test.txt"}'

# 6. 初始化上传
test_api "初始化上传" "POST" "/files/upload/init" '{"fileHash":"test123456789","fileSize":1024,"fileName":"test.txt","contentType":"text/plain","totalChunks":1}'

echo "=========================================="
echo "测试完成"
echo "=========================================="
echo ""
echo "提示: 更多测试需要实际文件上传，请在前端界面进行完整测试"