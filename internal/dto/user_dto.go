package dto

type UserLoginDTO struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}
type UserRegisterDTO struct {
	Username string `json:"username" binding:"required,min=3,max=32"` // 用户名
	Password string `json:"password" binding:"required,min=6,max=64"` // 密码
	//Email    string `json:"email" binding:"required,email"`           // 邮箱
}

type UserInfo struct {
	ID       int    `json:"id"`       // 用户ID
	Username string `json:"username"` // 用户名
}
type LoginResponse struct {
	AccessToken           string   `json:"access_token"`             // 访问令牌
	RefreshToken          string   `json:"refresh_token"`            // 刷新令牌
	AccessTokenExpiresIn  int64    `json:"access_token_expires_in"`  // AccessToken过期时间（秒）
	RefreshTokenExpiresIn int64    `json:"refresh_token_expires_in"` // RefreshToken过期时间（秒）
	TokenType             string   `json:"token_type"`               // 令牌类型，固定为Bearer
	User                  UserInfo `json:"user"`                     // 用户信息
}

type RegisterResponse struct {
	UserInfo
}

// RefreshTokenResponse Token刷新响应
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`  // 新的访问令牌
	RefreshToken string `json:"refresh_token"` // 新的刷新令牌
	ExpiresIn    int64  `json:"expires_in"`    // AccessToken过期时间（秒）
	TokenType    string `json:"token_type"`    // 令牌类型
}
