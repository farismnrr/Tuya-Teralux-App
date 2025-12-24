package main

import (
	"net/url"
	common_controllers "teralux_app/domain/common/controllers"
	tuya_controllers "teralux_app/domain/tuya/controllers"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/middlewares"
	common_routes "teralux_app/domain/common/routes"
	tuya_routes "teralux_app/domain/tuya/routes"
	"teralux_app/domain/common/infrastructure/persistence"
	"teralux_app/domain/tuya/services"
	"teralux_app/domain/tuya/usecases"
	"teralux_app/domain/common/utils"

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

// @tag.name 01. Auth
// @tag.description Authentication endpoints

// @tag.name 02. Devices
// @tag.description Device management endpoints

// @tag.name 03. Device Control
// @tag.description Device control endpoints

// @tag.name 04. Device Sensor
// @tag.description Sensor data endpoints

// @tag.name 05. Flush
// @tag.description Cache management endpoints

// @tag.name 06. Health
// @tag.description Health check endpoints
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

	// Initialize database connection
	_, err := infrastructure.InitDB()
	if err != nil {
		utils.LogInfo("Warning: Failed to initialize database: %v", err)
	} else {
		defer infrastructure.CloseDB()
		utils.LogInfo("Database initialized successfully")
	}

	router := gin.Default()

	// Health check endpoint
	healthController := common_controllers.NewHealthController()
	router.GET("/health", healthController.CheckHealth)

	router.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "" || c.Param("any") == "/" || c.Param("any") == "/index.html" {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(200, docs.CustomSwaggerHTML)
		} else {
			ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
		}
	})

	badgerService, err := persistence.NewBadgerService("./tmp/badger")
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
	tuyaGetDeviceByIDUseCase := usecases.NewTuyaGetDeviceByIDUseCase(tuyaDeviceService, badgerService, deviceStateUseCase)
	tuyaDeviceControlUseCase := usecases.NewTuyaDeviceControlUseCase(tuyaDeviceService, deviceStateUseCase, badgerService)
	tuyaSensorUseCase := usecases.NewTuyaSensorUseCase(tuyaGetDeviceByIDUseCase)

	tuyaAuthController := tuya_controllers.NewTuyaAuthController(tuyaAuthUseCase)
	tuyaGetAllDevicesController := tuya_controllers.NewTuyaGetAllDevicesController(tuyaGetAllDevicesUseCase)
	tuyaGetDeviceByIDController := tuya_controllers.NewTuyaGetDeviceByIDController(tuyaGetDeviceByIDUseCase)
	tuyaDeviceControlController := tuya_controllers.NewTuyaDeviceControlController(tuyaDeviceControlUseCase)
	tuyaSensorController := tuya_controllers.NewTuyaSensorController(tuyaSensorUseCase)
	cacheController := common_controllers.NewCacheController(badgerService)

	authGroup := router.Group("/")
	authGroup.Use(middlewares.ApiKeyMiddleware())
	tuya_routes.SetupTuyaAuthRoutes(authGroup, tuyaAuthController)

	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	protected.Use(middlewares.TuyaErrorMiddleware())
	{
		tuya_routes.SetupTuyaDeviceRoutes(protected, tuyaGetAllDevicesController, tuyaGetDeviceByIDController, tuyaSensorController)
		tuya_routes.SetupTuyaControlRoutes(protected, tuyaDeviceControlController)
		common_routes.SetupCacheRoutes(protected, cacheController)
	}
	
	utils.LogInfo("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		utils.LogInfo("Failed to start server: %v", err)
	}
}