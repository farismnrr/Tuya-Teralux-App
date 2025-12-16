package com.example.teraluxapp.ui.devices

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.teraluxapp.data.network.Command
import com.example.teraluxapp.data.network.CommandRequest
import com.example.teraluxapp.data.network.RetrofitClient
import kotlinx.coroutines.launch

/**
 * SwitchDeviceScreen for controlling 2-gang switch devices
 * Matches custom UI design with gradient background and dual controls
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
    
    // Independent states for 2 switches
    var switch1On by remember { mutableStateOf(false) }
    var switch2On by remember { mutableStateOf(false) }
    
    var isProcessing by remember { mutableStateOf(false) }

    // Persistent Storage
    val context = androidx.compose.ui.platform.LocalContext.current
    val prefs = remember { com.example.teraluxapp.utils.DevicePreferences(context) }

    // Fetch initial state
    LaunchedEffect(Unit) {
        // Load cache first (Optimistic)
        val cached = prefs.getSwitchState(deviceId)
        switch1On = cached.switch1
        switch2On = cached.switch2

        try {
            val response = RetrofitClient.instance.getDeviceById(token, deviceId)
            val statuses = response.device.status
            // Update from API
            statuses?.forEach { status ->
                when (status.code) {
                    "switch_1" -> switch1On = status.value.toString().toBoolean()
                    "switch_2" -> switch2On = status.value.toString().toBoolean()
                }
            }
            // Update cache with fresh data
            prefs.saveSwitchState(deviceId, switch1On, switch2On)
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    // Helper to send command
    fun sendCommand(code: String, value: Boolean) {
        // Optimistic update
        if (code == "switch_1") switch1On = value
        if (code == "switch_2") switch2On = value
        prefs.saveSwitchState(deviceId, switch1On, switch2On)

        scope.launch {
            isProcessing = true
            try {
                val cmd = Command(code, value)
                val request = CommandRequest(listOf(cmd))
                val response = RetrofitClient.instance.sendDeviceCommand(token, deviceId, request)
                if (response.isSuccessful) {
                    // Update cache again (just in case)
                    prefs.saveSwitchState(deviceId, switch1On, switch2On)
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
                title = { Text(deviceName, fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    IconButton(onClick = { /* Edit action */ }) {
                        Icon(Icons.Default.Edit, contentDescription = "Edit")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.White)
            )
        }
    ) { paddingValues ->
        Box(
            modifier = Modifier
                .padding(paddingValues)
                .fillMaxSize()
                .background(
                    brush = Brush.verticalGradient(
                        colors = listOf(
                            Color(0xFFE3F2FD), // Light Blue 50
                            Color(0xFFF5F9FF), // Very light blue
                            Color(0xFFFFFFFF)  // White
                        )
                    )
                ),
            contentAlignment = Alignment.Center
        ) {
            Row(
                horizontalArrangement = Arrangement.spacedBy(20.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Switch 1
                SwitchControl(
                    label = "Switch",
                    isOn = switch1On,
                    onClick = { sendCommand("switch_1", !switch1On) }
                )

                // Switch 2
                SwitchControl(
                    label = "Switch",
                    isOn = switch2On,
                    onClick = { sendCommand("switch_2", !switch2On) }
                )
            }
        }
    }
}

@Composable
fun SwitchControl(
    label: String,
    isOn: Boolean,
    onClick: () -> Unit
) {
    // Custom switch UI based on the image provided
    Box(
        modifier = Modifier
            .width(130.dp)
            .height(260.dp)
            .shadow(
                elevation = if (isOn) 4.dp else 2.dp,
                shape = RoundedCornerShape(24.dp),
                spotColor = Color.Black.copy(alpha = 0.1f)
            )
            .clip(RoundedCornerShape(24.dp))
            .background(
                brush = Brush.linearGradient(
                    colors = if (isOn) listOf(
                        Color(0xFFEAF4FF), // Slightly brighter blueish white when on
                        Color(0xFFF8FBFF)
                    ) else listOf(
                        Color(0xFFF0F4F8), // Grayer when off
                        Color(0xFFFFFFFF)
                    )
                )
            )
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null // No ripple for clean custom look, or add if desired
            ) { onClick() }
            .border(
                width = 1.dp,
                color = Color.White.copy(alpha = 0.5f),
                shape = RoundedCornerShape(24.dp)
            )
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(bottom = 40.dp),
            verticalArrangement = Arrangement.Bottom,
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(
                text = label,
                fontSize = 18.sp,
                color = Color.Gray,
                fontWeight = FontWeight.Medium
            )
            
            Spacer(modifier = Modifier.height(12.dp))
            
            // Indicator bar
            Box(
                modifier = Modifier
                    .width(60.dp)
                    .height(6.dp)
                    .clip(RoundedCornerShape(3.dp))
                    .background(
                        if (isOn) Color(0xFF4CAF50) // Green when on 
                        else Color.LightGray.copy(alpha = 0.5f) // Gray when off
                    )
            )
        }
        
        // Optional: faint overlay to show pressed state could be added here
    }
}
