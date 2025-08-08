package dto

type UserInfo struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type AuthResponse struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	ProfilePic      string `json:"profile_pic"`
	AccessToken     string `json:"access_token"`
	AccessTokenExp  int64  `json:"access_token_exp"`
	RefreshToken    string `json:"refresh_token"`
	RefreshTokenExp int64  `json:"refresh_token_exp"`
}

type GuestAuthResponse struct {
	Token string `json:"token"`
	Exp   int64  `json:"exp"`
	Name  string `json:"name"`
}

type AccessTokenResponse struct {
	AccessToken     string `json:"access_token"`
	AccessTokenExp  int64  `json:"access_token_exp"`
}

type User struct {
	Type      string
	UserID    *uint
	Username  *string
	GuestID   *string
	GuestName *string
}
