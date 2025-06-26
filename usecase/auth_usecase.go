package usecase

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain"
	"github.com/lakshya1goel/Playzio/domain/dto"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/repository"
)

type AuthUseCase interface {
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
	exp := time.Now().Add(24 * time.Hour).Unix()
	token, err := util.GenerateGuestToken(name, exp)
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
