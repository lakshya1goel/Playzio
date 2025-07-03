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
)

type AuthUseCase interface {
	HandleGoogleConfig(c *gin.Context)
	HandleGoogleLogin(c *gin.Context, code string) (*dto.AuthResponse, *domain.HttpError)
	Authenticate(c *gin.Context, user model.User) (*dto.AuthResponse, *domain.HttpError)
	AuthenticateGuest(c *gin.Context, name string) (*dto.GuestAuthResponse, *domain.HttpError)
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
	json.Unmarshal(body, &userInfo)

	resp, err := uc.userRepo.GetUserByEmail(c, userInfo.Email)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to get user by email",
		}
	}

	accessTokenExp := time.Now().Add(24 * time.Hour).Unix()
	refreshTokenExp := time.Now().Add(24 * 30 * time.Hour).Unix()

	accessToken, err := util.GenerateToken(resp.ID, accessTokenExp)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to generate access token",
		}
	}
	refreshToken, err := util.GenerateToken(resp.ID, refreshTokenExp)
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

	accessToken, err := util.GenerateToken(resp.ID, accessTokenExp)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to generate access token",
		}
	}
	refreshToken, err := util.GenerateToken(resp.ID, refreshTokenExp)
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
