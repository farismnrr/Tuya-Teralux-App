package com.example.teraluxapp.ui.devices

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.example.teraluxapp.data.model.Device
import com.example.teraluxapp.data.network.RetrofitClient
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DeviceListScreen(token: String, uid: String, onDeviceClick: (deviceId: String, category: String, deviceName: String, gatewayId: String?) -> Unit) {
    val scope = rememberCoroutineScope()
    var devices by remember { mutableStateOf<List<Device>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }
    var error by remember { mutableStateOf<String?>(null) }

    val fetchDevices = {
        scope.launch {
            isLoading = true
            error = null
            try {
                val response = RetrofitClient.instance.getDevices("Bearer $token")
                if (response.isSuccessful && response.body() != null) {
                    val rawDevices = response.body()?.data?.devices ?: emptyList()
                    val flatList = ArrayList<Device>()
                    for (d in rawDevices) {
                        flatList.add(d)
                        d.collections?.let { flatList.addAll(it) }
                    }
                    devices = flatList
                } else {
                    val errorBody = response.errorBody()?.string()
                    error = "Failed: ${response.code()}"
                }
            } catch (e: Exception) {
                error = "Error: ${e.message}"
                e.printStackTrace()
            } finally {
                isLoading = false
            }
        }
    }

    LaunchedEffect(Unit) {
        fetchDevices()
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("My Devices") },
                actions = {
                    IconButton(onClick = { fetchDevices() }) {
                        Icon(Icons.Default.Refresh, contentDescription = "Refresh")
                    }
                }
            )
        }
    ) { paddingValues ->
        Box(modifier = Modifier.padding(paddingValues).fillMaxSize()) {
            if (isLoading) {
                Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator()
                }
            } else if (error != null) {
                Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(text = error!!, color = MaterialTheme.colorScheme.error)
                        Spacer(modifier = Modifier.height(8.dp))
                        Button(onClick = { fetchDevices() }) {
                            Text("Retry")
                        }
                    }
                }
            } else {
                LazyVerticalGrid(
                    columns = GridCells.Adaptive(minSize = 150.dp),
                    contentPadding = PaddingValues(16.dp),
                    horizontalArrangement = Arrangement.spacedBy(16.dp),
                    verticalArrangement = Arrangement.spacedBy(16.dp),
                    modifier = Modifier.fillMaxSize()
                ) {
                    items(devices) { device ->
                        DeviceItem(device = device, onClick = { onDeviceClick(device.id, device.category, device.name, device.gatewayId) })
                    }
                }
            }
        }
    }
}

@Composable
fun DeviceItem(device: Device, onClick: () -> Unit) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .aspectRatio(1f) // Square shape
            .clickable(onClick = onClick),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp),
        colors = CardDefaults.cardColors(
            containerColor = if (device.online) MaterialTheme.colorScheme.surface else Color.LightGray.copy(alpha = 0.3f)
        )
    ) {
        Column(
            modifier = Modifier
                .padding(12.dp)
                .fillMaxSize(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.SpaceBetween // Distribute space
        ) {
            // Top: Status Dot
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.End
            ) {
                Box(
                    modifier = Modifier
                        .size(8.dp)
                        .background(
                            color = if (device.online) Color.Green else Color.Gray,
                            shape = RoundedCornerShape(50)
                        )
                )
            }

            // Center: Icon
            Icon(
                imageVector = Icons.Default.Home, // Placeholder icon
                contentDescription = null,
                modifier = Modifier.size(48.dp),
                tint = if (device.online) MaterialTheme.colorScheme.primary else Color.Gray
            )

            // Bottom: Name
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Text(
                    text = device.name,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                Text(
                    text = if (device.online) "Online" else "Offline",
                    style = MaterialTheme.typography.labelMedium,
                    color = if (device.online) MaterialTheme.colorScheme.onSurfaceVariant else Color.Gray
                )
            }
        }
    }
}
