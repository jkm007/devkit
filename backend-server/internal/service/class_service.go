package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

// ClassService 班级服务
type ClassService struct {
	repo     *repository.ClassRepo
	userRepo *repository.UserRepo
}

// NewClassService 创建班级服务
func NewClassService() *ClassService {
	db := database.GetMySQL()
	return &ClassService{
		repo:     repository.NewClassRepo(db),
		userRepo: repository.NewUserRepo(db),
	}
}

// CreateClassRequest 创建班级请求
type CreateClassRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateClassRequest 更新班级请求
type UpdateClassRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      *int   `json:"status"`
}

// ClassResponse 班级响应
type ClassResponse struct {
	model.Class
	MemberCount int64  `json:"memberCount"`
	CreatorName string `json:"creatorName"`
	MyRole      string `json:"myRole,omitempty"`
}

// MemberResponse 班级成员响应
type MemberResponse struct {
	ID        uint                 `json:"id"`
	UserID    uint                 `json:"userId"`
	Nickname  string               `json:"nickname"`
	Username  string               `json:"username"`
	Avatar    string               `json:"avatar"`
	Role      model.ClassMemberRole `json:"role"`
	Status    int                  `json:"status"`
	JoinedAt  time.Time            `json:"joinedAt"`
}

// CreateInvitationRequest 创建邀请码请求
type CreateInvitationRequest struct {
	ExpireAt *time.Time `json:"expireAt"`
	MaxUses  int        `json:"maxUses"`
}

// JoinByCodeRequest 通过邀请码加入班级请求
type JoinByCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

// CreateClass 创建班级
func (s *ClassService) CreateClass(userID uint, req *CreateClassRequest) (*model.Class, error) {
	code, err := s.generateUniqueCode()
	if err != nil {
		return nil, err
	}

	class := &model.Class{
		Name:        req.Name,
		Code:        code,
		Description: req.Description,
		Status:      1,
		CreatedBy:   userID,
	}

	if err := s.repo.Create(class); err != nil {
		return nil, err
	}

	// 创建者默认成为班主任
	member := &model.ClassMember{
		ClassID: class.ID,
		UserID:  userID,
		Role:    model.ClassRoleTeacher,
		Status:  1,
	}
	if err := s.repo.AddMember(member); err != nil {
		return nil, err
	}

	return class, nil
}

// UpdateClass 更新班级
func (s *ClassService) UpdateClass(userID uint, classID uint, req *UpdateClassRequest) error {
	if err := s.checkPermission(userID, classID, model.ClassRoleTeacher); err != nil {
		return err
	}

	class, err := s.repo.GetByID(classID)
	if err != nil {
		return err
	}

	if req.Name != "" {
		class.Name = req.Name
	}
	if req.Description != "" {
		class.Description = req.Description
	}
	if req.Status != nil {
		class.Status = *req.Status
	}

	return s.repo.Update(class)
}

// DeleteClass 删除班级
func (s *ClassService) DeleteClass(userID uint, classID uint) error {
	if err := s.checkPermission(userID, classID, model.ClassRoleTeacher); err != nil {
		return err
	}
	return s.repo.DeleteClassWithRelations(classID)
}

// GetClassDetail 获取班级详情
func (s *ClassService) GetClassDetail(userID uint, classID uint) (*ClassResponse, error) {
	class, err := s.repo.GetByID(classID)
	if err != nil {
		return nil, err
	}

	// 校验用户是否有权查看班级（成员或创建者）
	role, err := s.getUserRoleInClass(userID, classID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && class.CreatedBy != userID {
		return nil, errors.New("无权查看该班级")
	}

	memberCount, _ := s.repo.CountMembers(classID)
	creatorName := ""
	if creator, err := s.userRepo.GetByID(class.CreatedBy); err == nil {
		creatorName = creator.Nickname
		if creatorName == "" {
			creatorName = creator.Name
		}
	}

	return &ClassResponse{
		Class:       *class,
		MemberCount: memberCount,
		CreatorName: creatorName,
		MyRole:      string(role),
	}, nil
}

// ListClasses 管理端班级列表
func (s *ClassService) ListClasses(page, pageSize int, keyword string) ([]ClassResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	items, total, err := s.repo.List(page, pageSize, keyword)
	if err != nil {
		return nil, 0, err
	}

	results := make([]ClassResponse, 0, len(items))
	for _, class := range items {
		memberCount, _ := s.repo.CountMembers(class.ID)
		creatorName := ""
		if creator, err := s.userRepo.GetByID(class.CreatedBy); err == nil {
			creatorName = creator.Nickname
			if creatorName == "" {
				creatorName = creator.Name
			}
		}
		results = append(results, ClassResponse{
			Class:       class,
			MemberCount: memberCount,
			CreatorName: creatorName,
		})
	}

	return results, total, nil
}

// ListMyClasses 获取我加入/创建的班级列表
func (s *ClassService) ListMyClasses(userID uint) ([]ClassResponse, error) {
	classIDs, err := s.repo.ListClassIDsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 也包含创建但未加入的班级（理论上创建时已加入，兼容处理）
	createdIDs, err := s.getCreatedClassIDs(userID)
	if err != nil {
		return nil, err
	}
	classIDs = mergeUnique(classIDs, createdIDs)

	if len(classIDs) == 0 {
		return []ClassResponse{}, nil
	}

	var classes []model.Class
	if err := s.repo.DB().Where("id IN ?", classIDs).Order("created_at DESC").Find(&classes).Error; err != nil {
		return nil, err
	}

	results := make([]ClassResponse, 0, len(classes))
	for _, class := range classes {
		memberCount, _ := s.repo.CountMembers(class.ID)
		role, _ := s.getUserRoleInClass(userID, class.ID)
		results = append(results, ClassResponse{
			Class:       class,
			MemberCount: memberCount,
			MyRole:      string(role),
		})
	}

	return results, nil
}

// ListMembers 获取班级成员列表
func (s *ClassService) ListMembers(userID uint, classID uint, page, pageSize int) ([]MemberResponse, int64, error) {
	if err := s.checkPermission(userID, classID, model.ClassRoleStudent); err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	members, total, err := s.repo.ListMembersByClassID(classID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	results := make([]MemberResponse, 0, len(members))
	for _, m := range members {
		resp := MemberResponse{
			ID:       m.ID,
			UserID:   m.UserID,
			Role:     m.Role,
			Status:   m.Status,
			JoinedAt: m.JoinedAt,
		}
		if user, err := s.userRepo.GetByID(m.UserID); err == nil {
			resp.Nickname = user.Nickname
			if resp.Nickname == "" {
				resp.Nickname = user.Name
			}
			resp.Username = user.Name
			resp.Avatar = user.Avatar
		}
		results = append(results, resp)
	}

	return results, total, nil
}

// AddMember 添加班级成员
func (s *ClassService) AddMember(operatorID uint, classID uint, targetUserID uint, role model.ClassMemberRole) error {
	// 只有老师可以添加 monitor/teacher；monitor/teacher 可以添加 student
	requiredRole := model.ClassRoleTeacher
	if role == model.ClassRoleStudent {
		requiredRole = model.ClassRoleMonitor
	}
	if err := s.checkPermission(operatorID, classID, requiredRole); err != nil {
		return err
	}

	// 检查目标用户是否已在班级中
	if _, err := s.repo.GetMemberByUserID(classID, targetUserID); err == nil {
		return errors.New("该用户已在班级中")
	}

	member := &model.ClassMember{
		ClassID: classID,
		UserID:  targetUserID,
		Role:    role,
		Status:  1,
	}
	return s.repo.AddMember(member)
}

// UpdateMemberRole 更新成员角色
func (s *ClassService) UpdateMemberRole(operatorID uint, classID uint, memberID uint, role model.ClassMemberRole) error {
	if err := s.checkPermission(operatorID, classID, model.ClassRoleTeacher); err != nil {
		return err
	}
	return s.repo.UpdateMemberRole(memberID, role)
}

// RemoveMember 移除成员
func (s *ClassService) RemoveMember(operatorID uint, classID uint, memberID uint) error {
	// 获取操作者角色
	operatorRole, err := s.getUserRoleInClass(operatorID, classID)
	if err != nil {
		return errors.New("无权操作")
	}

	// 获取目标成员角色
	var members []model.ClassMember
	if err := s.repo.DB().Where("id = ? AND class_id = ?", memberID, classID).Find(&members).Error; err != nil {
		return err
	}
	if len(members) == 0 {
		return errors.New("成员不存在")
	}
	targetRole := members[0].Role

	// monitor 只能移除 student；teacher 可以移除所有人
	if operatorRole == model.ClassRoleMonitor && targetRole != model.ClassRoleStudent {
		return errors.New("无权移除该角色")
	}
	if operatorRole == model.ClassRoleStudent {
		return errors.New("无权移除成员")
	}

	return s.repo.RemoveMember(memberID)
}

// CreateInvitation 创建邀请码
func (s *ClassService) CreateInvitation(userID uint, classID uint, req *CreateInvitationRequest) (*model.ClassInvitation, error) {
	if err := s.checkPermission(userID, classID, model.ClassRoleMonitor); err != nil {
		return nil, err
	}

	code, err := s.generateUniqueCode()
	if err != nil {
		return nil, err
	}

	invitation := &model.ClassInvitation{
		ClassID: classID,
		Code:    code,
		ExpireAt: req.ExpireAt,
		MaxUses: req.MaxUses,
		Status:  1,
		CreatedBy: userID,
	}
	if err := s.repo.CreateInvitation(invitation); err != nil {
		return nil, err
	}
	return invitation, nil
}

// DisableInvitation 禁用邀请码
func (s *ClassService) DisableInvitation(userID uint, invitationID uint) error {
	// 需要知道邀请码所属班级并校验权限
	var invitation model.ClassInvitation
	if err := s.repo.DB().Where("id = ?", invitationID).First(&invitation).Error; err != nil {
		return err
	}
	if err := s.checkPermission(userID, invitation.ClassID, model.ClassRoleMonitor); err != nil {
		return err
	}
	return s.repo.DisableInvitation(invitationID)
}

// JoinByCode 通过邀请码加入班级
func (s *ClassService) JoinByCode(userID uint, code string) (*model.Class, error) {
	invitation, err := s.repo.GetInvitationByCode(code)
	if err != nil {
		return nil, errors.New("邀请码无效")
	}

	// 校验邀请码状态
	if invitation.Status != 1 {
		return nil, errors.New("邀请码已失效")
	}

	// 校验是否过期
	if invitation.ExpireAt != nil && invitation.ExpireAt.Before(time.Now()) {
		return nil, errors.New("邀请码已过期")
	}

	// 校验使用次数
	if invitation.MaxUses > 0 && invitation.UsedCount >= invitation.MaxUses {
		return nil, errors.New("邀请码使用次数已达上限")
	}

	class, err := s.repo.GetByID(invitation.ClassID)
	if err != nil {
		return nil, errors.New("班级不存在")
	}
	if class.Status != 1 {
		return nil, errors.New("班级已禁用")
	}

	// 检查是否已在班级中
	if _, err := s.repo.GetMemberByUserID(class.ID, userID); err == nil {
		return class, nil // 已在班级中，幂等返回
	}

	// 添加为学生成员
	member := &model.ClassMember{
		ClassID: class.ID,
		UserID:  userID,
		Role:    model.ClassRoleStudent,
		Status:  1,
	}
	if err := s.repo.AddMember(member); err != nil {
		return nil, err
	}

	// 增加使用次数
	_ = s.repo.IncrementInvitationUsedCount(invitation.ID)

	return class, nil
}

// ListInvitations 获取班级邀请码列表
func (s *ClassService) ListInvitations(userID uint, classID uint) ([]model.ClassInvitation, error) {
	if err := s.checkPermission(userID, classID, model.ClassRoleMonitor); err != nil {
		return nil, err
	}
	return s.repo.ListInvitationsByClassID(classID)
}

// GetUserClassIDs 获取用户加入的班级ID列表
func (s *ClassService) GetUserClassIDs(userID uint) ([]uint, error) {
	return s.repo.ListClassIDsByUserID(userID)
}

// CheckClassPermission 公开方法：检查用户在班级中是否具备至少指定角色权限
func (s *ClassService) CheckClassPermission(userID uint, classID uint, requiredRole model.ClassMemberRole) error {
	return s.checkPermission(userID, classID, requiredRole)
}

// checkPermission 检查用户在班级中是否具备至少指定角色权限
func (s *ClassService) checkPermission(userID uint, classID uint, requiredRole model.ClassMemberRole) error {
	role, err := s.getUserRoleInClass(userID, classID)
	if err != nil {
		return errors.New("无权操作")
	}

	roleLevel := map[model.ClassMemberRole]int{
		model.ClassRoleStudent: 1,
		model.ClassRoleMonitor: 2,
		model.ClassRoleTeacher: 3,
	}

	if roleLevel[role] < roleLevel[requiredRole] {
		return errors.New("无权操作")
	}
	return nil
}

// getUserRoleInClass 获取用户在班级中的角色
func (s *ClassService) getUserRoleInClass(userID uint, classID uint) (model.ClassMemberRole, error) {
	member, err := s.repo.GetMemberByUserID(classID, userID)
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

// getCreatedClassIDs 获取用户创建的班级ID
func (s *ClassService) getCreatedClassIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := s.repo.DB().Model(&model.Class{}).Where("created_by = ?", userID).Pluck("id", &ids).Error
	return ids, err
}

// generateUniqueCode 生成唯一 6 位数字邀请码
func (s *ClassService) generateUniqueCode() (string, error) {
	for i := 0; i < 10; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(900000))
		if err != nil {
			return "", err
		}
		code := fmt.Sprintf("%06d", n.Int64()+100000)
		if !s.repo.CodeExists(code) {
			return code, nil
		}
	}
	return "", errors.New("无法生成唯一邀请码")
}

// mergeUnique 合并去重
func mergeUnique(a, b []uint) []uint {
	seen := make(map[uint]bool)
	var result []uint
	for _, v := range a {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	for _, v := range b {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
