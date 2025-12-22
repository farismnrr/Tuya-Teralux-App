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
    
    // Determine switch configuration
    val switchConfigs = remember(deviceId) {
        when (deviceId) {
            "a378e2ef14e8748cf2cimq" -> listOf(SwitchConfig("switch", "Switch"))
            "a37b710af2df8a6cdcxqlv" -> listOf(SwitchConfig("switch1", "Switch"), SwitchConfig("switch2", "Switch"))
            "a33e768a19b3edc6a98rga" -> listOf(SwitchConfig("switch_1", "Switch 1"), SwitchConfig("switch_2", "Switch 2"), SwitchConfig("switch_3", "Switch 3"))
            else -> listOf(SwitchConfig("switch_1", "Switch 1"), SwitchConfig("switch_2", "Switch 2"))
        }
    }

    val switchStates = remember { mutableStateMapOf<String, Boolean>() }
    val context = androidx.compose.ui.platform.LocalContext.current
    val prefs = remember { com.example.teraluxapp.utils.DevicePreferences(context) }

    // Init & Sync
    LaunchedEffect(deviceId) {
        switchConfigs.forEach { config ->
            switchStates[config.code] = prefs.getGenericSwitchState(deviceId, config.code)
        }
        try {
            val response = RetrofitClient.instance.getDeviceById("Bearer $token", deviceId)
            response.data?.device?.status?.forEach { status ->
               if (switchConfigs.any { it.code == status.code }) {
                   val isOn = status.value.toString().toBoolean()
                   switchStates[status.code] = isOn
                   prefs.saveGenericSwitchState(deviceId, status.code, isOn)
               }
            }
        } catch (e: Exception) { e.printStackTrace() }
    }

    fun sendCommand(code: String, value: Boolean) {
        switchStates[code] = value
        prefs.saveGenericSwitchState(deviceId, code, value)
        scope.launch {
            try {
                RetrofitClient.instance.sendDeviceCommand("Bearer $token", deviceId, Command(code, value))
            } catch (e: Exception) { e.printStackTrace() }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(deviceName, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold) },
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
                    
                    // Responsive sizing
                    val weight = if (switchConfigs.size > 1) 1f else 0f
                    val width = if (switchConfigs.size == 1) 160.dp else Dp.Unspecified

                    ModernSwitchPanel(
                        label = config.label,
                        isOn = isOn,
                        onClick = { sendCommand(config.code, !isOn) },
                        modifier = Modifier
                            .then(if (weight > 0f) Modifier.weight(weight) else Modifier.width(width))
                    )
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
    // Panel/Card implementation matching the image
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
