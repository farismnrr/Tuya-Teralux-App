package com.example.teraluxapp.ui.devices

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Delete
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

import androidx.compose.ui.graphics.vector.rememberVectorPainter
import coil.compose.AsyncImage
import androidx.compose.ui.graphics.ColorMatrix
import androidx.compose.ui.graphics.ColorFilter
import androidx.compose.ui.draw.alpha

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DeviceListScreen(token: String, uid: String, onDeviceClick: (deviceId: String, category: String, deviceName: String, gatewayId: String?) -> Unit) {
    val scope = rememberCoroutineScope()
    var devices by remember { mutableStateOf<List<Device>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }
    var error by remember { mutableStateOf<String?>(null) }
    var page by remember { mutableIntStateOf(1) }
    var totalDevices by remember { mutableIntStateOf(0) }
    val limit = 6
    val snackbarHostState = remember { SnackbarHostState() }
    var isFlushing by remember { mutableStateOf(false) }

    val fetchDevices = { pageNum: Int ->
        scope.launch {
            isLoading = true
            error = null
            try {
                // Pass page and limit=6
                val response = RetrofitClient.instance.getDevices("Bearer $token", page = pageNum, limit = limit)
                if (response.isSuccessful && response.body() != null) {
                    val body = response.body()!!.data
                    val rawDevices = body?.devices ?: emptyList()
                    totalDevices = body?.totalDevices ?: 0
                    
                    val flatList = ArrayList<Device>()
                    for (d in rawDevices) {
                        // If device has collections (e.g., IR Hub with AC remotes), 
                        // skip the hub itself and only add the remotes
                        if (d.collections.isNullOrEmpty()) {
                            flatList.add(d)
                        } else {
                            // Only add the collections (AC remotes), not the hub
                            flatList.addAll(d.collections)
                        }
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

    LaunchedEffect(page) {
        fetchDevices(page)
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("My Devices") },
                actions = {
                    // Cache flush button
                    IconButton(
                        onClick = {
                            scope.launch {
                                isFlushing = true
                                try {
                                    val response = RetrofitClient.instance.flushCache("Bearer $token")
                                    if (response.isSuccessful) {
                                        snackbarHostState.showSnackbar("Cache cleared successfully")
                                        fetchDevices(page) // Refresh device list
                                    } else {
                                        snackbarHostState.showSnackbar("Failed to clear cache: ${response.code()}")
                                    }
                                } catch (e: Exception) {
                                    snackbarHostState.showSnackbar("Error: ${e.message}")
                                } finally {
                                    isFlushing = false
                                }
                            }
                        },
                        enabled = !isFlushing
                    ) {
                        Icon(Icons.Default.Delete, contentDescription = "Clear Cache")
                    }
                    // Refresh button
                    IconButton(onClick = { fetchDevices(page) }) {
                        Icon(Icons.Default.Refresh, contentDescription = "Refresh")
                    }
                }
            )
        },
        snackbarHost = { SnackbarHost(snackbarHostState) }
    ) { paddingValues ->
        Column(modifier = Modifier.padding(paddingValues).fillMaxSize()) {
            if (isLoading) {
                Box(Modifier.weight(1f).fillMaxWidth(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator()
                }
            } else if (error != null) {
                Box(Modifier.weight(1f).fillMaxWidth(), contentAlignment = Alignment.Center) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(text = error!!, color = MaterialTheme.colorScheme.error)
                        Spacer(modifier = Modifier.height(8.dp))
                        Button(onClick = { fetchDevices(page) }) {
                            Text("Retry")
                        }
                    }
                }
            } else {
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxWidth()
                        .padding(8.dp), // Reduced padding
                    verticalArrangement = Arrangement.spacedBy(8.dp) // Reduced spacing
                ) {
                    val firstRowDevices = devices.take(3)
                    val secondRowDevices = if (devices.size > 3) devices.drop(3).take(3) else emptyList()

                    // Row 1
                    Row(
                        modifier = Modifier
                            .weight(1f)
                            .fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(8.dp) // Reduced spacing
                    ) {
                        for (i in 0 until 3) {
                            if (i < firstRowDevices.size) {
                                val device = firstRowDevices[i]
                                // For IR devices, use remote_id as deviceId and id as gatewayId
                                val actualDeviceId = device.remoteId ?: device.id
                                val actualGatewayId = if (device.remoteId != null) device.id else device.gatewayId
                                val actualCategory = device.remoteCategory ?: device.category
                                DeviceItem(
                                    device = device,
                                    modifier = Modifier.weight(1f),
                                    onClick = {
                                        if (!device.online) {
                                            scope.launch {
                                                snackbarHostState.showSnackbar("Device is offline and cannot be controlled")
                                            }
                                        } else {
                                            onDeviceClick(actualDeviceId, actualCategory, device.name, actualGatewayId)
                                        }
                                    }
                                )
                            } else {
                                Spacer(modifier = Modifier.weight(1f))
                            }
                        }
                    }

                    // Row 2
                    Row(
                        modifier = Modifier
                            .weight(1f)
                            .fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(8.dp) // Reduced spacing
                    ) {
                        for (i in 0 until 3) {
                            if (i < secondRowDevices.size) {
                                val device = secondRowDevices[i]
                                // For IR devices, use remote_id as deviceId and id as gatewayId
                                val actualDeviceId = device.remoteId ?: device.id
                                val actualGatewayId = if (device.remoteId != null) device.id else device.gatewayId
                                val actualCategory = device.remoteCategory ?: device.category
                                DeviceItem(
                                    device = device,
                                    modifier = Modifier.weight(1f),
                                    onClick = {
                                        if (!device.online) {
                                            scope.launch {
                                                snackbarHostState.showSnackbar("Device is offline and cannot be controlled")
                                            }
                                        } else {
                                            onDeviceClick(actualDeviceId, actualCategory, device.name, actualGatewayId)
                                        }
                                    }
                                )
                            } else {
                                Spacer(modifier = Modifier.weight(1f))
                            }
                        }
                    }
                }
            }

            // Pagination Controls
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Button(
                    onClick = { if (page > 1) page-- },
                    enabled = page > 1
                ) {
                    Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Previous")
                    Spacer(Modifier.width(8.dp))
                    Text("Prev")
                }

                Text("Page $page of ${if (totalDevices > 0) kotlin.math.ceil(totalDevices.toDouble() / limit).toInt() else 1}")

                val maxPage = if (totalDevices > 0) kotlin.math.ceil(totalDevices.toDouble() / limit).toInt() else 1
                Button(
                    onClick = { if (page < maxPage) page++ },
                    enabled = page < maxPage
                ) {
                    Text("Next")
                    Spacer(Modifier.width(8.dp))
                    Icon(Icons.AutoMirrored.Filled.ArrowForward, contentDescription = "Next")
                }
            }
        }
    }
}


@Composable
fun DeviceItem(device: Device, modifier: Modifier = Modifier, onClick: () -> Unit) {
    val saturationMatrix = remember { ColorMatrix() }
    LaunchedEffect(device.online) {
        saturationMatrix.setToSaturation(if (device.online) 1f else 0f)
    }

    Card(
        modifier = modifier
            .fillMaxWidth()
            .padding(4.dp)
            .clickable(onClick = onClick),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        )
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            Column(
                modifier = Modifier
                    .padding(8.dp)
                    .fillMaxSize()
                    .alpha(if (device.online) 1f else 0.6f), // Dim content if offline
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center
            ) {
                Spacer(modifier = Modifier.height(16.dp)) // Space for badge

                // Icon
                val iconUrl = "https://images.tuyacn.com/${device.icon}"
                AsyncImage(
                    model = iconUrl,
                    contentDescription = null,
                    modifier = Modifier.size(64.dp),
                    placeholder = rememberVectorPainter(Icons.Default.Home),
                    error = rememberVectorPainter(Icons.Default.Home),
                    colorFilter = ColorFilter.colorMatrix(saturationMatrix)
                )

                Spacer(modifier = Modifier.height(12.dp))

                // Name
                Text(
                    text = device.name,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                    textAlign = androidx.compose.ui.text.style.TextAlign.Center,
                    color = if (device.online) Color.Unspecified else Color.Gray
                )
            }

            // Status Badge
            Surface(
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .padding(4.dp),
                shape = RoundedCornerShape(4.dp),
                color = if (device.online) Color(0xFFE8F5E9) else Color(0xFFF5F5F5)
            ) {
                Text(
                    text = if (device.online) "Online" else "Offline",
                    style = MaterialTheme.typography.labelSmall,
                    color = if (device.online) Color(0xFF2E7D32) else Color.Gray,
                    modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                )
            }
        }
    }
}
