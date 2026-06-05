package service

import (
	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHistoryService 密码历史服务
type PasswordHistoryService struct {
	repo *repository.PasswordHistoryRepo
}

// NewPasswordHistoryService 创建密码历史服务
func NewPasswordHistoryService() *PasswordHistoryService {
	return &PasswordHistoryService{
		repo: repository.NewPasswordHistoryRepo(database.GetMySQL()),
	}
}

// CheckPasswordRequest 检查密码请求
type CheckPasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required"`
}

// CheckRepeated 检查新密码是否与最近 N 次密码重复
func (s *PasswordHistoryService) CheckRepeated(userID uint, newPassword string) (bool, error) {
	histories, err := s.repo.GetRecent(userID, 5)
	if err != nil {
		return false, err
	}

	for _, h := range histories {
		if err := bcrypt.CompareHashAndPassword([]byte(h.Password), []byte(newPassword)); err == nil {
			return true, nil
		}
	}

	return false, nil
}

// SavePassword 保存旧密码到历史
func (s *PasswordHistoryService) SavePassword(userID uint, hashedPassword string) error {
	history := &model.PasswordHistory{
		UserID:   userID,
		Password: hashedPassword,
	}
	return s.repo.Create(history)
}
