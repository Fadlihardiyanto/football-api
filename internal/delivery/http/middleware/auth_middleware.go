package middleware

import (
	"net/http"
	"strings"

	"github.com/Fadlihardiyanto/football-api/internal/common"
	"github.com/Fadlihardiyanto/football-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authUC usecase.AuthUseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.JSON(http.StatusUnauthorized, common.NewStandardErrorResponse(
				common.ErrUnauthorized("Missing or invalid Authorization header"),
			))
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		auth, err := authUC.ValidateToken(tokenString)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, common.NewStandardErrorResponse(
				common.ErrUnauthorized("Invalid token"),
			))
			ctx.Abort()
			return
		}

		// Simpan info user ke context agar bisa diakses di handler
		ctx.Set("auth", auth) // *model.Auth
		ctx.Next()
	}
}
