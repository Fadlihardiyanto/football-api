package config

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func NewGin(config *viper.Viper) *gin.Engine {
	// Set Gin mode: release/debug/test
	gin.SetMode(config.GetString("GIN_MODE")) // e.g. "release", "debug"

	app := gin.New()
	app.Use(gin.Recovery())
	app.Use(NewErrorHandler())

	err := app.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	if err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	return app
}

func NewErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // execute all handlers

		if len(c.Errors) > 0 {
			// Take the last error (or loop if you want)
			err := c.Errors.Last().Err
			c.JSON(c.Writer.Status(), gin.H{
				"errors": err.Error(),
			})
		}
	}
}
