package http

import (
	"net/http"

	"github.com/Fadlihardiyanto/football-api/internal/common"
	"github.com/Fadlihardiyanto/football-api/internal/model"
	"github.com/Fadlihardiyanto/football-api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthController struct {
	UseCase usecase.AuthUseCase
	Log     *logrus.Logger
}

func NewAuthController(useCase usecase.AuthUseCase, log *logrus.Logger) *AuthController {
	return &AuthController{
		UseCase: useCase,
		Log:     log,
	}
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req model.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Invalid JSON format"),
		))
		return
	}

	res, err := c.UseCase.Login(ctx, &req)
	if err != nil {
		c.Log.Errorf("Login failed: %v", err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	// Set HTTPOnly cookie for refresh_token_id
	expiry := c.UseCase.GetRefreshTokenExpiry()
	ctx.SetCookie(
		"refresh_token_id",
		res.RefreshTokenID,
		int(expiry.Seconds()),
		"/", "", true, true,
	)

	response := model.LoginResponseToFrontend{
		User:         res.User,
		AccessToken:  res.AccessToken,
		AccessExpiry: res.AccessExpiry,
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(response, "Login successful"))
}

func (c *AuthController) Register(ctx *gin.Context) {
	var req model.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.NewStandardErrorResponse(
			common.ErrInvalidInput("Invalid JSON format"),
		))
		return
	}

	res, err := c.UseCase.Register(ctx, &req)
	if err != nil {
		c.Log.Errorf("Registration failed: %v", err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusCreated, model.NewSuccessResponse(res, "Registration successful"))
}

func (c *AuthController) Refresh(ctx *gin.Context) {
	refreshTokenID, err := ctx.Cookie("refresh_token_id")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, common.NewStandardErrorResponse(
			common.ErrUnauthorized("Missing refresh token"),
		))
		return
	}

	req := &model.RefreshTokenRequest{
		ID: refreshTokenID,
	}

	res, err := c.UseCase.Refresh(ctx, req)
	if err != nil {
		c.Log.Errorf("Token refresh failed: %v", err)
		if appErr, ok := common.IsAppError(err); ok {
			ctx.JSON(appErr.HTTPCode, common.NewStandardErrorResponse(appErr))
		} else {
			ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
				common.ErrInternalServer("Unknown error"),
			))
		}
		return
	}

	ctx.JSON(http.StatusOK, model.NewSuccessResponse(res, "Token refreshed successfully"))
}

func (c *AuthController) Logout(ctx *gin.Context) {
	refreshTokenID, err := ctx.Cookie("refresh_token_id")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, common.NewStandardErrorResponse(
			common.ErrUnauthorized("Missing refresh token"),
		))
		return
	}
	req := &model.LogoutRequest{
		RefreshTokenID: refreshTokenID,
	}

	if err := c.UseCase.Logout(ctx, req); err != nil {
		c.Log.Errorf("Logout failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, common.NewStandardErrorResponse(
			common.ErrInternalServer("Logout failed"),
		))
		return
	}

	// Clear the cookie
	ctx.SetCookie("refresh_token_id", "", -1, "/", "", true, true)

	ctx.JSON(http.StatusOK, model.NewSuccessResponse[any](nil, "Logged out successfully"))
}
