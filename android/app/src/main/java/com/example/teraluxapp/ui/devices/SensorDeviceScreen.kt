package com.example.teraluxapp.ui.devices

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
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
                    val data = response.body()!!.data
                    temperature = data!!.temperature
                    humidity = data.humidity
                    statusText = data.status_text
                    tempUnit = data.temp_unit
                } else {
                    // Only update error text if we haven't loaded data yet, otherwise keep showing stale data 
                    // or show a snackbar (omitted for simplicity, keeping existing logic but safe)
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
            contentAlignment = Alignment.TopCenter // Align content to top center
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState())
                    .padding(16.dp)
                    .widthIn(max = 600.dp), // Limit width for landscape/tablets
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Spacer(modifier = Modifier.height(20.dp))

                // Main Stats Row
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.Center,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    // Temperature
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text("Temperature", color = Color.White.copy(alpha = 0.9f), fontSize = 16.sp, fontWeight = FontWeight.Medium)
                        Spacer(modifier = Modifier.height(8.dp))
                        Row(verticalAlignment = Alignment.Top) {
                            Text(
                                text = String.format("%.1f", temperature),
                                fontSize = 72.sp,
                                fontWeight = FontWeight.Bold,
                                color = Color.White
                            )
                            Text(
                                text = tempUnit,
                                fontSize = 28.sp,
                                fontWeight = FontWeight.SemiBold,
                                color = Color.White.copy(alpha = 0.9f),
                                modifier = Modifier.padding(top = 16.dp)
                            )
                        }
                    }

                    Spacer(modifier = Modifier.width(48.dp))
                    
                    // Vertical Divider
                    Box(
                        modifier = Modifier
                            .width(1.5.dp)
                            .height(80.dp)
                            .background(Color.White.copy(alpha = 0.4f))
                            .clip(CircleShape)
                    )

                    Spacer(modifier = Modifier.width(48.dp))

                    // Humidity
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text("Humidity", color = Color.White.copy(alpha = 0.9f), fontSize = 16.sp, fontWeight = FontWeight.Medium)
                        Spacer(modifier = Modifier.height(8.dp))
                        Row(verticalAlignment = Alignment.Top) {
                            Text(
                                text = "$humidity",
                                fontSize = 72.sp,
                                fontWeight = FontWeight.Bold,
                                color = Color.White
                            )
                            Text(
                                text = "%",
                                fontSize = 28.sp,
                                fontWeight = FontWeight.SemiBold,
                                color = Color.White.copy(alpha = 0.9f),
                                modifier = Modifier.padding(top = 16.dp)
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.height(50.dp))

                // Status Card
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(24.dp),
                    colors = CardDefaults.cardColors(containerColor = Color.White),
                    elevation = CardDefaults.cardElevation(defaultElevation = 4.dp)
                ) {
                    Row(
                        modifier = Modifier
                            .padding(20.dp)
                            .fillMaxWidth(),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        // Icon
                        Box(
                            modifier = Modifier
                                .size(48.dp)
                                .clip(CircleShape)
                                .background(Color(0xFFE1F5FE)), // Very light blue
                            contentAlignment = Alignment.Center
                        ) {
                           Icon(
                               imageVector = Icons.Rounded.Face,
                               contentDescription = null,
                               tint = Color(0xFF0288D1),
                               modifier = Modifier.size(28.dp)
                           )
                        }
                        
                        Spacer(modifier = Modifier.width(20.dp))
                        
                        Text(
                            text = statusText,
                            fontSize = 18.sp,
                            color = Color(0xFF455A64),
                            fontWeight = FontWeight.Medium
                        )
                    }
                }

                Spacer(modifier = Modifier.height(24.dp))

                // Sliders/Gauges Card
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(24.dp),
                    colors = CardDefaults.cardColors(containerColor = Color.White),
                    elevation = CardDefaults.cardElevation(defaultElevation = 4.dp)
                ) {
                    Column(modifier = Modifier.padding(28.dp)) {
                        /* Temperature Gauge */
                        Row(verticalAlignment = Alignment.CenterVertically) {
                             Icon(
                                 imageVector = Icons.Rounded.DeviceThermostat,
                                 contentDescription = null,
                                 tint = Color(0xFF0288D1), // Match primary
                                 modifier = Modifier.size(20.dp)
                             )
                             Spacer(modifier = Modifier.width(8.dp))
                             Text("Temperature $tempUnit", color = Color(0xFF546E7A), fontSize = 16.sp, fontWeight = FontWeight.Medium)
                        }
                        Spacer(modifier = Modifier.height(12.dp))
                        
                        // Custom Gauge Track
                        Box(contentAlignment = Alignment.CenterStart) {
                             // Track
                             Box(
                                 modifier = Modifier
                                     .fillMaxWidth()
                                     .height(10.dp)
                                     .clip(RoundedCornerShape(5.dp))
                                     .background(Color(0xFFECEFF1))
                             )
                             // Progress
                             Box(
                                 modifier = Modifier
                                     .fillMaxWidth((temperature.toFloat() / 50f).coerceIn(0f, 1f))
                                     .height(10.dp)
                                     .clip(RoundedCornerShape(5.dp))
                                     .background(
                                         Brush.horizontalGradient(
                                             colors = listOf(Color(0xFF4FC3F7), Color(0xFF0288D1))
                                         )
                                     )
                             )
                        }
                        
                        Box(modifier = Modifier.fillMaxWidth().padding(top = 4.dp)) {
                             Text(
                                 text = String.format("%.1f%s", temperature, tempUnit),
                                 modifier = Modifier.align(Alignment.CenterEnd),
                                 fontSize = 14.sp,
                                 color = Color(0xFF78909C),
                                 fontWeight = FontWeight.Medium
                             )
                        }

                        Spacer(modifier = Modifier.height(32.dp))

                        /* Humidity Gauge */
                        Row(verticalAlignment = Alignment.CenterVertically) {
                             Icon(
                                 imageVector = Icons.Rounded.WaterDrop,
                                 contentDescription = null,
                                 tint = Color(0xFF26A69A), // Teal for humidity
                                 modifier = Modifier.size(20.dp)
                             )
                             Spacer(modifier = Modifier.width(8.dp))
                             Text("Humidity %", color = Color(0xFF546E7A), fontSize = 16.sp, fontWeight = FontWeight.Medium)
                        }
                        Spacer(modifier = Modifier.height(12.dp))
                        
                         // Custom Gauge Track Humidity
                        Box(contentAlignment = Alignment.CenterStart) {
                             // Track
                             Box(
                                 modifier = Modifier
                                     .fillMaxWidth()
                                     .height(10.dp)
                                     .clip(RoundedCornerShape(5.dp))
                                     .background(Color(0xFFECEFF1))
                             )
                             // Progress
                             Box(
                                 modifier = Modifier
                                     .fillMaxWidth((humidity.toFloat() / 100f).coerceIn(0f, 1f))
                                     .height(10.dp)
                                     .clip(RoundedCornerShape(5.dp))
                                     .background(
                                         Brush.horizontalGradient(
                                             colors = listOf(Color(0xFF80CBC4), Color(0xFF00897B))
                                         )
                                     )
                             )
                        }

                         Box(modifier = Modifier.fillMaxWidth().padding(top = 4.dp)) {
                             Text(
                                 text = "$humidity%",
                                 modifier = Modifier.align(Alignment.CenterEnd),
                                 fontSize = 14.sp,
                                 color = Color(0xFF78909C),
                                 fontWeight = FontWeight.Medium
                             )
                        }
                        
                        Spacer(modifier = Modifier.height(30.dp)) // Extra space at bottom
                        Text(
                            "Updated just now", 
                            fontSize = 12.sp, 
                            color = Color(0xFFB0BEC5), 
                            modifier = Modifier.align(Alignment.CenterHorizontally)
                        )
                    }
                }
            }
        }
    }
}
