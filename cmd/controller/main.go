package main

import (
	"time"

	"github.com/leetsecure/qryptic-controller/cmd/controller/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/leetsecure/qryptic-controller/internal/config"
	"github.com/leetsecure/qryptic-controller/internal/database"
	"github.com/leetsecure/qryptic-controller/internal/routes"
	"github.com/leetsecure/qryptic-controller/internal/services"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title			Qryptic Controller API
// @version		1.0
// @description	This is a Qryptic Controller Service to manage the users, groups, gateways and clients.
// @contact.name	Leetsecure Support
// @contact.url	leetsecure.com
// @contact.email	hello@leetsecure.com
// @BasePath		/
func main() {
	log := logger.Default()

	err := config.UpdateEnvConfig()
	if err != nil {
		log.Error(err)
		return
	}

	err = database.ConnectDatabase()
	if err != nil {
		log.Error(err)
		return
	}

	err = database.AutomigrateDatabase()
	if err != nil {
		log.Error(err)
		return
	}

	err = services.InitAdminConfig()
	if err != nil {
		log.Error(err)
		return
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: config.CORSAllowCredentials,
		MaxAge:           1 * time.Hour,
		AllowOrigins:     config.CORSAllowedOrigins,
	}))

	routes.SetupControllerRoutes(router)
	if config.Environment != "production" {
		if config.Environment == "local" {
			docs.SwaggerInfo.Host = "localhost:8080"
			docs.SwaggerInfo.Schemes = []string{"http"}
		} else if config.Environment == "development" {
			docs.SwaggerInfo.Host = "leet.controller.leetsecure.com"
			docs.SwaggerInfo.Schemes = []string{"https"}
		}

		// Setup the routes in the router
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Start the server
	router.Run(":8080")
}
