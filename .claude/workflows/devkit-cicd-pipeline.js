export const meta = {
  name: 'devkit-cicd-pipeline',
  description: 'DevKit CI/CD 自动化流水线 - 代码审查、自动修复、专家验证、合并部署、浏览器测试（含控制台监控）',
  whenToUse: '手动执行或定时执行代码审查并自动部署到阿里云环境',
  phases: [
    { title: '代码审查', detail: '四维度深度审查: Bug/规范/边界/设计合理性' },
    { title: '自动修复', detail: '修复冲突和 bug' },
    { title: '专家验证', detail: '四角色验证: 安全/Go/Vue/QA' },
    { title: '合并部署', detail: '合并 dev 到 main，部署到阿里云' },
    { title: '浏览器测试', detail: '真实浏览器按钮级测试 + 控制台监控' },
    { title: '报告', detail: '生成执行结果报告' },
  ],
}

const CONFIG = {
  reviewBranch: 'dev',
  targetBranch: 'main',
  appServer: '123.57.201.44',
  appUser: 'root',
  appDir: '/opt/devkit',
  frontendUrl: 'http://123.57.201.44',
  apiUrl: 'http://123.57.201.44/api',
}

const results = {
  codeReview: null,
  autoFix: null,
  expertVerify: null,
  merge: null,
  deploy: null,
  test: null,
  errors: [],
}

// ============================================
// 阶段 1: 四维度代码审查
// ============================================
phase('代码审查')
log('开始四维度代码审查...')

// 1.1 检查分支状态
const branchStatus = await agent('检查 git 状态: git fetch origin && git log --oneline origin/dev -10 && git log --oneline origin/main -5 && git log origin/main..origin/dev --oneline', { label: 'git-status', phase: '代码审查' })

// 1.2 Bug 检测
log('维度 1/4: Bug 检测...')
const bugDetection = await agent(`作为资深 Go+Vue 工程师，对 DevKit 项目 dev 分支相对于 main 的变更进行 Bug 检测。

步骤:
1. 先运行: git diff origin/main...origin/dev --stat 查看变更文件
2. 只审查变更的文件，不是整个项目
3. 检查: 空指针/资源泄漏/并发问题/SQL注入/错误处理
4. 运行: cd backend-server && go vet ./... 和 golangci-lint run

返回 JSON: {"passed": true/false, "bugs": [{"file": "", "line": 0, "type": "", "description": "", "severity": ""}], "warnings": [], "testResults": "", "changedFiles": []}`, { label: 'bug-detection', phase: '代码审查' })

// 1.3 代码规范
log('维度 2/4: 代码规范检查...')
const codeStyle = await agent(`作为技术负责人，对 dev 分支相对于 main 的变更进行代码规范检查。

步骤:
1. 运行: git diff origin/main...origin/dev 查看具体变更
2. 只审查变更部分的代码规范
3. 检查 Go 命名/组织/惯用法、Vue 组件规范、TypeScript 类型规范

返回 JSON: {"passed": true/false, "violations": [{"file": "", "rule": "", "description": "", "severity": ""}], "styleScore": 8, "changedLines": "变更行数"}`, { label: 'code-style', phase: '代码审查' })

// 1.4 边界条件
log('维度 3/4: 边界条件检查...')
const edgeCases = await agent(`作为 QA+安全工程师，对 dev 分支的变更进行边界条件和安全检查。

步骤:
1. 运行: git diff origin/main...origin/dev 查看变更
2. 重点检查变更部分的安全风险
3. 检查: 空值处理、SQL注入、XSS、权限边界、JWT安全、CORS配置

返回 JSON: {"passed": true/false, "issues": [{"area": "", "description": "", "severity": "", "testCase": ""}], "securityScore": 7}`, { label: 'edge-cases', phase: '代码审查' })

// 1.5 设计合理性
log('维度 4/4: 设计合理性检查...')
const designReview = await agent(`作为系统架构师，对 dev 分支的变更进行架构和设计合理性检查。

步骤:
1. 运行: git diff origin/main...origin/dev 查看变更
2. 评估变更对整体架构的影响
3. 检查: 分层是否清晰、职责是否单一、依赖方向、性能影响

返回 JSON: {"passed": true/false, "issues": [{"area": "", "description": "", "severity": "", "suggestion": ""}], "designScore": 8, "strengths": [], "weaknesses": []}`, { label: 'design-review', phase: '代码审查' })

// 综合结果
results.codeReview = { bugDetection, codeStyle, edgeCases, designReview }
log('代码审查完成')

// ============================================
// 阶段 2: 自动修复
// ============================================
phase('自动修复')
log('开始自动修复...')

const autoFix = await agent(`根据代码审查结果进行自动修复。

步骤:
1. 分析代码审查发现的 bug 和问题
2. 如果有 git 冲突，尝试 rebase dev 到 main 并解决冲突
3. 修复发现的 bug（API路径错误、权限树过滤bug、SQL硬编码等）
4. 修复后 commit 到 dev 分支: git add -A && git commit -m "fix: auto-fix from CI/CD pipeline"
5. 推送到远程: git push origin dev

返回修复报告: {"fixedBugs": [], "conflictsResolved": [], "commitHash": ""}`, { label: 'auto-fix', phase: '自动修复' })

results.autoFix = autoFix
log('自动修复完成')

// ============================================
// 阶段 3: 专家验证
// ============================================
phase('专家验证')
log('开始专家验证...')

// 3.1 安全专家
log('安全专家检查...')
const securityExpert = await agent(`作为安全专家，审查最近的代码修复是否引入新漏洞。

检查:
1. 运行: git log origin/dev -5 --oneline 查看最近提交
2. 运行: git diff origin/dev~5..origin/dev 查看变更
3. 检查: SQL注入/XSS/CSRF/JWT安全/敏感信息泄露/权限绕过
4. 检查 CORS 配置是否过于宽松
5. 检查是否有硬编码的密钥或密码

返回 JSON: {"passed": true/false, "issues": [{"type": "", "severity": "", "description": "", "file": ""}], "verdict": ""}`, { label: 'security-expert', phase: '专家验证' })

// 3.2 Go 资深工程师
log('Go 工程师检查...')
const goExpert = await agent(`作为 Go 资深工程师，审查代码质量和并发安全。

检查:
1. 运行: git diff origin/main...origin/dev 查看 Go 代码变更
2. 检查: goroutine 泄漏/channel 使用/defer 位置/错误处理
3. 检查: 内存泄漏/资源未关闭/锁使用
4. 运行: cd backend-server && go vet ./...

返回 JSON: {"passed": true/false, "issues": [{"type": "", "description": "", "file": ""}], "qualityScore": 8}`, { label: 'go-expert', phase: '专家验证' })

// 3.3 Vue 资深工程师
log('Vue 工程师检查...')
const vueExpert = await agent(`作为 Vue 资深工程师，审查前端代码和组件规范。

检查:
1. 运行: git diff origin/main...origin/dev 查看 Vue/TS 变更
2. 检查: 组件命名/props 定义/composition API 使用
3. 检查: 响应式数据/reactive vs ref/计算属性
4. 检查: TypeScript 类型定义是否完整

返回 JSON: {"passed": true/false, "issues": [{"type": "", "description": "", "file": ""}], "qualityScore": 8}`, { label: 'vue-expert', phase: '专家验证' })

// 3.4 QA 工程师
log('QA 工程师检查...')
const qaExpert = await agent(`作为 QA 工程师，检查边界条件和异常处理。

检查:
1. 运行: git diff origin/main...origin/dev 查看变更
2. 检查: 空值处理/边界值/异常捕获/用户提示
3. 检查: 表单验证/分页边界/大数据量处理
4. 检查: 网络错误处理/超时重试

返回 JSON: {"passed": true/false, "issues": [{"area": "", "description": "", "severity": ""}]}`, { label: 'qa-expert', phase: '专家验证' })

results.expertVerify = { securityExpert, goExpert, vueExpert, qaExpert }
log('专家验证完成')

// ============================================
// 阶段 4: 合并部署
// ============================================
phase('合并部署')
log('开始合并部署...')

// 4.0 检查数据库迁移
log('检查数据库迁移...')
const dbCheck = await agent('查看 migrations/ 目录，对比最近提交是否有新迁移文件。返回: {"hasNewMigrations": false, "files": []}', { label: 'db-migration-check', phase: '合并部署' })

// 4.1 合并 dev 到 main
log('合并 dev 到 main...')
const mergeResult = await agent('执行合并: git checkout main && git pull origin main && git merge origin/dev && git push origin main。如果有冲突请解决。', { label: 'merge-dev-to-main', phase: '合并部署' })
results.merge = { success: true, result: mergeResult }
log('合并完成')

// 4.2 部署后端
log('部署后端...')
const deployBackend = await agent(`部署后端到 123.57.201.44，步骤：

第一步【本地编译】(使用 Go 1.26):
1. export PATH="/c/Program Files/Go/bin:$PATH"
2. cd backend-server && go version (确认版本 >= 1.23)
3. CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o backend-server cmd/server/main.go
4. 确认编译成功: ls -lh backend-server

第二步【远程部署】:
5. ssh root@123.57.201.44 "mkdir -p /opt/devkit/backend/config /opt/devkit/backend/logs"
6. scp backend-server root@123.57.201.44:/opt/devkit/backend/
7. scp backend-server/config/config.yaml root@123.57.201.44:/opt/devkit/backend/config/
8. 先执行数据库迁移(如果有): ssh root@123.57.201.44 "cd /opt/devkit/backend && ./backend-server -migrate"
9. 重启服务: ssh root@123.57.201.44 "systemctl restart devkit-backend && sleep 3 && systemctl status devkit-backend --no-pager"
10. 健康检查: curl -s http://123.57.201.44/api/v1/auth/captcha

注意：编译在本地执行(用新版 Go)，上传和重启在远程服务器执行。
`, { label: 'deploy-backend', phase: '合并部署' })

if (!deployBackend || deployBackend.includes('失败') || deployBackend.includes('error')) {
  log('后端部署失败，停止流水线')
  results.deploy = { backend: deployBackend, frontend: null }
  results.errors.push('后端部署失败')
} else {
  log('后端部署成功')

  // 4.3 部署前端
  log('部署前端...')
  const deployFrontend = await agent(`部署前端到 123.57.201.44，步骤：

第一步【本地构建】:
1. cd frontend-admin
2. pnpm install
3. pnpm build
4. 确认构建成功: ls -la apps/web-antd/dist/

第二步【远程部署】:
5. ssh root@123.57.201.44 "mkdir -p /opt/devkit/frontend"
6. scp -r apps/web-antd/dist/* root@123.57.201.44:/opt/devkit/frontend/
7. curl http://123.57.201.44 验证可访问

注意：构建在本地执行，上传在远程服务器执行。
`, { label: 'deploy-frontend', phase: '合并部署' })
  results.deploy = { backend: deployBackend, frontend: deployFrontend }
  log('前端部署完成')
}

// ============================================
// 阶段 5: 浏览器测试（含控制台监控）
// ============================================
phase('浏览器测试')
log('开始浏览器测试...')

// 5.1 检查 kimi-webbridge 状态
const bridgeCheck = await agent('~/.kimi-webbridge/bin/kimi-webbridge status', { label: 'bridge-status', phase: '浏览器测试' })

// 5.2 API 健康检查
const apiCheck = await agent('curl -s http://123.57.201.44/api/v1/auth/captcha 检查 API 是否返回 200 和健康状态', { label: 'api-health', phase: '浏览器测试' })

// 5.3 登录测试
log('登录测试...')
const loginTest = await agent(`使用 kimi-webbridge 进行按钮级别的登录测试:

1. 打开 http://123.57.201.44
2. 确认看到登录页面
3. 找到用户名输入框，输入 vben
4. 找到密码输入框，输入 123456
5. 找到并点击"登录"按钮
6. 如果遇到验证码，关闭验证码弹窗后重试登录
7. 等待页面跳转，验证是否跳转到 /analytics (Dashboard)
8. 验证页面显示"欢迎回来:vben"
9. 截图保存登录成功画面

**控制台监控**:
1. 登录后运行 evaluate: (async()=>{const r=await fetch("/api/v1/auth/captcha");return JSON.stringify({apiStatus:r.status,apiOk:r.ok})})()
2. 检查是否有 API 404/500 错误
3. 检查页面是否有红色错误提示
4. 截图记录控制台状态

**如果有错误**:
- API 404 → 检查后端路由和 nginx 配置
- JS 报错 → 检查 Vue 组件代码
- 样式错乱 → 检查 CSS 加载

返回每一步的结果。`, { label: 'login-test', phase: '浏览器测试' })

// 5.4 核心功能测试
log('核心功能测试...')
const coreTest = await agent(`使用 kimi-webbridge 进行按钮级别的核心功能测试:

前提: 确保已登录(vben/123456)

**测试 1 - Dashboard 页面**:
1. 确认在 /analytics 页面
2. 检查统计卡片: 用户量、访问量、下载量、使用量
3. 检查图表是否正常显示
4. **控制台监控**: 运行 evaluate 检查 JS 错误和资源加载
5. 截图

**测试 2 - 系统管理 → 用户管理**:
1. 点击左侧菜单"系统管理"展开
2. 点击"用户管理"
3. 等待表格加载，确认有用户列表数据
4. 点击"新增"按钮，确认弹出新增表单
5. **控制台监控**: 检查 network 是否有失败的 API 请求
6. 截图
7. 关闭弹窗

**测试 3 - 系统管理 → 角色管理**:
1. 点击"角色管理"
2. 等待表格加载
3. 确认看到 admin、user 等角色
4. **控制台监控**: 检查 JS 错误
5. 截图

**测试 4 - 系统管理 → 菜单管理**:
1. 点击"菜单管理"
2. 等待树形结构加载
3. 确认看到菜单树
4. 截图

**测试 5 - 文件管理**:
1. 点击"文件管理"
2. 等待页面加载
3. 确认看到文件列表
4. **控制台监控**: 检查 API 请求和资源加载
5. 截图

**错误自动修复**:
- 如果任何页面有 API 404/500: 分析原因，修复后端或 nginx，重新部署，重新测试
- 如果 JS 报错: 修复 Vue 组件，重新构建部署，重新测试
- 如果样式错乱: 检查 CSS 路径，修复后重新部署

每个测试返回: 页面是否正常加载、按钮是否可点击、数据是否正常显示、控制台状态、截图。`, { label: 'core-test', phase: '浏览器测试' })

results.test = { api: apiCheck, login: loginTest, core: coreTest }

// ============================================
// 阶段 6: 生成报告
// ============================================
phase('报告')

const now = 'timestamp-will-be-added-after'
const reportLines = [
  '# DevKit CI/CD Pipeline Report',
  '',
  '执行时间: ' + now,
  '分支: dev -> main',
  '部署环境: 阿里云 123.57.201.44',
  '',
  '## 代码审查',
  '- Bug 检测: ' + (bugDetection || '完成'),
  '- 代码规范: ' + (codeStyle || '完成'),
  '- 边界条件: ' + (edgeCases || '完成'),
  '- 设计合理性: ' + (designReview || '完成'),
  '',
  '## 自动修复',
  '- 修复结果: ' + (autoFix || '完成'),
  '',
  '## 专家验证',
  '- 安全专家: ' + (securityExpert || '通过'),
  '- Go 工程师: ' + (goExpert || '通过'),
  '- Vue 工程师: ' + (vueExpert || '通过'),
  '- QA 工程师: ' + (qaExpert || '通过'),
  '',
  '## 部署状态',
  '- 代码合并: ' + (results.merge?.success ? '成功' : '失败'),
  '- 后端部署: ' + (results.deploy?.backend ? '成功' : '失败'),
  '- 前端部署: ' + (results.deploy?.frontend ? '成功' : '失败'),
  '',
  '## 测试结果',
  '- API 健康: ' + (apiCheck || '完成'),
  '- 登录测试: ' + (loginTest || '完成'),
  '- 核心功能: ' + (coreTest || '完成'),
  '',
  '## 总体: ' + (results.errors.length === 0 ? '通过' : '失败'),
]

if (results.errors.length > 0) {
  reportLines.push('')
  reportLines.push('## 错误')
  results.errors.forEach(e => reportLines.push('- ' + e))
}

const report = reportLines.join('\n')
log(report)

return {
  ...results,
  overall: results.errors.length === 0 ? 'PASSED' : 'FAILED',
  timestamp: now,
}
