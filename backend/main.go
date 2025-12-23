package main

import (
	"net/url"
	"teralux_app/controllers"
	"teralux_app/middlewares"
	"teralux_app/routes"
	"teralux_app/services"
	"teralux_app/usecases"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"

	"teralux_app/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Teralux API
// @version         1.0
// @description     This is the API server for Teralux App.
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	utils.LoadConfig()

	if swaggerURL := utils.AppConfig.SwaggerBaseURL; swaggerURL != "" {
		parsedURL, err := url.Parse(swaggerURL)
		if err != nil {
			utils.LogInfo("Warning: Invalid SWAGGER_BASE_URL: %v", err)
		} else {
			docs.SwaggerInfo.Host = parsedURL.Host
			docs.SwaggerInfo.Schemes = []string{parsedURL.Scheme}
		}
	}

	router := gin.Default()

	router.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "" || c.Param("any") == "/" || c.Param("any") == "/index.html" {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(200, docs.CustomSwaggerHTML)
		} else {
			ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
		}
	})

	badgerService, err := services.NewBadgerService("./tmp/badger")
	if err != nil {
		utils.LogInfo("Warning: Failed to initialize BadgerDB: %v", err)
	} else {
		defer badgerService.Close()
	}

	tuyaAuthService := services.NewTuyaAuthService()
	tuyaAuthUseCase := usecases.NewTuyaAuthUseCase(tuyaAuthService)

	tuyaDeviceService := services.NewTuyaDeviceService()

	// Initialize Device State UseCase (needed by other use cases)
	deviceStateUseCase := usecases.NewDeviceStateUseCase(badgerService)

	tuyaGetAllDevicesUseCase := usecases.NewTuyaGetAllDevicesUseCase(tuyaDeviceService, badgerService, deviceStateUseCase)
	tuyaGetDeviceByIDUseCase := usecases.NewTuyaGetDeviceByIDUseCase(tuyaDeviceService, badgerService)
	tuyaDeviceControlUseCase := usecases.NewTuyaDeviceControlUseCase(tuyaDeviceService, deviceStateUseCase)
	tuyaSensorUseCase := usecases.NewTuyaSensorUseCase(tuyaGetDeviceByIDUseCase)

	tuyaAuthController := controllers.NewTuyaAuthController(tuyaAuthUseCase)
	tuyaGetAllDevicesController := controllers.NewTuyaGetAllDevicesController(tuyaGetAllDevicesUseCase)
	tuyaGetDeviceByIDController := controllers.NewTuyaGetDeviceByIDController(tuyaGetDeviceByIDUseCase)
	tuyaDeviceControlController := controllers.NewTuyaDeviceControlController(tuyaDeviceControlUseCase)
	tuyaSensorController := controllers.NewTuyaSensorController(tuyaSensorUseCase)
	cacheController := controllers.NewCacheController(badgerService)
	deviceStateController := controllers.NewDeviceStateController(deviceStateUseCase)

	authGroup := router.Group("/")
	authGroup.Use(middlewares.ApiKeyMiddleware())
	routes.SetupTuyaAuthRoutes(authGroup, tuyaAuthController)

	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	protected.Use(middlewares.TuyaErrorMiddleware())
	{
		routes.SetupTuyaDeviceRoutes(protected, tuyaGetAllDevicesController, tuyaGetDeviceByIDController, tuyaSensorController)
		routes.SetupTuyaControlRoutes(protected, tuyaDeviceControlController)
		routes.SetupCacheRoutes(protected, cacheController)
		routes.SetupDeviceStateRoutes(protected, deviceStateController)
	}
	
	utils.LogInfo("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		utils.LogInfo("Failed to start server: %v", err)
	}
}
