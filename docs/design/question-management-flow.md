# 题目管理完整流程设计

## 一、状态流转图

```
                              ┌──────────────────────────────────┐
                              │          新建题目                 │
                              │  resourceType 选择决定后续流程     │
                              └──────────────┬───────────────────┘
                                             │
                                             ▼
                                      ┌──────────┐
                                      │   草稿    │
                                      │  draft    │
                                      └─────┬────┘
                                            │
                    ┌───────────────────────┼───────────────────────┐
                    │                       │                       │
            resourceType             resourceType             resourceType
            = private                = public                 = group / user
            (仅自己可见)              (所有人可见)              (分组/指定用户)
                    │                       │                       │
                    ▼                       ▼                       ▼
            ┌──────────┐            ┌──────────┐            ┌──────────┐
            │  直接发布  │            │  提交审核  │            │  提交审核  │
            └─────┬────┘            └─────┬────┘            └─────┬────┘
                  │                       │                       │
                  ▼                       ▼                       ▼
            ┌──────────┐            ┌──────────┐            ┌──────────┐
            │  已发布    │            │  待审核    │            │  待审核    │
            │ published │            │  pending  │            │  pending  │
            └─────┬────┘            └─────┬────┘            └─────┬────┘
                  │                       │                       │
                  │               ┌───────┴───────┐               │
                  │               │               │               │
                  │               ▼               ▼               │
                  │         ┌──────────┐   ┌──────────┐          │
                  │         │ 审核驳回  │   │ 审核通过  │          │
                  │         │ rejected │   │ approved │          │
                  │         └─────┬────┘   └─────┬────┘          │
                  │               │               │               │
                  │               │ 编辑后        │ 点击发布       │
                  │               │ 重新提交       │               │
                  │               │               ▼               │
                  │               └──────────→┌──────────┐       │
                  │                           │  已发布    │←──────┘
                  │                           │ published │
                  │                           └─────┬────┘
                  │                                 │
                  │ 下架                    下架     │
                  ▼                                 ▼
            ┌──────────┐                     ┌──────────┐
            │  已下架    │                     │  已下架    │
            │ archived │                     │ archived │
            └──────────┘                     └──────────┘
```

## 二、状态定义

| 状态值 | 中文名 | 含义 |
|--------|--------|------|
| draft | 草稿 | 新建或编辑后，未提交 |
| pending | 待审核 | 已提交审核，等待审核人处理 |
| approved | 审核通过 | 审核通过，等待发布 |
| rejected | 已驳回 | 审核未通过，需修改 |
| published | 已发布 | 已对外发布可见 |
| archived | 已下架 | 已下架，不可见 |

## 三、状态 × 操作矩阵

```
操作            │ 草稿    │ 待审核  │ 审核通过 │ 已驳回  │ 已发布  │ 已下架
────────────────┼─────────┼─────────┼─────────┼─────────┼─────────┼────────
预览            │   ✅    │   ✅    │   ✅    │   ✅    │   ✅    │   ✅
编辑            │   ✅    │   ❌    │   ❌    │   ✅    │   ❌    │   ❌
提交审核        │   ✅ ¹  │   ❌    │   ❌    │   ✅    │   ❌    │   ❌
审核通过        │   ❌    │   ✅    │   ❌    │   ❌    │   ❌    │   ❌
审核驳回        │   ❌    │   ✅    │   ❌    │   ❌    │   ❌    │   ❌
撤回到草稿      │   ❌    │   ✅    │   ✅    │   ❌    │   ✅    │   ❌
发布            │   ✅ ¹  │   ❌    │   ✅    │   ❌    │   ❌    │   ❌
下架            │   ❌    │   ❌    │   ❌    │   ❌    │   ✅    │   ❌
重新上架        │   ❌    │   ❌    │   ❌    │   ❌    │   ❌    │   ✅
删除            │   ✅    │   ❌    │   ❌    │   ✅    │   ❌    │   ✅

¹ 仅私有题目(resourceType=private)显示"发布"按钮，其他类型显示"提交审核"
```

## 四、按钮显示逻辑（前端）

```javascript
// 根据 status + resourceType 决定显示哪些按钮
function getActions(row) {
  const actions = [];
  const { status, resourceType } = row;
  const isPrivate = resourceType === 'private';

  // 预览 - 所有状态都有
  actions.push({ text: '预览', icon: 'eye' });

  switch (status) {
    case 'draft':
      actions.push({ text: '编辑', icon: 'edit', auth: 'question:edit' });
      if (isPrivate) {
        // 私有题目：直接发布
        actions.push({ text: '发布', icon: 'check-circle', auth: 'question:publish' });
      } else {
        // 公共题目：提交审核
        actions.push({ text: '提交审核', icon: 'send', auth: 'question:audit:submit' });
      }
      actions.push({ text: '删除', icon: 'trash', auth: 'question:delete', danger: true });
      break;

    case 'pending':
      actions.push({ text: '审核通过', icon: 'check', auth: 'question:audit:approve' });
      actions.push({ text: '审核驳回', icon: 'x', auth: 'question:audit:reject' });
      actions.push({ text: '撤回到草稿', icon: 'undo', auth: 'question:publish' });
      break;

    case 'approved':
      actions.push({ text: '发布', icon: 'check-circle', auth: 'question:publish' });
      actions.push({ text: '撤回到草稿', icon: 'undo', auth: 'question:publish' });
      break;

    case 'rejected':
      actions.push({ text: '编辑', icon: 'edit', auth: 'question:edit' });
      actions.push({ text: '重新提交审核', icon: 'send', auth: 'question:audit:submit' });
      actions.push({ text: '删除', icon: 'trash', auth: 'question:delete', danger: true });
      break;

    case 'published':
      actions.push({ text: '下架', icon: 'archive', auth: 'question:publish' });
      actions.push({ text: '撤回到草稿', icon: 'undo', auth: 'question:publish' });
      break;

    case 'archived':
      actions.push({ text: '重新上架', icon: 'refresh', auth: 'question:publish' });
      actions.push({ text: '删除', icon: 'trash', auth: 'question:delete', danger: true });
      break;
  }

  return actions;
}
```

## 五、后端 API 设计

### 5.1 现有 API（保留）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/system/questions | 题目列表（分页、筛选） |
| GET | /api/v1/system/questions/:id | 题目详情 |
| POST | /api/v1/system/questions | 新建题目 |
| PUT | /api/v1/system/questions/:id | 编辑题目 |
| DELETE | /api/v1/system/questions/:id | 删除题目 |

### 5.2 需要修改的 API

| 方法 | 路径 | 说明 | 修改内容 |
|------|------|------|---------|
| POST | /api/v1/system/questions/:id/publish | 发布 | 增加状态校验：draft(private) / approved 可发布 |
| POST | /api/v1/system/questions/:id/archive | 下架 | 仅 published 可下架 |
| POST | /api/v1/system/questions/:id/submit-audit | 提交审核 | draft/rejected → pending |
| POST | /api/v1/system/questions/:id/audit/approve | 审核通过 | pending → approved（非直接发布） |
| POST | /api/v1/system/questions/:id/audit/reject | 审核驳回 | pending → rejected |

### 5.3 需要新增的 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/system/questions/:id/withdraw | 撤回到草稿 |
| POST | /api/v1/system/questions/:id/reactivate | 重新上架（archived → published） |

### 5.4 状态校验规则（后端）

```go
// 状态流转映射
var allowedTransitions = map[string][]string{
    "draft":     {"pending", "published"},  // published 仅 private
    "pending":   {"approved", "rejected", "draft"},
    "approved":  {"published", "draft"},
    "rejected":  {"pending"},
    "published": {"archived", "draft"},
    "archived":  {"published"},
}

// publish 接口校验
func (s *QuestionService) Publish(id uint) error {
    question, _ := s.repo.GetByID(id)

    // 私有题目：draft 直接发布
    if question.ResourceType == "private" && question.Status == "draft" {
        question.Status = "published"
        return s.repo.Update(question)
    }

    // 其他类型：必须审核通过后才能发布
    if question.Status != "approved" {
        return fmt.Errorf("当前状态【%s】不允许发布，请先提交审核", question.Status)
    }

    question.Status = "published"
    return s.repo.Update(question)
}

// approve 接口：审核通过 → approved（不是直接 published）
func (s *QuestionService) Approve(id uint) error {
    question, _ := s.repo.GetByID(id)
    if question.Status != "pending" {
        return fmt.Errorf("当前状态【%s】不允许审核", question.Status)
    }
    question.Status = "approved"
    return s.repo.Update(question)
}
```

## 六、权限设计

| 权限码 | 说明 | 对应操作 |
|--------|------|---------|
| question:view | 查看题目列表 | 列表页访问 |
| question:create | 新建题目 | 新建按钮 |
| question:edit | 编辑题目 | 编辑按钮 |
| question:delete | 删除题目 | 删除按钮 |
| question:publish | 发布/下架/上架/撤回 | 发布、下架、重新上架、撤回到草稿 |
| question:audit:submit | 提交审核 | 提交审核按钮 |
| question:audit:view | 审核查看 | 审核页面访问 |
| question:audit:approve | 审核通过 | 审核通过按钮 |
| question:audit:reject | 审核驳回 | 审核驳回按钮 |

## 七、数据模型（qb_questions 表相关字段）

```
字段                类型           说明
──────────────────────────────────────────────
id                  BIGINT        主键
title               VARCHAR       题目标题
question_type       VARCHAR       题型
stem                JSON          题干（富文本）
content             JSON          选项内容（选择题）
answer              JSON          答案
analysis            JSON          解析
materials           JSON          材料
score_rule          JSON          评分规则
exam_id             BIGINT        考试ID
subject_id          BIGINT        科目ID
category_id         BIGINT        章节分类ID
source_id           BIGINT        来源ID
difficulty          INT           难度 1/2/3
resource_type       VARCHAR       资源类型 public/private/group/user
status              VARCHAR       状态 draft/pending/approved/rejected/published/archived
created_by          BIGINT        创建人
updated_by          BIGINT        更新人
reviewed_by         BIGINT        审核人
reviewed_at         DATETIME      审核时间
reject_reason       VARCHAR       驳回原因
published_at        DATETIME      发布时间
created_at          DATETIME      创建时间
updated_at          DATETIME      更新时间
```

## 八、表单设计（新增/编辑）

```
┌─────────────────────────────────────────────────────┐
│  新增/编辑题目                                        │
├─────────────────────────────────────────────────────┤
│                                                     │
│  题目标题:  [________________________]  *必填        │
│                                                     │
│  题    型:  [单选题 ▾]                 *必填        │
│                                                     │
│  难    度:  [简单 ▾]                                │
│                                                     │
│  资源类型:  [私有 ▾]                   决定审核流程   │
│                                                     │
│  ── 分类选择（4级联动）──                            │
│  考试大类:  [软考 ▾]                                │
│  具体考试:  [软件设计师 ▾]                           │
│  科    目:  [综合知识 ▾]                             │
│  章节分类:  [计算机网络 ▾]  可选                      │
│                                                     │
│  ── 题干内容 ──                                      │
│  ┌─────────────────────────────────────────┐        │
│  │  富文本编辑器（支持图片/视频）             │        │
│  │                                         │        │
│  └─────────────────────────────────────────┘        │
│                                                     │
│  ── 选项（选择题时显示）──                            │
│  ○ A [______________]                               │
│  ○ B [______________]                               │
│  ○ C [______________]                               │
│  ○ D [______________]                               │
│  [+ 添加选项]                                       │
│                                                     │
│  ── 答案 ──                                         │
│  单选/多选: 点击选项标记                              │
│  判断题: 正确 ○  错误 ○                              │
│  填空/简答: 富文本编辑器                              │
│                                                     │
│  ── 文字解析 ──                                      │
│  ┌─────────────────────────────────────────┐        │
│  │  富文本编辑器                              │        │
│  └─────────────────────────────────────────┘        │
│                                                     │
│  ── 图片/视频解析 ──                                 │
│  ┌─────────────────────────────────────────┐        │
│  │  富文本编辑器（支持图片/视频上传）          │        │
│  └─────────────────────────────────────────┘        │
│                                                     │
│  ── 材料（可选）──                                   │
│  ┌─────────────────────────────────────────┐        │
│  │  富文本编辑器（材料题的附加材料）           │        │
│  └─────────────────────────────────────────┘        │
│                                                     │
├─────────────────────────────────────────────────────┤
│                           [取消]  [保存草稿]  [提交]  │
└─────────────────────────────────────────────────────┘

提交按钮逻辑:
- 私有题目 → 直接发布
- 其他类型 → 提交审核
- 编辑已驳回题目 → 重新提交审核
```

## 九、列表页筛选条件

```
┌─────────────────────────────────────────────────────────────┐
│  [题目标题______] [题型▾] [状态▾] [难度▾] [搜索] [重置]      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ID │ 题目标题    │ 题型   │ 难度 │ 资源类型 │ 状态   │ 操作  │
│  ───┼────────────┼───────┼──────┼─────────┼───────┼──────│
│  1  │ TCP三次握手 │ 单选题 │ 中等 │ 公共    │ 已发布 │ ···  │
│  2  │ HTTP状态码  │ 判断题 │ 简单 │ 私有    │ 草稿   │ ···  │
│  3  │ 数据库索引  │ 简答题 │ 困难 │ 公共    │ 待审核 │ ···  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  状态标签颜色:                                        │   │
│  │  草稿=default  待审核=processing  审核通过=success    │   │
│  │  已驳回=error   已发布=success    已下架=warning      │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                             │
│  [批量删除]  [批量发布]         共 128 条  < 1 2 3 ... 13 >  │
└─────────────────────────────────────────────────────────────┘
```

## 十、完整时序图（新建题目 → 发布）

```
用户(作者)                    系统                     审核人
    │                          │                        │
    │  1. 点击"新增题目"        │                        │
    │─────────────────────────→│                        │
    │                          │                        │
    │  2. 填写表单              │                        │
    │  选择 resourceType=public │                        │
    │─────────────────────────→│                        │
    │                          │                        │
    │  3. 点击"提交"            │                        │
    │─────────────────────────→│                        │
    │                          │                        │
    │                          │  创建题目              │
    │                          │  status=draft          │
    │                          │                        │
    │                          │  因为 public 类型       │
    │                          │  自动提交审核           │
    │                          │  status → pending      │
    │                          │                        │
    │                          │  4. 通知审核人          │
    │                          │───────────────────────→│
    │                          │                        │
    │                          │  5. 审核人查看题目      │
    │                          │←───────────────────────│
    │                          │                        │
    │                          │  6. 审核通过            │
    │                          │←───────────────────────│
    │                          │                        │
    │                          │  status → approved     │
    │                          │                        │
    │  7. 通知作者审核通过       │                        │
    │←─────────────────────────│                        │
    │                          │                        │
    │  8. 作者点击"发布"        │                        │
    │─────────────────────────→│                        │
    │                          │                        │
    │                          │  status → published    │
    │                          │  published_at = now    │
    │                          │                        │
    │  9. 题目对外可见           │                        │
    │←─────────────────────────│                        │
    │                          │                        │
```

## 十一、私有题目时序图（跳过审核）

```
用户(作者)                    系统
    │                          │
    │  1. 新增题目              │
    │  resourceType = private   │
    │─────────────────────────→│
    │                          │
    │  2. 填写完成，点击"发布"   │
    │─────────────────────────→│
    │                          │
    │                          │  创建题目
    │                          │  status = draft
    │                          │
    │                          │  检测: private 类型
    │                          │  跳过审核，直接发布
    │                          │  status → published
    │                          │
    │  3. 发布成功              │
    │←─────────────────────────│
    │                          │
    │  4. 仅自己可见            │
```
