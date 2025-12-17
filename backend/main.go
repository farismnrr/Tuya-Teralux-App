package main

import (
	"log"
	"teralux_app/controllers"
	"teralux_app/middlewares"
	"teralux_app/routes"
	"teralux_app/services"
	"teralux_app/usecases"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	utils.LoadConfig()

	// Initialize Gin router
	router := gin.Default()

	// Initialize dependency chain: service -> usecase -> controller
	tuyaAuthService := services.NewTuyaAuthService()
	tuyaAuthUseCase := usecases.NewTuyaAuthUseCase(tuyaAuthService)

	tuyaDeviceService := services.NewTuyaDeviceService()

	// Initialize Get All Devices chain
	tuyaGetAllDevicesUseCase := usecases.NewTuyaGetAllDevicesUseCase(tuyaDeviceService)
	tuyaGetDeviceByIDUseCase := usecases.NewTuyaGetDeviceByIDUseCase(tuyaDeviceService)
	tuyaDeviceControlUseCase := usecases.NewTuyaDeviceControlUseCase(tuyaDeviceService)

	tuyaAuthController := controllers.NewTuyaAuthController(tuyaAuthUseCase)
	tuyaGetAllDevicesController := controllers.NewTuyaGetAllDevicesController(tuyaGetAllDevicesUseCase)
	tuyaGetDeviceByIDController := controllers.NewTuyaGetDeviceByIDController(tuyaGetDeviceByIDUseCase)
	tuyaDeviceControlController := controllers.NewTuyaDeviceControlController(tuyaDeviceControlUseCase)
	tuyaSensorController := controllers.NewTuyaSensorController(tuyaGetDeviceByIDUseCase)

	// Public Routes (Protected by API Key)
	authGroup := router.Group("/")
	authGroup.Use(middlewares.ApiKeyMiddleware())
	routes.SetupTuyaAuthRoutes(authGroup, tuyaAuthController)

	// Protected Routes
	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	{
		routes.SetupTuyaDeviceRoutes(protected, tuyaGetAllDevicesController, tuyaGetDeviceByIDController, tuyaSensorController)
		routes.SetupTuyaControlRoutes(protected, tuyaDeviceControlController)
	}
	// Start server
	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
