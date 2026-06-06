package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 应用总配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LogConfig       `mapstructure:"log"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`
	Storage   StorageConfig   `mapstructure:"storage"`
	CDN       CDNConfig       `mapstructure:"cdn"`
	ClamAV    ClamAVConfig    `mapstructure:"clamav"`
	Crypto    CryptoConfig    `mapstructure:"crypto"`
	Captcha   CaptchaConfig   `mapstructure:"captcha"`
	SMS       SMSConfig       `mapstructure:"sms"`
	Email     EmailConfig     `mapstructure:"email"`
	WeChat    WeChatConfig    `mapstructure:"wechat"`
	OAuth     OAuthConfig     `mapstructure:"oauth"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	User           string        `mapstructure:"user"`
	Password       string        `mapstructure:"password"`
	Database       string        `mapstructure:"database"`
	MaxOpenConns   int           `mapstructure:"max_open_conns"`
	MaxIdleConns   int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	Charset        string        `mapstructure:"charset"`
}

// DSN 返回 MySQL 连接字符串
func (c *MySQLConfig) DSN() string {
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" +
		intToStr(c.Port) + ")/" + c.Database + "?charset=" + c.Charset +
		"&parseTime=True&loc=Local"
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Addr 返回 Redis 地址
func (c *RedisConfig) Addr() string {
	return c.Host + ":" + intToStr(c.Port)
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
	Issuer          string        `mapstructure:"issuer"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowOrigins     []string      `mapstructure:"allow_origins"`
	AllowMethods     []string      `mapstructure:"allow_methods"`
	AllowHeaders     []string      `mapstructure:"allow_headers"`
	ExposeHeaders    []string      `mapstructure:"expose_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Rate  float64 `mapstructure:"rate"`
	Burst int     `mapstructure:"burst"`
}

// StorageConfig 文件存储配置
type StorageConfig struct {
	Driver string          `mapstructure:"driver"`
	Local  LocalConfig     `mapstructure:"local"`
	MinIO  MinIOConfig     `mapstructure:"minio"`
	OSS    OSSConfig       `mapstructure:"oss"`
	COS    COSConfig       `mapstructure:"cos"`
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	Path      string `mapstructure:"path"`
	URLPrefix string `mapstructure:"url_prefix"`
}

// MinIOConfig MinIO 配置
type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

// OSSConfig 阿里云 OSS 配置
type OSSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Bucket          string `mapstructure:"bucket"`
	CDNDomain       string `mapstructure:"cdn_domain"`
}

// COSConfig 腾讯云 COS 配置
type COSConfig struct {
	Region    string `mapstructure:"region"`
	SecretID  string `mapstructure:"secret_id"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	CDNDomain string `mapstructure:"cdn_domain"`
}

// CDNConfig CDN 配置
type CDNConfig struct {
	Domain             string `mapstructure:"domain"`
	ImageStyleSeparator string `mapstructure:"image_style_separator"`
}

// ClamAVConfig ClamAV 病毒扫描配置
type ClamAVConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	Socket  string        `mapstructure:"socket"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// CryptoConfig 加密配置
type CryptoConfig struct {
	AESKey string `mapstructure:"aes_key"`
}

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	Width    int     `mapstructure:"width"`
	Height   int     `mapstructure:"height"`
	Length   int     `mapstructure:"length"`
	MaxSkew  float64 `mapstructure:"max_skew"`
	DotCount int     `mapstructure:"dot_count"`
	Secret   string  `mapstructure:"secret"`
}

// SMSConfig 短信服务配置
type SMSConfig struct {
	Driver  string       `mapstructure:"driver"`
	Aliyun  AliyunSMS    `mapstructure:"aliyun"`
	Tencent TencentSMS   `mapstructure:"tencent"`
}

// AliyunSMS 阿里云短信配置
type AliyunSMS struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	SignName        string `mapstructure:"sign_name"`
	TemplateCode    string `mapstructure:"template_code"`
}

// TencentSMS 腾讯云短信配置
type TencentSMS struct {
	SecretID   string `mapstructure:"secret_id"`
	SecretKey  string `mapstructure:"secret_key"`
	AppID      string `mapstructure:"app_id"`
	SignName   string `mapstructure:"sign_name"`
	TemplateID string `mapstructure:"template_id"`
}

// EmailConfig 邮件服务配置
type EmailConfig struct {
	Driver string    `mapstructure:"driver"`
	SMTP   SMTPConfig `mapstructure:"smtp"`
}

// SMTPConfig SMTP 配置
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

// WeChatConfig 微信配置
type WeChatConfig struct {
	MiniProgram      WeChatMiniProgram `mapstructure:"mini_program"`
	OfficialAccount  WeChatOfficial    `mapstructure:"official_account"`
	OAuth            WeChatOAuth       `mapstructure:"oauth"`
}

// WeChatMiniProgram 小程序配置
type WeChatMiniProgram struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
}

// WeChatOfficial 公众号配置
type WeChatOfficial struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
}

// WeChatOAuth 微信 OAuth 配置
type WeChatOAuth struct {
	RedirectURL string `mapstructure:"redirect_url"`
}

// OAuthConfig 第三方 OAuth 配置
type OAuthConfig struct {
	GitHub  OAuthProviderConfig `mapstructure:"github"`
	Google  OAuthProviderConfig `mapstructure:"google"`
	WeChat  OAuthProviderConfig `mapstructure:"wechat"`
}

// OAuthProviderConfig 单个 OAuth 提供商配置
type OAuthProviderConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)

	// 支持环境变量覆盖
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	globalConfig = cfg
	return cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

func intToStr(i int) string {
	return fmt.Sprintf("%d", i)
}
