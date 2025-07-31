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

func (ctrl *AuthController) GoogleSignIn(c *gin.Context) {
	ctrl.authUseCase.HandleGoogleConfig(c)
}

func (ctrl *AuthController) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Message: "Code is required",
		})
		return
	}

	response, err := ctrl.authUseCase.HandleGoogleLogin(c, code)

	if err != nil {
		c.JSON(err.StatusCode, domain.ErrorResponse{
			Message: err.Message,
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "Google login successful",
		Data:    response,
	})
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

func (ctrl *AuthController) GetAccessTokenFromRefreshToekn(c *gin.Context) {
	refreshToken := c.Query("refresh_token")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Message: "Refresh token is required",
		})
		return
	}

	response, httpErr := ctrl.authUseCase.GetAccessTokenFromRefreshToken(c, refreshToken)
	if httpErr != nil {
		c.JSON(httpErr.StatusCode, domain.ErrorResponse{
			Message: httpErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "Access token retrieved successfully",
		Data:    response,
	})
}
