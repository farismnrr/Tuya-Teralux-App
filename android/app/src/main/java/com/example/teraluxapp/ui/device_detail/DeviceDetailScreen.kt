package com.example.teraluxapp.ui.device_detail

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.example.teraluxapp.data.model.Device
import com.example.teraluxapp.data.network.RetrofitClient
import kotlinx.coroutines.launch

@Composable
fun DeviceDetailScreen(deviceId: String, token: String, onBack: () -> Unit) {
    val scope = rememberCoroutineScope()
    var device by remember { mutableStateOf<Device?>(null) }
    var isLoading by remember { mutableStateOf(true) }
    var error by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(deviceId) {
        scope.launch {
            try {
                // If getDeviceById returns SingleDeviceResponse wrapping the device
                val response = RetrofitClient.instance.getDeviceById("Bearer $token", deviceId)
                device = response.data?.device
            } catch (e: Exception) {
                error = "Failed to load device details: ${e.message}"
            } finally {
                isLoading = false
            }
        }
    }

    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        Button(onClick = onBack) {
            Text("Back")
        }
        Spacer(modifier = Modifier.height(16.dp))
        
        if (isLoading) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else if (error != null) {
            Text(text = error!!, color = MaterialTheme.colorScheme.error)
        } else if (device != null) {
            Text(text = device!!.name, style = MaterialTheme.typography.headlineLarge)
            Spacer(modifier = Modifier.height(8.dp))
            Text(text = "ID: ${device!!.id}")
            Text(text = "Category: ${device!!.category}")
            Text(text = "Product: ${device!!.productName}")
            Text(text = "Online: ${device!!.online}")
            Text(text = "IP: ${device!!.ip ?: "N/A"}")
            
            Spacer(modifier = Modifier.height(16.dp))
            Text("Status:", style = MaterialTheme.typography.titleMedium)
            device!!.status?.forEach { status ->
               Text("${status.code}: ${status.value}")
            }
        }
    }
}
