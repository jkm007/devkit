package jwt

import (
	"errors"
	"time"

	"backend-server/config"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token 已过期")
	ErrTokenInvalid = errors.New("token 无效")
)

// Claims JWT 声明
type Claims struct {
	UserID   uint     `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// TokenPair Token 对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Generate 生成 Token 对
func Generate(userID uint, username string, roles []string) (*TokenPair, error) {
	cfg := config.Get().JWT

	now := time.Now()

	// Access Token
	accessClaims := &Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    cfg.Issuer,
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := accessToken.SignedString([]byte(cfg.Secret))
	if err != nil {
		return nil, err
	}

	// Refresh Token
	refreshClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    cfg.Issuer,
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString([]byte(cfg.Secret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
	}, nil
}

// Parse 解析 Token
func Parse(tokenStr string) (*Claims, error) {
	cfg := config.Get().JWT

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 校验签名算法，防止 alg:none 攻击
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// PreviewClaims 预签名 Token 声明
type PreviewClaims struct {
	UserID uint `json:"user_id"`
	FileID uint `json:"file_id"`
	jwt.RegisteredClaims
}

// GeneratePreviewToken 生成预签名 Token（用于视频预览，有效期 1 小时）
func GeneratePreviewToken(userID uint, fileID uint) (string, error) {
	cfg := config.Get().JWT

	claims := &PreviewClaims{
		UserID: userID,
		FileID: fileID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ValidatePreviewToken 验证预签名 Token
func ValidatePreviewToken(tokenStr string) (userID uint, fileID uint, err error) {
	cfg := config.Get().JWT

	token, err := jwt.ParseWithClaims(tokenStr, &PreviewClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, 0, ErrTokenExpired
		}
		return 0, 0, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*PreviewClaims)
	if !ok || !token.Valid {
		return 0, 0, ErrTokenInvalid
	}

	return claims.UserID, claims.FileID, nil
}
