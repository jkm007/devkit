package service

import (
	"errors"
	"time"

	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

// OAuthService 第三方登录绑定服务
type OAuthService struct {
	repo     *repository.OAuthUserRepo
	userRepo *repository.UserRepo
}

// NewOAuthService 创建第三方登录绑定服务
func NewOAuthService() *OAuthService {
	db := database.GetMySQL()
	return &OAuthService{
		repo:     repository.NewOAuthUserRepo(db),
		userRepo: repository.NewUserRepo(db),
	}
}

// OAuthBindURLRequest 获取授权URL请求
type OAuthBindURLRequest struct {
	Provider   string `form:"provider" binding:"required"`
	RedirectURI string `form:"redirectUri"`
}

// OAuthUnbindRequest 解绑请求
type OAuthUnbindRequest struct {
	Provider string `json:"provider" binding:"required"`
}

// OAuthCallbackRequest 回调请求
type OAuthCallbackRequest struct {
	Provider string `form:"provider" binding:"required"`
	Code     string `form:"code" binding:"required"`
	State    string `form:"state" binding:"required"`
}

// OAuthBindingItem 绑定列表项
type OAuthBindingItem struct {
	ID              uint       `json:"id"`
	Provider        string     `json:"provider"`
	ProviderUsername string    `json:"providerUsername"`
	ProviderAvatar  string     `json:"providerAvatar"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// ListBindings 获取用户的第三方绑定列表
func (s *OAuthService) ListBindings(userID uint) ([]OAuthBindingItem, error) {
	bindings, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	var items []OAuthBindingItem
	for _, b := range bindings {
		items = append(items, OAuthBindingItem{
			ID:              b.ID,
			Provider:        b.Provider,
			ProviderUsername: b.ProviderUsername,
			ProviderAvatar:  b.ProviderAvatar,
			CreatedAt:       b.CreatedAt,
		})
	}
	return items, nil
}

// Unbind 解绑第三方账号
func (s *OAuthService) Unbind(userID uint, provider string) error {
	binding, err := s.repo.GetByUserAndProvider(userID, provider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("This provider is not bound.")
		}
		return err
	}

	// 如果用户没有设置密码，则至少保留一个 OAuth 绑定
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user.Password == "" {
		count, err := s.repo.CountByUser(userID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return errors.New("Cannot unbind the only login method.")
		}
	}

	return s.repo.Delete(binding.ID)
}
