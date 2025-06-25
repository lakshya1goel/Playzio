package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/database"
	"github.com/lakshya1goel/Playzio/domain/model"
)

type UserRepository interface {
	GetUserByID(c *gin.Context, id uint) (model.User, error)
	GetUserByEmail(c *gin.Context, email string) (model.User, error)
	CreateUser(c *gin.Context, user *model.User) error
	UpdateUser(c *gin.Context, user *model.User) error
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

func (r *userRepository) GetUserByEmail(c *gin.Context, email string) (model.User, error) {
	var user model.User
	if err := database.Db.Where("email = ?", email).First(&user).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (r *userRepository) CreateUser(c *gin.Context, user *model.User) error {
	if err := database.Db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpdateUser(c *gin.Context, user *model.User) error {
	if err := database.Db.Save(user).Error; err != nil {
		return err
	}
	return nil
}
