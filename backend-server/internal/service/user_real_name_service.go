package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/crypto"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

// UserRealNameService 实名认证服务
type UserRealNameService struct {
	repo     *repository.UserRealNameRepo
	userRepo *repository.UserRepo
}

// NewUserRealNameService 创建实名认证服务
func NewUserRealNameService() *UserRealNameService {
	db := database.GetMySQL()
	return &UserRealNameService{
		repo:     repository.NewUserRealNameRepo(db),
		userRepo: repository.NewUserRepo(db),
	}
}

// RealNameSubmitRequest 提交实名认证请求
type RealNameSubmitRequest struct {
	RealName string `json:"realName" binding:"required"`
	IDCard   string `json:"idCard" binding:"required"`
}

// RealNameRejectRequest 拒绝实名认证请求
type RealNameRejectRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// RealNameListRequest 实名认证列表请求
type RealNameListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status"`
	UserID   string `form:"userId"`
	RealName string `form:"realName"`
}

// RealNameStatusResponse 实名认证状态响应
type RealNameStatusResponse struct {
	Status       int        `json:"status"`
	RealName     string     `json:"realName"`
	IDCard       string     `json:"idCard"`
	RejectReason string     `json:"rejectReason"`
	SubmittedAt  *time.Time `json:"submittedAt"`
	ReviewedAt   *time.Time `json:"reviewedAt"`
}

// GetStatus 获取当前用户的实名认证状态
func (s *UserRealNameService) GetStatus(userID uint) (*RealNameStatusResponse, error) {
	rn, err := s.repo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &RealNameStatusResponse{Status: 0}, nil
		}
		return nil, err
	}

	// 解密身份证号后脱敏
	idCardPlain := decryptIDCard(rn.IDCard)
	idCardMasked := maskIDCard(idCardPlain)

	return &RealNameStatusResponse{
		Status:       rn.Status,
		RealName:     rn.RealName,
		IDCard:       idCardMasked,
		RejectReason: rn.RejectReason,
		SubmittedAt:  rn.SubmittedAt,
		ReviewedAt:   rn.ReviewedAt,
	}, nil
}

// Submit 提交实名认证
func (s *UserRealNameService) Submit(userID uint, req *RealNameSubmitRequest) error {
	// 检查是否已认证
	existing, err := s.repo.GetByUserID(userID)
	if err == nil {
		if existing.Status == 1 {
			return errors.New("Real name already verified.")
		}
		if existing.Status == 0 {
			return errors.New("Verification pending review.")
		}
		// 状态为2（认证失败），允许重新提交
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 检查身份证号是否已被使用（通过哈希比对）
	hash := sha256Hash(req.IDCard)
	existingByHash, _ := s.repo.GetByIDCardHash(hash)
	if existingByHash != nil && existingByHash.UserID != userID {
		return errors.New("This ID card is already registered.")
	}

	// AES-GCM 加密身份证号
	encryptedIDCard, err := crypto.Encrypt(req.IDCard)
	if err != nil {
		// 加密失败时降级存储明文（记录日志）
		encryptedIDCard = req.IDCard
	}

	now := time.Now()
	rn := &model.UserRealName{
		UserID:      userID,
		RealName:    req.RealName,
		IDCard:      encryptedIDCard,
		IDCardHash:  hash,
		Status:      0,
		SubmittedAt: &now,
	}

	if existing != nil && existing.ID > 0 {
		// 更新已有记录（重新提交）
		existing.RealName = req.RealName
		existing.IDCard = encryptedIDCard
		existing.IDCardHash = hash
		existing.Status = 0
		existing.RejectReason = ""
		existing.SubmittedAt = &now
		return s.repo.Update(existing)
	}

	return s.repo.Create(rn)
}

// List 获取实名认证列表（管理员）
func (s *UserRealNameService) List(req *RealNameListRequest) ([]model.UserRealName, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	filters := map[string]interface{}{
		"status":   req.Status,
		"userId":   req.UserID,
		"realName": req.RealName,
	}

	return s.repo.List(req.Page, req.PageSize, filters)
}

// Approve 审核通过
func (s *UserRealNameService) Approve(id, reviewerID uint) error {
	rn, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if rn.Status != 0 {
		return errors.New("Application is not pending review.")
	}

	now := time.Now()
	rn.Status = 1
	rn.ReviewedBy = &reviewerID
	rn.ReviewedAt = &now

	if err := s.repo.Update(rn); err != nil {
		return err
	}

	// 更新用户的实名信息
	user, err := s.userRepo.GetByID(rn.UserID)
	if err != nil {
		return err
	}
	user.RealName = rn.RealName
	user.IDCard = rn.IDCard // 已加密存储
	user.IsReal = 1
	return s.userRepo.Update(user)
}

// Reject 审核拒绝
func (s *UserRealNameService) Reject(id, reviewerID uint, reason string) error {
	rn, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if rn.Status != 0 {
		return errors.New("Application is not pending review.")
	}

	now := time.Now()
	rn.Status = 2
	rn.RejectReason = reason
	rn.ReviewedBy = &reviewerID
	rn.ReviewedAt = &now

	return s.repo.Update(rn)
}

// decryptIDCard 解密身份证号（兼容明文数据）
func decryptIDCard(encrypted string) string {
	if encrypted == "" {
		return ""
	}
	decrypted, err := crypto.Decrypt(encrypted)
	if err != nil {
		// 解密失败说明是明文存储的旧数据，直接返回
		return encrypted
	}
	return decrypted
}

// maskIDCard 脱敏身份证号
func maskIDCard(idCard string) string {
	if len(idCard) <= 8 {
		if len(idCard) == 0 {
			return ""
		}
		return "***"
	}
	return idCard[:3] + "***********" + idCard[len(idCard)-4:]
}

// sha256Hash SHA256 哈希
func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
