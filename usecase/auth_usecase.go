package usecase

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/repository"
)

type AuthUseCase interface {
	Authenticate(c *gin.Context, user *model.User) (model.User, error)
}

type authUseCase struct {
	userRepo repository.UserRepository
}

func NewAuthUseCase() AuthUseCase {
	return &authUseCase{
		userRepo: repository.NewUserRepository(),
	}
}

func (uc *authUseCase) Authenticate(c *gin.Context, user *model.User) (model.User, error) {
	found, err := uc.userRepo.GetUserByEmail(c, user.Email)
	if err == nil {
		return found, nil
	}
	err = uc.userRepo.CreateUser(c, user)
	if err != nil {
		return model.User{}, err
	}
	return *user, nil
}
