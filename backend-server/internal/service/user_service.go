package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/cache"
	"backend-server/pkg/database"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
)

// validateEmail 校验邮箱格式
func validateEmail(email string) bool {
	if email == "" {
		return true // 空值允许
	}
	return emailRegex.MatchString(email)
}

// validatePhone 校验手机号格式
func validatePhone(phone string) bool {
	if phone == "" {
		return true // 空值允许
	}
	return phoneRegex.MatchString(phone)
}

// UserService 用户服务
type UserService struct {
	userRepo *repository.UserRepo
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepo(database.GetMySQL()),
	}
}

// ListRequest 用户列表请求
type ListRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Name      string `form:"name"`
	ID        string `form:"id"`
	Status    string `form:"status"`
	GroupID   string `form:"groupId"`
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
	Remark    string `form:"remark"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Gender   int    `json:"gender"`
	Birthday string `json:"birthday"`
	Bio      string `json:"bio"`
	Status   int    `json:"status" binding:"required"`
	GroupID  uint   `json:"groupId"`
	Remark   string `json:"remark"`
	Password string `json:"password"`
	RoleIDs  []uint `json:"roleIds"` // 角色ID列表
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Gender   int    `json:"gender"`
	Birthday string `json:"birthday"`
	Bio      string `json:"bio"`
	Status   *int   `json:"status"`
	GroupID  uint   `json:"groupId"`
	Remark   string `json:"remark"`
	RoleIDs  []uint `json:"roleIds"` // 角色ID列表
}

// UserResponse 用户响应（包含角色ID）
type UserResponse struct {
	model.User
	RoleIDs []uint `json:"roleIds"`
}

// List 获取用户列表
func (s *UserService) List(req *ListRequest) ([]UserResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	filters := map[string]interface{}{
		"name":      req.Name,
		"id":        req.ID,
		"status":    req.Status,
		"groupId":   req.GroupID,
		"startTime": req.StartTime,
		"endTime":   req.EndTime,
		"remark":    req.Remark,
	}

	users, total, err := s.userRepo.List(req.Page, req.PageSize, filters)
	if err != nil {
		return nil, 0, err
	}

	// 批量获取用户角色ID（N+1 优化）
	userIDs := make([]uint, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}
	roleIDsMap, _ := s.userRepo.GetUserRoleIDsByUserIDs(userIDs)

	// 转换为响应格式，包含角色ID
	var userResponses []UserResponse
	for _, user := range users {
		roleIDs := roleIDsMap[user.ID]
		if roleIDs == nil {
			roleIDs = []uint{}
		}
		userResponses = append(userResponses, UserResponse{
			User:    user,
			RoleIDs: roleIDs,
		})
	}

	return userResponses, total, nil
}

// GetByID 根据 ID 获取用户
func (s *UserService) GetByID(id uint) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

// Create 创建用户
func (s *UserService) Create(req *CreateUserRequest) error {
	// 校验邮箱和手机号
	if !validateEmail(req.Email) {
		return fmt.Errorf("邮箱格式不正确")
	}
	if !validatePhone(req.Phone) {
		return fmt.Errorf("手机号格式不正确")
	}

	// 加密密码
	password := req.Password
	if password == "" {
		// 生成随机临时密码（8 位十六进制）
		b := make([]byte, 4)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("生成临时密码失败: %w", err)
		}
		password = fmt.Sprintf("Tmp%s!", fmt.Sprintf("%x", b))
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Name:     req.Name,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Gender:   req.Gender,
		Bio:      req.Bio,
		Password: string(hashedPassword),
		Status:   req.Status,
		GroupID:  req.GroupID,
		Remark:   req.Remark,
	}

	// 使用事务确保用户创建和角色分配原子性
	db := database.GetMySQL()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// 同步用户角色
		if len(req.RoleIDs) > 0 {
			for _, roleID := range req.RoleIDs {
				userRole := &model.UserRole{
					UserID: user.ID,
					RoleID: roleID,
				}
				if err := tx.Create(userRole).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// Update 更新用户
func (s *UserService) Update(id uint, req *UpdateUserRequest) error {
	// 校验邮箱和手机号
	if !validateEmail(req.Email) {
		return fmt.Errorf("邮箱格式不正确")
	}
	if !validatePhone(req.Phone) {
		return fmt.Errorf("手机号格式不正确")
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Gender != 0 {
		user.Gender = req.Gender
	}
	if req.Birthday != "" {
		if t, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			user.Birthday = &t
		}
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	// Status 仅在请求中明确提供时更新（避免零值误覆盖）
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.GroupID != 0 {
		user.GroupID = req.GroupID
	}
	if req.Remark != "" {
		user.Remark = req.Remark
	}

	// 使用事务确保用户更新和角色同步原子性
	db := database.GetMySQL()
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(user).Error; err != nil {
			return err
		}

		// 同步用户角色
		if req.RoleIDs != nil {
			// 先删除旧的关联
			if err := tx.Where("user_id = ?", user.ID).Delete(&model.UserRole{}).Error; err != nil {
				return err
			}
			// 创建新的关联
			for _, roleID := range req.RoleIDs {
				userRole := &model.UserRole{
					UserID: user.ID,
					RoleID: roleID,
				}
				if err := tx.Create(userRole).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// 角色或分组变更后，清除权限缓存
	s.invalidatePermissionCache(user.ID)
	return nil
}

// Delete 删除用户（级联清理关联数据）
func (s *UserService) Delete(id uint) error {
	if err := s.userRepo.DeleteWithCleanup(id); err != nil {
		return err
	}
	s.invalidatePermissionCache(id)

	// 清理 Redis 中的 refresh token 和设备黑名单
	ctx := context.Background()
	rdb := database.GetRedis()
	rdb.Del(ctx, fmt.Sprintf("refresh_token:%d", id))

	return nil
}

// invalidatePermissionCache 清除用户权限码缓存
func (s *UserService) invalidatePermissionCache(userID uint) {
	ctx := context.Background()
	_ = cache.Delete(ctx, fmt.Sprintf("permission_codes:%d", userID))
}
