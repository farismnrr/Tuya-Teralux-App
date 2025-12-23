package com.example.teraluxapp.ui.devices

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.teraluxapp.data.network.Command
import com.example.teraluxapp.data.network.RetrofitClient
import kotlinx.coroutines.launch

data class SwitchConfig(val code: String, val label: String)

/**
 * SwitchDeviceScreen - Visual Replica of Reference Image
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SwitchDeviceScreen(
    deviceId: String,
    deviceName: String,
    token: String,
    onBack: () -> Unit
) {
    val scope = rememberCoroutineScope()
    val snackbarHostState = remember { SnackbarHostState() }
    
    // Dynamic switch configuration from device status
    var switchConfigs by remember { mutableStateOf<List<SwitchConfig>>(emptyList()) }
    val switchStates = remember { mutableStateMapOf<String, Boolean>() }
    var isOnline by remember { mutableStateOf(false) }
    var isLoading by remember { mutableStateOf(true) }
    var errorMessage by remember { mutableStateOf<String?>(null) }


    // Helper function to check if status is a valid switch control
    fun isValidSwitchControl(status: com.example.teraluxapp.data.model.DeviceStatus): Boolean {
        val code = status.code.lowercase()
        return code.contains("switch") && 
               !code.contains("countdown") &&
               !code.contains("relay") &&
               !code.contains("light") &&
               status.value is Boolean
    }

    // Helper function to generate switch label
    fun generateSwitchLabel(code: String, index: Int, totalSwitches: Int): String {
        if (totalSwitches == 1) return "Switch"
        
        val num = code.filter { it.isDigit() }.toIntOrNull() ?: (index + 1)
        return "Switch $num"
    }

    // Helper function to initialize switch states from device status
    fun initializeSwitchStates(switches: List<com.example.teraluxapp.data.model.DeviceStatus>) {
        switches.forEach { status ->
            val isOn = status.value.toString().toBoolean()
            switchStates[status.code] = isOn
        }
    }
    
    // Load state from backend
    suspend fun loadStateFromBackend() {
        try {
            val response = RetrofitClient.instance.getDeviceState("Bearer $token", deviceId)
            if (response.isSuccessful && response.body()?.status == true) {
                val state = response.body()?.data
                state?.last_commands?.forEach { cmd ->
                    // Backend can send as Boolean or Number (0/1)
                    val switchValue = when (val v = cmd.value) {
                        is Boolean -> v
                        is Number -> v.toInt() == 1
                        else -> false
                    }
                    switchStates[cmd.code] = switchValue
                }
                android.util.Log.d("SwitchDevice", "Loaded state from backend: ${state?.last_commands?.size} commands")
            }
        } catch (e: Exception) {
            android.util.Log.e("SwitchDevice", "Failed to load state: ${e.message}")
        }
    }

    // Init & Sync
    LaunchedEffect(deviceId) {
        isLoading = true
        errorMessage = null
        
        try {
            val response = RetrofitClient.instance.getDeviceById("Bearer $token", deviceId)
            val device = response.data?.device
            
            // Early return if device is null
            if (device == null) {
                errorMessage = "Failed to load device information"
                return@LaunchedEffect
            }
            
            isOnline = device.online
            
            // Extract and filter switch controls
            val switches = device.status?.filter { isValidSwitchControl(it) } ?: emptyList()
            
            // Early return if no switches found
            if (switches.isEmpty()) {
                errorMessage = "No switch controls found for this device"
                return@LaunchedEffect
            }
            
            // Generate switch configurations
            switchConfigs = switches.mapIndexed { index, status ->
                SwitchConfig(
                    code = status.code,
                    label = generateSwitchLabel(status.code, index, switches.size)
                )
            }
            
            // Initialize switch states from device status
            initializeSwitchStates(switches)
            
            // Load saved state from backend (overrides device status)
            loadStateFromBackend()
            
        } catch (e: Exception) {
            e.printStackTrace()
            errorMessage = "Error: ${e.message}"
        } finally {
            isLoading = false
        }
    }

    fun sendCommand(code: String, value: Boolean) {
        if (!isOnline) {
            scope.launch {
                snackbarHostState.showSnackbar("Device is offline")
            }
            return
        }
        
        // Optimistic update
        switchStates[code] = value
        
        scope.launch {
            try {
                // Send command to device
                val cmdResponse = RetrofitClient.instance.sendDeviceCommand("Bearer $token", deviceId, Command(code, value))
                
                if (cmdResponse.isSuccessful && cmdResponse.body()?.status == true) {
                    // Save ALL switch states to backend (not just the one that changed)
                    val allSwitchCommands = switchStates.map { (switchCode, switchValue) ->
                        com.example.teraluxapp.data.network.StateCommand(switchCode, switchValue)
                    }
                    val stateRequest = com.example.teraluxapp.data.network.SaveDeviceStateRequest(
                        commands = allSwitchCommands
                    )
                    RetrofitClient.instance.saveDeviceState("Bearer $token", deviceId, stateRequest)
                } else {
                    // Revert on failure
                    switchStates[code] = !value
                    snackbarHostState.showSnackbar("Command failed")
                }
            } catch (e: Exception) {
                e.printStackTrace()
                // Revert on error
                switchStates[code] = !value
                snackbarHostState.showSnackbar("Error: ${e.message}")
            }
        }
    }

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
        topBar = {
            TopAppBar(
                title = {
                    Column {
                        Text(deviceName, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Box(modifier = Modifier.size(6.dp).background(if (isOnline) Color(0xFF4CAF50) else Color.Red, androidx.compose.foundation.shape.CircleShape))
                            Spacer(modifier = Modifier.width(4.dp))
                            Text(
                                text = if (isOnline) "Online" else "Offline",
                                style = MaterialTheme.typography.labelSmall,
                                color = if (isOnline) Color(0xFF4CAF50) else Color.Red
                            )
                        }
                    }
                },
                navigationIcon = {
                    IconButton(onClick = onBack) { Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back") }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.White)
            )
        }
    ) { paddingValues ->
        // Main Background Gradient - Light Blue
        Box(
            modifier = Modifier
                .padding(paddingValues)
                .fillMaxSize()
                .alpha(if (isOnline) 1f else 0.5f)
                .background(
                    brush = Brush.verticalGradient(
                        colors = listOf(
                            Color(0xFFE3F2FD), // Top: Light Blue
                            Color(0xFFF0F7FF), // Mid
                            Color(0xFFFFFFFF)  // Bottom: White
                        )
                    )
                ),
            contentAlignment = Alignment.Center
        ) {
            when {
                isLoading -> {
                    CircularProgressIndicator()
                }
                errorMessage != null -> {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        modifier = Modifier.padding(24.dp)
                    ) {
                        Text(
                            text = errorMessage!!,
                            color = MaterialTheme.colorScheme.error,
                            style = MaterialTheme.typography.bodyLarge,
                            textAlign = androidx.compose.ui.text.style.TextAlign.Center
                        )
                    }
                }
                switchConfigs.isEmpty() -> {
                    Text(
                        text = "No switches available",
                        color = Color.Gray,
                        style = MaterialTheme.typography.bodyLarge
                    )
                }
                else -> {
                    // Container for Switches
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 24.dp),
                        horizontalArrangement = Arrangement.spacedBy(16.dp, Alignment.CenterHorizontally),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        switchConfigs.forEach { config ->
                            val isOn = switchStates[config.code] == true
                            
                            ModernSwitchPanel(
                                label = config.label,
                                isOn = isOn,
                                onClick = { sendCommand(config.code, !isOn) },
                                modifier = Modifier
                                    .width(120.dp) // Fixed small width as requested ("kecil aja")
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun ModernSwitchPanel(
    label: String,
    isOn: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    // Reverted to Original Default Style (Card/Gradient) as requested
    Box(
        modifier = modifier
            .height(280.dp) // Tall vertical ratio
            .shadow(
                elevation = 6.dp, 
                shape = RoundedCornerShape(20.dp),
                spotColor = Color.Black.copy(alpha = 0.15f)
            )
            .clip(RoundedCornerShape(20.dp))
            .background(
                brush = Brush.linearGradient(
                    colors = if (isOn) listOf(
                        Color(0xFFF0F4FF), // Slightly brighter/whiter when ON
                        Color(0xFFFFFFFF)
                    ) else listOf(
                        Color(0xFFE8EAF6), // Slightly grayish/blue when OFF
                        Color(0xFFF5F5F5)
                    ),
                    start = androidx.compose.ui.geometry.Offset(0f, 0f),
                    end = androidx.compose.ui.geometry.Offset(0f, Float.POSITIVE_INFINITY)
                )
            )
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null
            ) { onClick() }
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(bottom = 32.dp),
            verticalArrangement = Arrangement.Bottom,
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            // Label
            Text(
                text = label,
                fontSize = 16.sp,
                color = Color.Gray,
                fontWeight = FontWeight.Normal
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            
            // Indicator Pill
            Box(
                modifier = Modifier
                    .width(40.dp)
                    .height(4.dp)
                    .clip(RoundedCornerShape(2.dp))
                    .background(
                        if (isOn) Color(0xFF4CAF50) // Green indicator when ON
                        else Color.LightGray // Grey indicator when OFF
                    )
            )
        }
    }
}
