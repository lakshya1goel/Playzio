package usecase

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain"
	"github.com/lakshya1goel/Playzio/domain/dto"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/repository"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type AuthUseCase interface {
	HandleGoogleConfig(c *gin.Context)
	HandleGoogleLogin(c *gin.Context, code string) (*dto.AuthResponse, *domain.HttpError)
	Authenticate(c *gin.Context, user model.User) (*dto.AuthResponse, *domain.HttpError)
	AuthenticateGuest(c *gin.Context, name string) (*dto.GuestAuthResponse, *domain.HttpError)
	GetAccessTokenFromRefreshToken(c *gin.Context, refreshToken string) (*dto.AccessTokenResponse, *domain.HttpError)
}

type authUseCase struct {
	userRepo repository.UserRepository
}

func NewAuthUseCase() AuthUseCase {
	return &authUseCase{
		userRepo: repository.NewUserRepository(),
	}
}

func (uc *authUseCase) HandleGoogleConfig(c *gin.Context) {
	url := util.GoogleOAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
	fmt.Println("Redirecting to Google OAuth URL:", url)
}

func (uc *authUseCase) HandleGoogleLogin(c *gin.Context, code string) (*dto.AuthResponse, *domain.HttpError) {
	token, err := util.GoogleOAuthConfig.Exchange(c, code)
	if err != nil {
		return nil, domain.NewHttpError(http.StatusInternalServerError, "Failed to exchange token with Google")
	}

	client := util.GoogleOAuthConfig.Client(c, token)
	userInfoResp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo?alt=json")
	if err != nil {
		return nil, domain.NewHttpError(http.StatusInternalServerError, "Failed to get user info from Google")
	}
	defer userInfoResp.Body.Close()

	var userInfo dto.UserInfo
	body, _ := io.ReadAll(userInfoResp.Body)
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return nil, domain.NewHttpError(http.StatusInternalServerError, "Failed to unmarshal user info")
	}

	resp, err := uc.userRepo.GetUserByEmail(c, userInfo.Email)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, &domain.HttpError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to get user by email",
			}
		} else {
			accessTokenExp := time.Now().Add(24 * time.Hour).Unix()
			refreshTokenExp := time.Now().Add(24 * 30 * time.Hour).Unix()

			accessToken, err := util.GenerateToken(resp.ID, userInfo.Name, accessTokenExp)
			if err != nil {
				return nil, &domain.HttpError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Failed to generate access token",
				}
			}
			refreshToken, err := util.GenerateToken(resp.ID, userInfo.Name, refreshTokenExp)
			if err != nil {
				return nil, &domain.HttpError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Failed to generate refresh token",
				}
			}

			user := &model.User{
				Name:       userInfo.Name,
				Email:      userInfo.Email,
				ProfilePic: &userInfo.Picture,
			}

			err = uc.userRepo.CreateUser(c, user)
			if err != nil {
				return nil, &domain.HttpError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Failed to create user",
				}
			}
			return &dto.AuthResponse{
				ID:              user.ID,
				Name:            user.Name,
				Email:           user.Email,
				ProfilePic:      *user.ProfilePic,
				AccessToken:     accessToken,
				AccessTokenExp:  accessTokenExp,
				RefreshToken:    refreshToken,
				RefreshTokenExp: refreshTokenExp,
			}, nil
		}
	}

	accessTokenExp := time.Now().Add(24 * time.Hour).Unix()
	refreshTokenExp := time.Now().Add(24 * 30 * time.Hour).Unix()

	accessToken, err := util.GenerateToken(resp.ID, userInfo.Name, accessTokenExp)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to generate access token",
		}
	}
	refreshToken, err := util.GenerateToken(resp.ID, userInfo.Name, refreshTokenExp)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to generate refresh token",
		}
	}

	return &dto.AuthResponse{
		ID:              resp.ID,
		Name:            resp.Name,
		Email:           resp.Email,
		ProfilePic:      *resp.ProfilePic,
		AccessToken:     accessToken,
		AccessTokenExp:  accessTokenExp,
		RefreshToken:    refreshToken,
		RefreshTokenExp: refreshTokenExp,
	}, nil
}

func (uc *authUseCase) Authenticate(c *gin.Context, user model.User) (*dto.AuthResponse, *domain.HttpError) {
	resp, err := uc.userRepo.GetUserByEmail(c, user.Email)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to get user by email",
		}
	}

	accessTokenExp := time.Now().Add(24 * time.Hour).Unix()
	refreshTokenExp := time.Now().Add(24 * 30 * time.Hour).Unix()

	accessToken, err := util.GenerateToken(resp.ID, user.Name, accessTokenExp)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to generate access token",
		}
	}
	refreshToken, err := util.GenerateToken(resp.ID, user.Name, refreshTokenExp)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to generate refresh token",
		}
	}

	err = uc.userRepo.CreateUser(c, &user)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to create user",
		}
	}
	return &dto.AuthResponse{
		ID:              resp.ID,
		Name:            resp.Name,
		Email:           resp.Email,
		ProfilePic:      *resp.ProfilePic,
		AccessToken:     accessToken,
		AccessTokenExp:  accessTokenExp,
		RefreshToken:    refreshToken,
		RefreshTokenExp: refreshTokenExp,
	}, nil
}

func (uc *authUseCase) AuthenticateGuest(c *gin.Context, name string) (*dto.GuestAuthResponse, *domain.HttpError) {
	guestID := uuid.NewString()
	exp := time.Now().Add(24 * time.Hour).Unix()
	token, err := util.GenerateGuestToken(name, guestID, exp)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to generate guest token",
		}
	}
	response := &dto.GuestAuthResponse{
		Token: token,
		Exp:   exp,
		Name:  name,
	}
	return response, nil
}

func (uc *authUseCase) GetAccessTokenFromRefreshToken(c *gin.Context, refreshToken string) (*dto.AccessTokenResponse, *domain.HttpError) {
	userID, err := util.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid refresh token",
		}
	}

	resp, err := uc.userRepo.GetUserByID(c, userID)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to get user by ID",
		}
	}

	accessTokenExp := time.Now().Add(24 * time.Hour).Unix()

	accessToken, err := util.GenerateToken(resp.ID, resp.Name, accessTokenExp)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to generate access token",
		}
	}

	return &dto.AccessTokenResponse{
		AccessToken:    accessToken,
		AccessTokenExp: accessTokenExp,
	}, nil
}
