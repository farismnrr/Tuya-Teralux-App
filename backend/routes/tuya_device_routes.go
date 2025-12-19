package routes

import (
	"teralux_app/controllers"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"
)

// SetupTuyaDeviceRoutes registers Tuya device routes
func SetupTuyaDeviceRoutes(
	router gin.IRouter,
	getAllDevicesController *controllers.TuyaGetAllDevicesController,
	getDeviceByIDController *controllers.TuyaGetDeviceByIDController,
	sensorController *controllers.TuyaSensorController,
) {
	utils.LogDebug("SetupTuyaDeviceRoutes initialized")
	// Group: /api/tuya
	api := router.Group("/api/tuya")
	{
		// Get all devices
		// URL: /api/tuya/devices
		// Method: GET
		// Headers:
		//    - Authorization: Bearer <token>
		// Response: {
		//    "status": true,
		//    "message": "Devices fetched successfully",
		//    "data": {
		//      "devices": [
		//        {
		//          "id": "device_id",
		//          "name": "Device Name",
		//          "category": "category_code",
		//          "product_name": "Product Name",
		//          "online": true,
		//          "icon": "http://img.url",
		//          "status": [
		//            { "code": "switch_1", "value": true }
		//          ],
		//          "custom_name": "Custom Name",
		//          "model": "Model Info",
		//          "ip": "192.168.1.x",
		//          "local_key": "key",
		//          "gateway_id": "gateway_id",
		//          "create_time": 1600000000,
		//          "update_time": 1600000000,
		//          "collections": [
		//             {
		//                "id": "sub_device_id",
		//                "name": "Sub Device",
		//                "category": "infrared_ac",
		//                "product_name": "AC",
		//                "online": true,
		//                "icon": "http://img.url",
		//                "status": [],
		//                "local_key": "key",
		//                "gateway_id": "parent_id",
		//                "create_time": 1600000000,
		//                "update_time": 1600000000
		//             }
		//          ]
		//        }
		//      ],
		//      "total": 1
		//    }
		// }
		api.GET("/devices", getAllDevicesController.GetAllDevices)

		// Get device by ID
		// URL: /api/tuya/devices/:id
		// Method: GET
		// Headers: 
		//    - Authorization: Bearer <token>
		// Param: id (string)
		// Response: {
		//    "status": true,
		//    "message": "Device fetched successfully",
		//    "data": {
		//      "device": {
		//          "id": "device_id",
		//          "name": "Device Name",
		//          "category": "category_code",
		//          "product_name": "Product Name",
		//          "online": true,
		//          "icon": "http://img.url",
		//          "status": [
		//             { "code": "switch_1", "value": true }
		//          ],
		//          "custom_name": "Custom Name",
		//          "model": "Model Info",
		//          "ip": "192.168.1.x",
		//          "local_key": "key",
		//          "gateway_id": "gateway_id",
		//          "create_time": 1600000000,
		//          "update_time": 1600000000,
		//          "collections": []
		//      }
		//    }
		// }
		api.GET("/devices/:id", getDeviceByIDController.GetDeviceByID)

		// Get Sensor Data
		// URL: /api/tuya/devices/:id/sensor
		// Method: GET
		// Headers: 
		//    - Authorization: Bearer <token>
		// Param: id (string)
		// Response: {
		//    "status": true,
		//    "message": "Sensor data fetched successfully",
		//    "data": {
		//        "temperature": 29.4,
		//        "humidity": 82,
		//        "battery_percentage": 100,
		//        "status_text": "Temperature hot, Air moist",
		//        "temp_unit": "°C"
		//    }
		// }
		api.GET("/devices/:id/sensor", sensorController.GetSensorData)
	}
}
