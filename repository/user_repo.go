package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/database"
	"github.com/lakshya1goel/Playzio/domain/model"
)

type UserRepository interface {
	GetUserByID(c *gin.Context, id uint) (model.User, error)
}

type userRepository struct{}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) GetUserByID(c *gin.Context, id uint) (model.User, error) {
	var user model.User
	if err := database.Db.First(&user, id).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}
