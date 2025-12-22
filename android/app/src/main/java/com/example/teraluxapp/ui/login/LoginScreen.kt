package com.example.teraluxapp.ui.login

import androidx.compose.foundation.layout.*
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.example.teraluxapp.data.network.RetrofitClient
import kotlinx.coroutines.launch

@Composable
fun LoginScreen(onLoginSuccess: (String, String) -> Unit) {
    val scope = rememberCoroutineScope()
    var isLoading by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }

    Column(
        modifier = Modifier.fillMaxSize().padding(16.dp),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(text = "Tuya Device Manager", style = MaterialTheme.typography.headlineMedium)
        
        Spacer(modifier = Modifier.height(32.dp))

        if (isLoading) {
            CircularProgressIndicator()
        } else {
            Button(
                onClick = {
                    isLoading = true
                    error = null
                    scope.launch {
                        try {
                            val response = RetrofitClient.instance.authenticate()
                            response.data?.let {
                                onLoginSuccess(it.accessToken, it.uid) 
                            } ?: run {
                                error = "Login failed: No data received"
                            }
                        } catch (e: Exception) {
                            error = "Login failed: ${e.message}"
                            e.printStackTrace()
                        } finally {
                            isLoading = false
                        }
                    }
                },
                modifier = Modifier.fillMaxWidth()
            ) {
                Text("Login to Tuya")
            }
        }

        if (error != null) {
            Spacer(modifier = Modifier.height(16.dp))
            Text(text = error!!, color = MaterialTheme.colorScheme.error)
        }
    }
}
