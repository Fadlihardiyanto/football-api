package route

import (
	"net/http"

	httpdelivery "github.com/Fadlihardiyanto/football-api/internal/delivery/http"
	"github.com/gin-gonic/gin"
)

// RouteConfig holds Gin engine and controllers
type RouteConfig struct {
	App               *gin.Engine
	AuthController    *httpdelivery.AuthController
	TeamsController   *httpdelivery.TeamsController
	PlayerController  *httpdelivery.PlayersController
	MatchesController *httpdelivery.MatchesController
	GoalsController   *httpdelivery.GoalsController
	AuthMiddleware    gin.HandlerFunc
}

func (c *RouteConfig) Setup() {
	api := c.App.Group("/api/v1")

	// public routes
	c.SetupGuestRoutes(api)

	protected := api.Group("/")

	// protected routes
	protected.Use(c.AuthMiddleware)

	c.SetupAuthRoutes(protected)
}

func (c *RouteConfig) SetupGuestRoutes(api *gin.RouterGroup) {
	c.App.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world!")
	})

	c.App.Static("/uploads", "./uploads")

	auth := api.Group("/auth")
	auth.POST("/login", c.AuthController.Login)
	auth.POST("/register", c.AuthController.Register)
	auth.POST("/refresh", c.AuthController.Refresh)
	auth.POST("/logout", c.AuthController.Logout)
}

func (c *RouteConfig) SetupAuthRoutes(api *gin.RouterGroup) {

	teams := api.Group("/teams")
	teams.GET("/", c.TeamsController.FindAll)
	teams.GET("/:id", c.TeamsController.FindByID)
	teams.POST("/", c.TeamsController.Create)
	teams.PUT("/:id", c.TeamsController.Update)
	teams.DELETE("/:id", c.TeamsController.SoftDelete)
	teams.POST("/:id/upload-logo", c.TeamsController.UploadLogo)

	players := api.Group("/players")
	players.GET("/", c.PlayerController.FindAll)
	players.GET("/:id", c.PlayerController.FindByID)
	players.POST("/", c.PlayerController.Create)
	players.PUT("/:id", c.PlayerController.Update)
	players.DELETE("/:id", c.PlayerController.SoftDelete)

	matches := api.Group("/matches")
	matches.GET("/", c.MatchesController.FindAll)
	matches.GET("/:id", c.MatchesController.FindByID)
	matches.POST("/", c.MatchesController.Create)
	matches.PUT("/:id", c.MatchesController.Update)
	matches.DELETE("/:id", c.MatchesController.SoftDelete)
	matches.GET("/:id/report", c.MatchesController.GetMatchReport)
	matches.POST("/:id/finish", c.MatchesController.FinishMatch)

	goals := api.Group("/goals")
	goals.GET("/", c.GoalsController.FindAll)
	goals.GET("/:id", c.GoalsController.FindByID)
	goals.POST("/", c.GoalsController.Create)
	goals.PUT("/:id", c.GoalsController.Update)
	goals.DELETE("/:id", c.GoalsController.SoftDelete)
}
