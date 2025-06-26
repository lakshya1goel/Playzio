package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/domain"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/usecase"
	"github.com/markbates/goth/gothic"
)

type AuthController struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthController() *AuthController {
	return &AuthController{
		authUseCase: usecase.NewAuthUseCase(),
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

	response, httpErr := ctrl.authUseCase.Authenticate(c, *appUser)
	if httpErr != nil {
		c.JSON(httpErr.StatusCode, domain.ErrorResponse{
			Message: httpErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "User authenticated successfully",
		Data:    response,
	})
}

func (ctrl *AuthController) GuestAuth(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Message: "Name is required",
		})
		return
	}

	response, httpErr := ctrl.authUseCase.AuthenticateGuest(c, name)
	if httpErr != nil {
		c.JSON(httpErr.StatusCode, domain.ErrorResponse{
			Message: httpErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Guest authenticated successfully",
		Data:    response,
	})
}
