package com.example.teraluxapp.ui

import androidx.compose.runtime.Composable
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.example.teraluxapp.ui.devices.SmartACScreen
import com.example.teraluxapp.ui.devices.SmartACScreen
import com.example.teraluxapp.ui.devices.SwitchDeviceScreen
import com.example.teraluxapp.ui.devices.SensorDeviceScreen
import com.example.teraluxapp.ui.devices.DeviceListScreen
import com.example.teraluxapp.ui.login.LoginScreen
import java.net.URLDecoder
import java.net.URLEncoder

@Composable
fun AppNavigation() {
    val navController = rememberNavController()

    NavHost(navController = navController, startDestination = "login") {
        composable("login") {
            LoginScreen(
                onLoginSuccess = { token, uid ->
                    navController.navigate("devices?token=$token&uid=$uid") {
                        popUpTo("login") { inclusive = true }
                    }
                }
            )
        }
        composable(
            route = "devices?token={token}&uid={uid}",
            arguments = listOf(
                navArgument("token") { type = NavType.StringType },
                navArgument("uid") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            val token = backStackEntry.arguments?.getString("token") ?: ""
            val uid = backStackEntry.arguments?.getString("uid") ?: ""
            DeviceListScreen(
                token = token,
                uid = uid,
                onDeviceClick = { deviceId, category, deviceName, gatewayId ->
                    val encodedName = URLEncoder.encode(deviceName, "UTF-8")
                    val safeGatewayId = gatewayId ?: ""
                    navController.navigate("device/$deviceId/$category/$encodedName?token=$token&gatewayId=$safeGatewayId")
                }
            )
        }
        
        // Device Detail Route with category-based rendering
        composable(
            route = "device/{deviceId}/{category}/{name}?token={token}&gatewayId={gatewayId}",
            arguments = listOf(
                navArgument("deviceId") { type = NavType.StringType },
                navArgument("category") { type = NavType.StringType },
                navArgument("name") { type = NavType.StringType },
                navArgument("token") { type = NavType.StringType },
                navArgument("gatewayId") { type = NavType.StringType; defaultValue = "" }
            )
        ) { backStackEntry ->
            val deviceId = backStackEntry.arguments?.getString("deviceId") ?: ""
            val category = backStackEntry.arguments?.getString("category") ?: ""
            val name = URLDecoder.decode(backStackEntry.arguments?.getString("name") ?: "Device", "UTF-8")
            val gatewayId = backStackEntry.arguments?.getString("gatewayId") ?: ""
            val token = backStackEntry.arguments?.getString("token") ?: ""
            
            // DEBUG: Log the category
            android.util.Log.d("AppNavigation", "Device: $name, Category: $category, ID: $deviceId")
            
            // Route based on category (from actual Tuya API response)
            // - dgnzk = Multi-function controller (Smart Central Control Panel) -> Switch
            // - kg, cz, pc, clkg, cjkg, tdq = Various switches -> Switch
            // - infrared_ac = IR AC remote -> SmartACScreen
            // - wnykq = Universal IR remote (Smart IR) -> SmartACScreen
            // - kt, ktkzq = AC controller -> SmartACScreen
            when {
                // Switch/Multi-function categories -> SwitchDeviceScreen
                category in listOf("dgnzk", "kg", "cz", "pc", "clkg", "cjkg", "tdq", "kgq", "tgkg", "tgq", "dj", "dd") -> {
                    SwitchDeviceScreen(
                        deviceId = deviceId,
                        deviceName = name,
                        token = token,
                        onBack = { navController.popBackStack() }
                    )
                }
                // Sensor categories
                category in listOf("wsdcg") -> {
                    SensorDeviceScreen(
                        deviceId = deviceId,
                        deviceName = name,
                        token = token,
                        onBack = { navController.popBackStack() }
                    )
                }
                // AC/IR categories -> SmartACScreen (IR AC control)
                else -> {
                    SmartACScreen(
                        deviceId = deviceId,
                        deviceName = name,
                        token = token,
                        infraredId = if (gatewayId.isNotEmpty()) gatewayId else "a36d8e212f67a0ea2dbgnl", // Use gatewayId if present, else fallback
                        onBack = { navController.popBackStack() }
                    )
                }
            }
        }
    }
}
