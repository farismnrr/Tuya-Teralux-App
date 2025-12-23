package com.example.teraluxapp.ui.devices

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.PowerSettingsNew
import androidx.compose.material.icons.filled.Remove
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.teraluxapp.data.network.IRACCommandRequest
import com.example.teraluxapp.data.network.RetrofitClient
import kotlinx.coroutines.launch

/**
 * SmartACScreen for controlling AC via Smart IR Hub
 * 
 * IR AC API uses integer values:
 * - power: 0 (off), 1 (on)
 * - mode: 0 (cool), 1 (heat), 2 (auto), 3 (fan), 4 (dry)
 * - temp: 16-30 (celsius)
 * - wind: 0 (auto), 1 (low), 2 (medium), 3 (high)
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SmartACScreen(
    deviceId: String,  // This is the remote_id (AC remote paired to IR hub)
    deviceName: String,
    token: String,
    infraredId: String = "a36d8e212f67a0ea2dbgnl", // Smart IR Hub ID (Corrected: 8 instead of 0)
    onBack: () -> Unit
) {
    val scope = rememberCoroutineScope()
    
    // DEBUG: Log parameters
    LaunchedEffect(Unit) {
        android.util.Log.d("SmartACScreen", "=== PARAMETERS ===")
        android.util.Log.d("SmartACScreen", "deviceId (remote_id): $deviceId")
        android.util.Log.d("SmartACScreen", "infraredId (hub): $infraredId")
        android.util.Log.d("SmartACScreen", "deviceName: $deviceName")
    }
    
    // State with integer values matching Tuya IR API
    var temp by remember { mutableStateOf(24) }
    var modeIndex by remember { mutableStateOf(0) } // 0=cool, 1=heat, 2=auto, 3=fan, 4=dry
    var windIndex by remember { mutableStateOf(0) } // 0=auto, 1=low, 2=medium, 3=high
    var isOn by remember { mutableStateOf(false) }
    var isProcessing by remember { mutableStateOf(false) }
    var isDeviceOnline by remember { mutableStateOf(false) }
    
    val modeLabels = listOf("Cool", "Heat", "Auto", "Fan", "Dry")
    val modeEmojis = listOf("❄️", "🔥", "🔄", "💨", "💧")
    val windLabels = listOf("Auto", "Low", "Medium", "High")
    
    var rawStatus by remember { mutableStateOf("Loading...") }

    // Persistent Storage
    val context = androidx.compose.ui.platform.LocalContext.current
    val prefs = remember { com.example.teraluxapp.utils.DevicePreferences(context) }

    // Fetch initial state
    LaunchedEffect(Unit) {
        try {
            val response = RetrofitClient.instance.getDeviceById("Bearer $token", deviceId)
            val dev = response.data?.device
            if (dev != null) {
                isDeviceOnline = dev.online
                val statuses = dev.status
            
                // DEBUG: Log all statuses
                rawStatus = statuses?.joinToString("\n") { "${it.code}: ${it.value}" } ?: "No Status"
                statuses?.forEach { 
                    android.util.Log.d("SmartACScreen", "Status: ${it.code} = ${it.value}") 
                }
                
                // Check if status is empty (stateless IR)
                if (statuses.isNullOrEmpty()) {
                    val cached = prefs.getACState(deviceId)
                    isOn = cached.isOn
                    temp = cached.temp
                    modeIndex = cached.mode
                    windIndex = cached.speed
                    android.util.Log.d("SmartACScreen", "Loaded cached state: $cached")
                } else {
                    statuses.forEach { status ->
                        val code = status.code.lowercase()
                        when (code) {
                            "switch", "power" -> isOn = status.value.toString().toBoolean()
                            "temp_set", "t", "temp" -> {
                                val t = status.value.toString().toDoubleOrNull()?.toInt()
                                if (t != null) temp = t
                            }
                            "mode" -> {
                                val modeStr = status.value.toString().lowercase()
                                modeIndex = when (modeStr) {
                                    "cool", "cold" -> 0
                                    "heat", "hot" -> 1
                                    "auto" -> 2
                                    "fan", "wind" -> 3
                                    "dry", "wet" -> 4
                                    else -> 0
                                }
                            }
                            "fan_speed_enum", "wind" -> {
                                val fanStr = status.value.toString().lowercase()
                                windIndex = when (fanStr) {
                                    "auto" -> 0
                                    "low" -> 1
                                    "medium" -> 2
                                    "high" -> 3
                                    else -> 0
                                }
                            }
                        }
                    }
                    // Save loaded state to cache
                    prefs.saveACState(deviceId, isOn, temp, modeIndex, windIndex)
                }
            }
        } catch (e: Exception) {
            e.printStackTrace()
            // Fallback to cache on error
            val cached = prefs.getACState(deviceId)
            isOn = cached.isOn
            temp = cached.temp
            modeIndex = cached.mode
            windIndex = cached.speed
        }
    }

    // Send Command (IR endpoint with client-side mapping)
    val sendIRCommand = { code: String, value: Any ->
        // Optimistic Save
        prefs.saveACState(deviceId, isOn, temp, modeIndex, windIndex)

        scope.launch {
            isProcessing = true
            try {
                // For IR devices, we still use the IR endpoint but with proper remote_id
                // deviceId = remote_id (the AC remote paired to the hub)
                // infraredId = the Smart IR Hub ID
                
                // Convert value to Int for IR API
                val intValue = when (value) {
                    is Boolean -> if (value) 1 else 0
                    is Int -> value
                    else -> 0
                }

                android.util.Log.d("SmartACScreen", "Sending IR Command: $code = $intValue to remote $deviceId via hub $infraredId")
                
                val request = IRACCommandRequest(
                    remote_id = deviceId,  // This is the AC remote ID
                    code = code,
                    value = intValue
                )
                val response = RetrofitClient.instance.sendIRACCommand("Bearer $token", infraredId, request)
                
                if (response.isSuccessful && response.body()?.status == true) {
                     android.util.Log.d("SmartACScreen", "Command Success")
                } else {
                     android.util.Log.e("SmartACScreen", "Command Failed: ${response.code()}")
                }

            } catch (e: Exception) {
                e.printStackTrace()
            } finally {
                isProcessing = false
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { 
                    Column {
                        Text(deviceName, fontWeight = FontWeight.Bold)
                        Row(verticalAlignment = Alignment.CenterVertically) {
                             Box(modifier = Modifier.size(6.dp).background(if (isDeviceOnline) Color(0xFF4CAF50) else Color.Red, androidx.compose.foundation.shape.CircleShape))
                             Spacer(modifier = Modifier.width(4.dp))
                             Text(
                                 text = if (isDeviceOnline) "Online" else "Offline",
                                 style = MaterialTheme.typography.labelSmall,
                                 color = if (isDeviceOnline) Color(0xFF4CAF50) else Color.Red
                             )
                        }
                    }
                },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    IconButton(onClick = { /* Edit action */ }) {
                        Icon(Icons.Default.Edit, contentDescription = "Edit")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.White)
            )
        },
        containerColor = Color.White
    ) { paddingValues ->
        Column(
            modifier = Modifier
                .padding(paddingValues)
                .fillMaxSize()
                .padding(horizontal = 24.dp)
                .alpha(if (isDeviceOnline) 1f else 0.5f),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.SpaceBetween
        ) {
            
            // Spacer for top visual balance
            Spacer(modifier = Modifier.height(20.dp))

            // Center: Temp Control
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.Center,
                modifier = Modifier.fillMaxWidth()
            ) {
                // Minus Button
                IconButton(
                    onClick = { 
                        if (temp > 16) {
                            temp--
                            sendIRCommand("temp", temp) 
                        }
                    }
                ) {
                    Icon(Icons.Default.Remove, contentDescription = "Decrease Temp", tint = Color.LightGray, modifier = Modifier.size(32.dp))
                }

                Spacer(modifier = Modifier.width(32.dp))

                // Big Temp Text
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(
                        text = "$temp°C", 
                        fontSize = 72.sp, 
                        fontWeight = FontWeight.Bold,
                        color = if (isOn) Color.Black else Color(0xFFE0E0E0)
                    )
                    Text(
                        text = "Set Temperature", 
                        color = Color.LightGray, 
                        fontSize = 14.sp
                    )
                }

                Spacer(modifier = Modifier.width(32.dp))

                // Plus Button
                IconButton(
                    onClick = { 
                        if (temp < 30) {
                            temp++
                            sendIRCommand("temp", temp) 
                        }
                    }
                ) {
                    Icon(Icons.Default.Add, contentDescription = "Increase Temp", tint = Color.LightGray, modifier = Modifier.size(32.dp))
                }
            }

            // Bottom: Mode/Speed and Switch
            Column(
                modifier = Modifier.fillMaxWidth(),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                // Mode and Speed Row
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceEvenly
                ) {
                    // Mode Button
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Surface(
                            shape = CircleShape,
                            color = Color.White,
                            shadowElevation = 2.dp,
                            modifier = Modifier.size(60.dp),
                            onClick = { 
                                modeIndex = (modeIndex + 1) % 5
                                sendIRCommand("mode", modeIndex)
                            }
                        ) {
                            Box(contentAlignment = Alignment.Center) {
                                Text(modeEmojis[modeIndex], fontSize = 24.sp)
                            }
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        Text("Mode: ${modeLabels[modeIndex]}", color = Color.Gray, fontSize = 12.sp)
                    }

                    // Speed Button
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Surface(
                            shape = CircleShape,
                            color = Color.White,
                            shadowElevation = 2.dp,
                            modifier = Modifier.size(60.dp),
                            onClick = { 
                                windIndex = (windIndex + 1) % 4
                                sendIRCommand("wind", windIndex)
                            }
                        ) {
                            Box(contentAlignment = Alignment.Center) {
                                Text("💨", fontSize = 24.sp)
                            }
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        Text("Speed: ${windLabels[windIndex]}", color = Color.Gray, fontSize = 12.sp)
                    }
                }

                Spacer(modifier = Modifier.height(40.dp))

                // Switch Button (Wide Pill Shape)
                Button(
                    onClick = { 
                        isOn = !isOn
                        sendIRCommand("power", if (isOn) 1 else 0)
                    },
                    modifier = Modifier
                        .fillMaxWidth(0.6f)
                        .height(56.dp),
                    shape = RoundedCornerShape(28.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = if (isOn) Color(0xFF4CAF50) else Color.Gray
                    )
                ) {
                    Icon(
                        Icons.Default.PowerSettingsNew, 
                        contentDescription = null,
                        modifier = Modifier.padding(end = 8.dp)
                    )
                    Text(text = "Switch", fontSize = 18.sp)
                }
                
                Spacer(modifier = Modifier.height(20.dp))
            }
        }
    }
}
