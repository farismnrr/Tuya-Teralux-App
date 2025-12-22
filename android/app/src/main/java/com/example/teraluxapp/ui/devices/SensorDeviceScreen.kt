package com.example.teraluxapp.ui.devices

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.rounded.DeviceThermostat
import androidx.compose.material.icons.rounded.Face
import androidx.compose.material.icons.rounded.WaterDrop
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.teraluxapp.data.network.RetrofitClient
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SensorDeviceScreen(
    deviceId: String,
    deviceName: String,
    token: String,
    onBack: () -> Unit
) {
    var temperature by remember { mutableDoubleStateOf(0.0) }
    var humidity by remember { mutableIntStateOf(0) }
    var statusText by remember { mutableStateOf("Loading...") }
    var tempUnit by remember { mutableStateOf("°C") }
    var isLoading by remember { mutableStateOf(true) }

    LaunchedEffect(Unit) {
        while (true) {
            try {
                val response = RetrofitClient.instance.getSensorData("Bearer $token", deviceId)
                if (response.isSuccessful && response.body()?.status == true) {
                    val data = response.body()?.data
                    if (data != null) {
                        temperature = data.temperature
                        humidity = data.humidity
                        statusText = data.status_text
                        tempUnit = data.temp_unit
                    } else {
                        if (isLoading) statusText = "No data available"
                    }
                } else {
                    if (isLoading) statusText = "Failed to load data"
                }
            } catch (e: Exception) {
                 if (isLoading) statusText = "Error: ${e.message}"
                e.printStackTrace()
            } finally {
                isLoading = false
            }
            kotlinx.coroutines.delay(5000) // Refresh every 5 seconds
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(deviceName, fontWeight = FontWeight.Bold, color = Color.White) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back", tint = Color.White)
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = Color.Transparent,
                    titleContentColor = Color.White,
                    navigationIconContentColor = Color.White
                )
            )
        },
        containerColor = Color.Transparent
    ) { paddingValues ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    brush = Brush.verticalGradient(
                        colors = listOf(
                            Color(0xFF29B6F6), // Light Blue 400
                            Color(0xFF039BE5), // Light Blue 600
                            Color(0xFF0288D1)  // Light Blue 700
                        )
                    )
                )
                .padding(paddingValues),
            contentAlignment = Alignment.TopCenter
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(16.dp)
                    .widthIn(max = 600.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.SpaceEvenly // Distribute space evenly
            ) {
                // Main Stats Row
                Row(
                    modifier = Modifier.fillMaxWidth().padding(top = 16.dp),
                    horizontalArrangement = Arrangement.Center,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    // Temperature
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text("Temperature", color = Color.White.copy(alpha = 0.9f), fontSize = 16.sp, fontWeight = FontWeight.Medium)
                        Spacer(modifier = Modifier.height(4.dp))
                        Row(verticalAlignment = Alignment.Top) {
                            Text(
                                text = String.format("%.1f", temperature),
                                fontSize = 56.sp, // Reduced from 72
                                fontWeight = FontWeight.Bold,
                                color = Color.White
                            )
                            Text(
                                text = tempUnit,
                                fontSize = 24.sp, // Reduced from 28
                                fontWeight = FontWeight.SemiBold,
                                color = Color.White.copy(alpha = 0.9f),
                                modifier = Modifier.padding(top = 12.dp)
                            )
                        }
                    }

                    Spacer(modifier = Modifier.width(32.dp))
                    
                    // Vertical Divider
                    Box(
                        modifier = Modifier
                            .width(1.5.dp)
                            .height(60.dp) // Reduced from 80
                            .background(Color.White.copy(alpha = 0.4f))
                            .clip(CircleShape)
                    )

                    Spacer(modifier = Modifier.width(32.dp))

                    // Humidity
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text("Humidity", color = Color.White.copy(alpha = 0.9f), fontSize = 16.sp, fontWeight = FontWeight.Medium)
                        Spacer(modifier = Modifier.height(4.dp))
                        Row(verticalAlignment = Alignment.Top) {
                            Text(
                                text = "$humidity",
                                fontSize = 56.sp, // Reduced from 72
                                fontWeight = FontWeight.Bold,
                                color = Color.White
                            )
                            Text(
                                text = "%",
                                fontSize = 24.sp, // Reduced from 28
                                fontWeight = FontWeight.SemiBold,
                                color = Color.White.copy(alpha = 0.9f),
                                modifier = Modifier.padding(top = 12.dp)
                            )
                        }
                    }
                }

                // Status Card
                Card(
                    modifier = Modifier.fillMaxWidth().padding(vertical = 8.dp),
                    shape = RoundedCornerShape(20.dp),
                    colors = CardDefaults.cardColors(containerColor = Color.White),
                    elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
                ) {
                    Row(
                        modifier = Modifier
                            .padding(16.dp)
                            .fillMaxWidth(),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        // Icon
                        Box(
                            modifier = Modifier
                                .size(40.dp) // Reduced from 48
                                .clip(CircleShape)
                                .background(Color(0xFFE1F5FE)),
                            contentAlignment = Alignment.Center
                        ) {
                           Icon(
                               imageVector = Icons.Rounded.Face,
                               contentDescription = null,
                               tint = Color(0xFF0288D1),
                               modifier = Modifier.size(24.dp)
                           )
                        }
                        
                        Spacer(modifier = Modifier.width(16.dp))
                        
                        Text(
                            text = statusText,
                            fontSize = 16.sp, // Reduced from 18
                            color = Color(0xFF455A64),
                            fontWeight = FontWeight.Medium
                        )
                    }
                }

                // Sliders/Gauges Card
                Card(
                    modifier = Modifier.fillMaxWidth().weight(1f, fill = false), // Allow it to take available space but not force
                    shape = RoundedCornerShape(24.dp),
                    colors = CardDefaults.cardColors(containerColor = Color.White),
                    elevation = CardDefaults.cardElevation(defaultElevation = 4.dp)
                ) {
                    Column(
                        modifier = Modifier.padding(24.dp),
                        verticalArrangement = Arrangement.Center
                    ) {
                        /* Temperature Gauge */
                        Row(verticalAlignment = Alignment.CenterVertically) {
                             Icon(
                                 imageVector = Icons.Rounded.DeviceThermostat,
                                 contentDescription = null,
                                 tint = Color(0xFF0288D1),
                                 modifier = Modifier.size(20.dp)
                             )
                             Spacer(modifier = Modifier.width(8.dp))
                             Text("Temperature $tempUnit", color = Color(0xFF546E7A), fontSize = 14.sp, fontWeight = FontWeight.Medium)
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        
                        // Custom Gauge Track
                        Box(contentAlignment = Alignment.CenterStart) {
                             // Track
                             Box(
                                 modifier = Modifier
                                     .fillMaxWidth()
                                     .height(8.dp) // Reduced height
                                     .clip(RoundedCornerShape(4.dp))
                                     .background(Color(0xFFECEFF1))
                             )
                             // Progress
                             Box(
                                 modifier = Modifier
                                     .fillMaxWidth((temperature.toFloat() / 50f).coerceIn(0f, 1f))
                                     .height(8.dp)
                                     .clip(RoundedCornerShape(4.dp))
                                     .background(
                                         Brush.horizontalGradient(
                                             colors = listOf(Color(0xFF4FC3F7), Color(0xFF0288D1))
                                         )
                                     )
                             )
                        }
                        
                        Box(modifier = Modifier.fillMaxWidth().padding(top = 2.dp)) {
                             Text(
                                 text = String.format("%.1f%s", temperature, tempUnit),
                                 modifier = Modifier.align(Alignment.CenterEnd),
                                 fontSize = 12.sp,
                                 color = Color(0xFF78909C),
                                 fontWeight = FontWeight.Medium
                             )
                        }

                        Spacer(modifier = Modifier.height(24.dp)) // Reduced spacing

                        /* Humidity Gauge */
                        Row(verticalAlignment = Alignment.CenterVertically) {
                             Icon(
                                 imageVector = Icons.Rounded.WaterDrop,
                                 contentDescription = null,
                                 tint = Color(0xFF26A69A),
                                 modifier = Modifier.size(20.dp)
                             )
                             Spacer(modifier = Modifier.width(8.dp))
                             Text("Humidity %", color = Color(0xFF546E7A), fontSize = 14.sp, fontWeight = FontWeight.Medium)
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        
                         // Custom Gauge Track Humidity
                        Box(contentAlignment = Alignment.CenterStart) {
                             // Track
                             Box(
                                 modifier = Modifier
                                     .fillMaxWidth()
                                     .height(8.dp)
                                     .clip(RoundedCornerShape(4.dp))
                                     .background(Color(0xFFECEFF1))
                             )
                             // Progress
                             Box(
                                 modifier = Modifier
                                     .fillMaxWidth((humidity.toFloat() / 100f).coerceIn(0f, 1f))
                                     .height(8.dp)
                                     .clip(RoundedCornerShape(4.dp))
                                     .background(
                                         Brush.horizontalGradient(
                                             colors = listOf(Color(0xFF80CBC4), Color(0xFF00897B))
                                         )
                                     )
                             )
                        }

                         Box(modifier = Modifier.fillMaxWidth().padding(top = 2.dp)) {
                             Text(
                                 text = "$humidity%",
                                 modifier = Modifier.align(Alignment.CenterEnd),
                                 fontSize = 12.sp,
                                 color = Color(0xFF78909C),
                                 fontWeight = FontWeight.Medium
                             )
                        }
                        
                        Spacer(modifier = Modifier.height(20.dp))
                        Text(
                            "Updated just now", 
                            fontSize = 11.sp, 
                            color = Color(0xFFB0BEC5), 
                            modifier = Modifier.align(Alignment.CenterHorizontally)
                        )
                    }
                }
            }
        }
    }
}
