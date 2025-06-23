package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/usecase"
	"github.com/markbates/goth/gothic"
)

type AuthController struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthController(authUC usecase.AuthUseCase) *AuthController {
	return &AuthController{
		authUseCase: authUC,
	}
}

func (ctrl *AuthController) BeginAuth(c *gin.Context) {
	c.Request = c.Request.WithContext(c)

	provider := c.Query("provider")
	if provider == "" {
		provider = "google"
	}

	q := c.Request.URL.Query()
	q.Set("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func (ctrl *AuthController) Callback(c *gin.Context) {
	c.Request = c.Request.WithContext(c)

	provider := c.Query("provider")
	if provider == "" {
		provider = "google"
	}
	q := c.Request.URL.Query()
	q.Set("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	appUser := &model.User{
		Name:       gothUser.Name,
		Email:      gothUser.Email,
		ProfilePic: &gothUser.AvatarURL,
	}

	finalUser, err := ctrl.authUseCase.Authenticate(c, appUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth failed"})
		return
	}

	c.JSON(http.StatusOK, finalUser)
}
