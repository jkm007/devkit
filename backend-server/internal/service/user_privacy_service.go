package service

import (
	"errors"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

// UserPrivacyService 用户隐私设置服务
type UserPrivacyService struct {
	repo *repository.UserPrivacyRepo
}

// NewUserPrivacyService 创建用户隐私设置服务
func NewUserPrivacyService() *UserPrivacyService {
	return &UserPrivacyService{
		repo: repository.NewUserPrivacyRepo(database.GetMySQL()),
	}
}

// PrivacyRequest 隐私设置请求
type PrivacyRequest struct {
	ProfileVisible  *int `json:"profileVisible"`
	RealnameVisible *int `json:"realnameVisible"`
	EmailVisible    *int `json:"emailVisible"`
	StatsVisible    *int `json:"statsVisible"`
	ClassVisible    *int `json:"classVisible"`
}

// Get 获取用户的隐私设置
func (s *UserPrivacyService) Get(userID uint) (*model.UserPrivacy, error) {
	privacy, err := s.repo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &model.UserPrivacy{
				UserID:         userID,
				ProfileVisible: 1,
				RealnameVisible: 1,
				EmailVisible:   1,
				StatsVisible:   1,
				ClassVisible:   1,
			}, nil
		}
		return nil, err
	}
	return privacy, nil
}

// Update 更新用户的隐私设置
func (s *UserPrivacyService) Update(userID uint, req *PrivacyRequest) error {
	privacy, err := s.repo.GetByUserID(userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		// 不存在则创建
		privacy = &model.UserPrivacy{
			UserID:         userID,
			ProfileVisible: 1,
			RealnameVisible: 1,
			EmailVisible:   1,
			StatsVisible:   1,
			ClassVisible:   1,
		}
	}

	if req.ProfileVisible != nil {
		privacy.ProfileVisible = *req.ProfileVisible
	}
	if req.RealnameVisible != nil {
		privacy.RealnameVisible = *req.RealnameVisible
	}
	if req.EmailVisible != nil {
		privacy.EmailVisible = *req.EmailVisible
	}
	if req.StatsVisible != nil {
		privacy.StatsVisible = *req.StatsVisible
	}
	if req.ClassVisible != nil {
		privacy.ClassVisible = *req.ClassVisible
	}

	if privacy.ID == 0 {
		return s.repo.Create(privacy)
	}
	return s.repo.Update(privacy)
}
